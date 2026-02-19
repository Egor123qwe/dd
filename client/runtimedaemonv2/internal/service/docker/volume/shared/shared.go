package shared

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/volume"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/util/command"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/util/sync/fnController"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/pkg/logger"
)

const (
	SyncInterval    = 1000 * time.Millisecond
	SyncTimeout     = 15 * time.Second
	MaxSyncErrCount = 3

	DefaultSharedVolumeMount = "/storage"

	ConfigName           = "roy9"
	RcloneConfigTemplate = `[%s]
type = s3
provider = Minio
access_key_id = %s
secret_access_key = %s
endpoint = %s
`
)

var log = logger.NewLogger("shared", logger.DefaultWithSentry())

type Service interface {
	Connect(ctx context.Context, req volume.SharedVolume) error
	Disconnect(ctx context.Context) error

	State(ctx context.Context) (volume.SharedVolumeState, error)
}

type service struct {
	config config

	command        command.Service
	connController fnController.Controller

	configPath      string
	localVolumePath string

	connected *atomic.Bool
	connMutex *sync.Mutex
	mutex     *sync.Mutex
}

func New(localVolumePath string) (Service, error) {
	configPath, err := initConfigRcloneDir()
	if err != nil {
		return nil, fmt.Errorf("failed to init rclone config dir: %w", err)
	}

	srv := service{
		config: newConfig(),

		command:        command.New(),
		connController: fnController.New(),

		configPath:      configPath,
		localVolumePath: localVolumePath,

		connected: &atomic.Bool{},
		connMutex: &sync.Mutex{},
		mutex:     &sync.Mutex{},
	}

	return srv, nil
}

func (s service) Connect(ctx context.Context, req volume.SharedVolume) error {
	log.Info("[SHARED FOLDER] shared folder connection started")
	defer log.Infof("[SHARED FOLDER] shared folder connection ended")

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// disconnect
	s.unsafeDisconnect()

	// init config
	cfg := fmt.Sprintf(
		RcloneConfigTemplate,
		ConfigName,
		req.AccessKeyID,
		req.SecretAccessKey,
		s.config.storageURL,
	)

	if err := os.WriteFile(s.configPath, []byte(cfg), 0644); err != nil {
		return fmt.Errorf("failed to write rclone config: %w", err)
	}

	// init volume
	if err := s.refreshFolder(s.localVolumePath); err != nil {
		return fmt.Errorf("failed to refresh shared volume laocal folder: %w", err)
	}

	copyParams := []string{
		s.config.rclonePath,
		"copy",
		ConfigName + ":/" + req.BucketName,
		s.localVolumePath,
	}

	if _, err := s.command.Run(ctx, copyParams); err != nil {
		return fmt.Errorf("failed to get user data: %w", err)
	}

	// start sync serve
	syncParams := []string{
		s.config.rclonePath,
		"sync",
		s.localVolumePath,
		ConfigName + ":/" + req.BucketName,
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.connController.SetCancelFn(cancel)

	s.connected.Store(true)

	go func() {
		s.connMutex.Lock()
		defer s.connMutex.Unlock()
		defer s.connected.Store(false)

		var errCounter int

		for {
			syncCtx, cancel := context.WithTimeout(ctx, SyncTimeout)

			if _, err := s.command.Run(syncCtx, syncParams); err != nil {
				errCounter++
				log.Errorf("failed to sunc volume: %v", err)

				if errCounter >= MaxSyncErrCount {
					log.Errorf("failed to sunc volume for %d times: %v", errCounter, err)
					cancel()

					return
				}

			} else {
				errCounter = 0
			}

			cancel()

			log.Info("volume synced")

			select {
			case <-time.After(SyncInterval):

			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (s service) Disconnect(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.unsafeDisconnect()

	return nil
}

func (s service) unsafeDisconnect() {
	s.connController.Cancel()

	log.Info("[SHARED FOLDER] piko proxy called stopped")
	s.connMutex.Lock()
	s.connMutex.Unlock()

	log.Info("[SHARED FOLDER] piko proxy stopped")
}

func (s service) State(ctx context.Context) (volume.SharedVolumeState, error) {
	state := volume.SharedVolumeState{
		Enabled: s.connected.Load(),
	}

	return state, nil
}

func (s service) refreshFolder(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("error deleting folder: %w", err)
	}

	if err := os.Mkdir(dir, 0755); err != nil {
		return fmt.Errorf("error creating folder: %w", err)
	}

	return nil
}
