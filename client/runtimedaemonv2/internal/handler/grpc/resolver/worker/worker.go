package worker

import (
	"context"

	model "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/handler/grpc/generate"
	srv "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/pkg/logger"
)

var log = logger.NewLogger("worker handler", logger.DefaultWithSentry())

type Handler struct {
	model.UnimplementedWorkerServer
	srv srv.Service
}

func New(srv srv.Service) Handler {
	return Handler{
		srv: srv,
	}
}

// GetInfo - debug request to check service availabilities
func (h Handler) GetInfo(ctx context.Context, req *model.InfoReq) (*model.Info, error) {
	info, err := h.srv.Info(ctx)
	if err != nil {
		return nil, err
	}

	resp := &model.Info{
		MachineId: info.UniqueMachineId,

		Docker: &model.DockerInfo{
			Available: info.Docker.Availability,
			Version:   info.Docker.Version,
		},

		Network: &model.NetworkInfo{
			Tailscale: &model.TailscaleInfo{
				Available: info.Network.Tailscale.Availability,
				Version:   info.Network.Tailscale.Version,
			},

			Piko: &model.PikoInfo{
				Available: info.Network.Piko.Availability,
				Version:   info.Network.Piko.Version,
			},
		},

		Version: info.Version,
	}

	return resp, nil
}
