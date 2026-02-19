package handler

import (
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/handler/grpc"
	srv "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service"
)

type Handler struct {
	Grpc grpc.Handler
}

func New(srv srv.Service) Handler {
	return Handler{
		Grpc: grpc.New(srv),
	}
}
