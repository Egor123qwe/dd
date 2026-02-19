package api

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	gpuParam = "all"
)

type Volume struct {
	LocalPath string
	HostPath  string
}

type CreateContainerReq struct {
	Image string
	Name  string

	UseGPU  bool
	Volumes []Volume
	Envs    []string
}

func (s service) CreateContainer(ctx context.Context, req CreateContainerReq) (string, error) {
	resources := container.Resources{}

	if req.UseGPU {
		// https://stackoverflow.com/questions/73742554/how-to-pass-gpus-all-option-to-docker-with-go-sdk
		gpuOpts := opts.GpuOpts{}
		gpuOpts.Set(gpuParam)

		resources = container.Resources{DeviceRequests: gpuOpts.Value()}
	}

	var Mounts []mount.Mount

	for _, v := range req.Volumes {
		err := os.Mkdir(v.HostPath, os.ModePerm)
		if err != nil && !os.IsExist(err) {
			return "", fmt.Errorf("falid to init volume: %w", err)
		}

		Mounts = append(Mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: v.HostPath,
			Target: v.LocalPath,
		})
	}

	result, err := s.dockerApi.ContainerCreate(
		ctx,
		&container.Config{
			Image: req.Image,
			Env:   req.Envs,
		},
		&container.HostConfig{
			PublishAllPorts: true,
			Resources:       resources,
			Mounts:          Mounts,
		},
		&network.NetworkingConfig{},
		&v1.Platform{},
		req.Name,
	)

	if err != nil {
		return "", err
	}

	return result.ID, nil
}
