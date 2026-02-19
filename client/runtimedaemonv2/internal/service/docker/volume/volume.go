package volume

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/volume"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/api"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/volume/shared"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/pkg/logger"
)

const (
	BaseVolumesDir = "roy9_volumes"
	LocalUsageDir  = "local"
	HostUsageDir   = "host"
	SharedUsageDir = "shared"
)

var log = logger.NewLogger("volume", logger.DefaultWithSentry())

// Service is abstract layer on container api to use container in container
type Service interface {
	Shared() shared.Service

	Volume(usage volume.Usage, templateID, userID, name string) api.Volume
	Path() Paths
}

type Paths struct {
	SharedVolumeDir string
	LocalVolumeDir  string
	HostVolumeDir   string
}
type service struct {
	path   Paths
	shared shared.Service
}

func New() (Service, error) {
	userDir, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to getting home directory %w", err)
	}

	volumeDir := filepath.Join(userDir.HomeDir, BaseVolumesDir)

	err = os.Mkdir(volumeDir, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return nil, fmt.Errorf("falid to init storage: %w", err)
	}

	if err := initVolumesDirs(volumeDir); err != nil {
		return nil, fmt.Errorf("falid to init volumes dirs: %w", err)
	}

	path := Paths{
		LocalVolumeDir:  filepath.Join(volumeDir, LocalUsageDir),
		HostVolumeDir:   filepath.Join(volumeDir, HostUsageDir),
		SharedVolumeDir: filepath.Join(volumeDir, SharedUsageDir),
	}

	shared, err := shared.New(path.SharedVolumeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to init shared volume service: %w", err)
	}

	srv := &service{
		path:   path,
		shared: shared,
	}

	return srv, nil
}

func (s service) Volume(usage volume.Usage, templateID, userID, name string) api.Volume {
	var basePath string

	switch usage {
	case volume.Local:
		basePath = s.path.LocalVolumeDir

	case volume.Host:
		basePath = s.path.HostVolumeDir

	case volume.Shared:
		return api.Volume{
			HostPath:  s.path.SharedVolumeDir,
			LocalPath: name,
		}
	}

	result := api.Volume{
		HostPath:  basePath + "/" + templateID + "_" + userID + "_" + strings.Replace(name, "/", "-", -1),
		LocalPath: name,
	}

	return result
}

func (s service) Path() Paths {
	return s.path
}

func initVolumesDirs(volumeDir string) error {
	err := os.Mkdir(filepath.Join(volumeDir, LocalUsageDir), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("falid to init storage: %w", err)
	}

	err = os.Mkdir(filepath.Join(volumeDir, HostUsageDir), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("falid to init storage: %w", err)
	}

	err = os.Mkdir(filepath.Join(volumeDir, SharedUsageDir), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("falid to init storage: %w", err)
	}

	return nil
}

func (s service) Shared() shared.Service {
	return s.shared
}
