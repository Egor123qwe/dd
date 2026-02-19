package api

import (
	"context"

	proto "gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/proto/runtimedaemon/generate"
	"google.golang.org/grpc"
)

type RuntimeDaemon interface {
	proto.WorkerClient
	proto.ContainerClient
	proto.TemplateClient

	GetHardware(ctx context.Context) (*proto.SystemInfo, error)
}

type client struct {
	proto.WorkerClient
	proto.ContainerClient
	proto.TemplateClient
}

func New(conn *grpc.ClientConn) RuntimeDaemon {
	client := client{
		WorkerClient:    proto.NewWorkerClient(conn),
		ContainerClient: proto.NewContainerClient(conn),
		TemplateClient:  proto.NewTemplateClient(conn),
	}

	return client
}
