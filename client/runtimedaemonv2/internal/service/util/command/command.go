package command

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/util"
)

// used as default
const (
	linuxCommandProcessing = "bash"
	linuxCommandArg        = "-c"
)

const (
	WindowsCommandProcessing = "cmd"
	WindowsCommandArg        = "/C"
)

var (
	commandProcessing = linuxCommandProcessing
	commandArg        = linuxCommandArg
)

type Service interface {
	Run(ctx context.Context, params []string) ([]byte, error)
	RunCommand(ctx context.Context, command string) ([]byte, error)
}

type service struct{}

func New() Service {
	initCmd()

	return &service{}
}

func (s *service) Run(ctx context.Context, params []string) ([]byte, error) {
	command := strings.Join(params, " ")

	cmd := exec.CommandContext(ctx, commandProcessing, commandArg, command)

	cmd.Env = append(os.Environ())

	output, err := cmd.Output()
	if err != nil {
		return output, err
	}

	return output, nil
}

func (s *service) RunCommand(ctx context.Context, command string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, commandProcessing, commandArg, command)

	output, err := cmd.Output()
	if err != nil {
		return output, err
	}

	return output, nil
}

func initCmd() {
	if util.IsWindows() {
		commandProcessing = WindowsCommandProcessing
		commandArg = WindowsCommandArg
	}
}
