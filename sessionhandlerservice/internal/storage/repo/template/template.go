package template

import (
	"context"

	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model/rent"
)

type Template interface {
	Get(ctx context.Context, id string) (rent.Template, error)
	ListAll(ctx context.Context) ([]rent.Template, error)
	Create(ctx context.Context, t rent.Template) error
	Update(ctx context.Context, id string, title, type_, description, shortDescription, version, imageName, imageTag string, useGPU bool, ports []rent.Port, envs []rent.Env, volumes []string, minCPU int32, minRAMBytes, minStorageBytes uint64, minVolumeStorageBytes []uint64) error
}
