package api

import (
	"context"
	"errors"
	"fmt"

	proto "gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/proto/runtimedaemon/generate"
)

var ErrRequestCancel = errors.New("request cancelled")

// GetHardware get hardware info from state stream
// this function will wait for all params to be ready
func (c client) GetHardware(ctx context.Context) (*proto.SystemInfo, error) {
	stream, err := c.GetStateStream(ctx, &proto.RuntimeStateStreamReq{})
	if err != nil {
		return &proto.SystemInfo{}, fmt.Errorf("falid to connect to daemon state stream: %v", err)
	}

	defer stream.CloseSend()

	resultCh := make(chan *proto.SystemInfo)
	errCh := make(chan error)

	go func() {
		for {
			if err := stream.Context().Err(); err != nil {
				errCh <- ErrRequestCancel
				break
			}

			data, err := stream.Recv()
			if err != nil {
				errCh <- err
				break
			}

			hardware := data.System

			if isHardwareReady(hardware) {
				resultCh <- hardware
				break
			}
		}
	}()

	select {
	case <-ctx.Done():
		return nil, ErrRequestCancel

	case err := <-errCh:
		return nil, err

	case result := <-resultCh:
		return result, nil
	}
}

func isHardwareReady(info *proto.SystemInfo) bool {
	network := info.Network.Upload != 0 && info.Network.Download != 0 && info.Network.Ping != 0

	return network
}
