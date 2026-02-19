package container

import (
	"bufio"
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	containerModel "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/container"
)

const (
	allLogs = "all" // const for getting all logs from container API
)

const (
	streamTypeLen = 8 // correct logs start at index 8, DockerAPI send STREAM_TYPE at indexes 0-7
)

func (s service) Logs(ctx context.Context, req containerModel.LogReq) (<-chan string, error) {
	tail := allLogs

	if req.Tail != nil {
		tail = fmt.Sprintf("%d", *req.Tail)
	}

	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Details:    true,
		Tail:       tail,
		Follow:     true,
	}

	var templateID string

	if req.TemplateID == nil {
		settings, err := s.storage.Settings().Get()
		if err != nil {
			return nil, fmt.Errorf("failed to get settings: %w", err)
		}

		if settings.Options == nil {
			return nil, fmt.Errorf("container settings not found")
		}

		templateID = settings.Options.Container.TemplateID
	} else {
		templateID = *req.TemplateID
	}

	logCh, err := s.getContainerLogsScanner(ctx, options, Name(templateID))
	if err != nil {
		return nil, fmt.Errorf("failed to read container logs: %w", err)
	}

	return logCh, nil
}

func (s service) getContainerLogsScanner(ctx context.Context, options container.LogsOptions, name string) (<-chan string, error) {
	logsReader, err := s.api.GetContainerLogsReader(ctx, options, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs reader: %w", err)
	}

	logCh := make(chan string)

	tty, err := s.api.GetContainerTTY(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get container tty: %w", err)
	}

	offset := streamTypeLen
	if tty {
		offset = 0
	}

	go func() {
		defer close(logCh)

		scanner := bufio.NewScanner(logsReader)
		defer logsReader.Close()

		for {
			select {
			case <-ctx.Done():
				return

			default:
				if !scanner.Scan() {
					if err := scanner.Err(); err != nil {
						log.Errorf("failed to read container logs: %s", err)
					}

					return
				}

				log := scanner.Bytes()
				if len(log) >= offset {
					log = log[offset:]
				}

				logCh <- string(log)
			}
		}
	}()

	return logCh, nil
}
