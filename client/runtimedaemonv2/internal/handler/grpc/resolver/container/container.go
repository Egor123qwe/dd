package container

import (
	model "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/handler/grpc/generate"
	srv "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/pkg/logger"
)

var log = logger.NewLogger("container", logger.DefaultWithSentry())

type Handler struct {
	model.UnimplementedContainerServer
	srv srv.Service
}

func New(srv srv.Service) Handler {
	return Handler{
		srv: srv,
	}
}
