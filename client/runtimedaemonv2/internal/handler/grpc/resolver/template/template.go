package template

import (
	model "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/handler/grpc/generate"
	srv "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/template"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/pkg/logger"
)

var log = logger.NewLogger("template", logger.DefaultWithSentry())

type Handler struct {
	model.UnimplementedTemplateServer
	srv srv.Service
}

func New(srv srv.Service) Handler {
	return Handler{
		srv: srv,
	}
}
