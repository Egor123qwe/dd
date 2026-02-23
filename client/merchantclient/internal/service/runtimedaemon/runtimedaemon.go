package runtimedaemon

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	config2 "gitlab.roy9.ru/roy9/backend/clientside/merchantclient/config"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/service/runtimedaemon/api"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/command"
	proto "gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/proto/runtimedaemon/generate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// runtimedaemon binary name per OS (no path)
const daemonBinaryName = "runtimedaemon"

func getRuntimeDaemonPath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("os.Executable: %w", err)
	}
	dir := filepath.Dir(execPath)
	name := daemonBinaryName
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(dir, "runtimedaemon", name), nil
}

var (
	ErrDaemonShutDown = errors.New("runtime daemon shut down")
	ErrNotEnabled     = errors.New("runtime daemon not enabled")
)

type API interface {
	api.RuntimeDaemon
	io.Closer

	WaitEnabled(ctx context.Context) error
	Serve(ctx context.Context) error
}

type client struct {
	api.RuntimeDaemon
	conn *grpc.ClientConn

	cmd command.Service
}

func New() (API, error) {
	config := newConfig()

	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", "localhost", config.port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return nil, err
	}

	client := client{
		RuntimeDaemon: api.New(conn),
		conn:          conn,

		cmd: command.New(),
	}

	return client, nil
}

func (c client) Serve(ctx context.Context) error {
	daemonPath, err := getRuntimeDaemonPath()
	if err != nil {
		return fmt.Errorf("runtime daemon path: %w", err)
	}
	args := []string{daemonPath}
	if host := strings.TrimSpace(os.Getenv(config2.EnvBackendHost)); host != "" {
		args = append(args, "--backend-host="+host)
	}
	_, err = c.cmd.Run(ctx, args)
	if err != nil && ctx.Err() == nil {
		return fmt.Errorf("runtime daemon failed: %v", err)
	}
	return ErrDaemonShutDown
}

func (c client) WaitEnabled(ctx context.Context) error {
	for {
		_, err := c.GetInfo(ctx, &proto.InfoReq{})
		if err != nil && strings.HasSuffix(err.Error(), "context deadline exceeded") {
			return ErrNotEnabled
		}

		if err == nil {
			break
		}
	}

	if _, err := c.ChangeMode(ctx, &proto.ChangeModeReq{Mode: proto.Mode_Disable}); err != nil {
		return fmt.Errorf("failed to configure runtime daemon: %w", err)
	}

	return nil
}

func (c client) Close() error {
	return c.conn.Close()
}
