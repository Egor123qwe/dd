package grpc

import (
	runtime "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/handler/grpc/generate"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/handler/grpc/resolver/container"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/handler/grpc/resolver/template"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/handler/grpc/resolver/worker"
	srv "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service"
	"google.golang.org/grpc"
)

type Handler interface {
	Subscribe(server *grpc.Server)
}

type handler struct {
	worker worker.Handler

	container container.Handler
	template  template.Handler
}

func New(srv srv.Service) Handler {
	return &handler{
		container: container.New(srv.Docker()),
		worker:    worker.New(srv),
		template:  template.New(srv.Docker().Template()),
	}
}

func (h *handler) Subscribe(server *grpc.Server) {
	runtime.RegisterWorkerServer(server, h.worker)
	runtime.RegisterContainerServer(server, h.container)
	runtime.RegisterTemplateServer(server, h.template)
}
