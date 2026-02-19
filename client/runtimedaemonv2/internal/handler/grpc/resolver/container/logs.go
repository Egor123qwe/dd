package container

import (
	"context"
	"fmt"
	"sync"
	"time"

	model "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/handler/grpc/generate"
	containerModel "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/container"
)

const (
	streamTimeout    = 1 * time.Second
	streamLogTimeout = 100 * time.Millisecond
)

func (h Handler) GetLogsStream(req *model.ContainerLogsStreamReq, sender model.Container_GetLogsStreamServer) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if req.RowsInResp <= 0 {
		return fmt.Errorf("invalid rows in response: %d", req.RowsInResp)
	}

	// Create logs reader
	opts := containerModel.LogReq{
		TemplateID: req.TemplateID,
		Tail:       req.Tail,
	}

	logCh, err := h.srv.Container().Logs(ctx, opts)
	if err != nil {
		log.Errorf("failed to get logs: %s", err)

		return err
	}

	// Throw away useless logs (less than offset value)
	for range req.Offset {
		select {
		case <-logCh:
		case <-ctx.Done():
			return nil
		}
	}

	logs := make([]string, 0)
	logsMutex := &sync.Mutex{}

	// Receive logs thread
	go func() {
		for log := range logCh {
			log += "\n\r"

			logsMutex.Lock()
			logs = append(logs, log)
			logsMutex.Unlock()
		}
	}()

	// Send logs func
	sendFunc := func() error {
		count := len(logs)

		if count == 0 {
			return nil
		}

		if count > int(req.RowsInResp) {
			count = int(req.RowsInResp)
		}

		// Send logs
		if err := sender.Send(&model.ContainerLogs{Rows: logs[:count]}); err != nil {
			return err
		}

		// throw away used logs
		logs = logs[count:]

		return nil
	}

	sendTicker := time.NewTicker(streamTimeout)
	checkAndSendTicker := time.NewTicker(streamLogTimeout)

	// Send logs loop
	// Logs will send if logs [count = requested rows in response] or if [streamTimeout seconds have passed]
	for {
		logsMutex.Lock()

		select {
		case <-sender.Context().Done():
			return nil

		case <-sendTicker.C:
			if err := sendFunc(); err != nil {
				return err
			}

		case <-checkAndSendTicker.C:
			if len(logs) >= int(req.RowsInResp) {
				if err := sendFunc(); err != nil {
					return err
				}
			}
		}

		logsMutex.Unlock()
	}
}
