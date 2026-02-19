package worker

import (
	"context"
	"errors"

	model "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/handler/grpc/generate"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/auth"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network/piko"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network/tailscale"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/runtime"
)

var ErrAuthCredentialsRequired = errors.New("auth credentials are required")

func (h Handler) ChangeMode(ctx context.Context, req *model.ChangeModeReq) (*model.ChangeModeResp, error) {
	log.Infof("[HANDLER] change mode started")
	defer log.Infof("[HANDLER] change mode finished")

	serviceReq := runtime.Configuration{
		Settings: runtime.Settings{
			Mode: runtime.Mode(req.Mode.Number()),
		},
	}

	if req.Docker != nil {
		containerOptions := docker.ContainerSettings{
			TemplateID: req.Docker.Container.TemplateID,
		}

		if req.Docker.Container.Options.SharedVolume != nil {
			containerOptions.SharedVolume = &docker.SharedVolume{
				AccessKeyID:     req.Docker.Container.Options.SharedVolume.Credentials.AccessKeyID,
				SecretAccessKey: req.Docker.Container.Options.SharedVolume.Credentials.SecretAccessKey,

				BucketName: req.Docker.Container.Options.SharedVolume.BucketName,
				Mount:      req.Docker.Container.Options.SharedVolume.VolumeMount,
			}
		}

		authOptions := auth.Settings{}

		if req.Docker.Auth.Enabled {
			if req.Docker.Auth.Credentials == nil {
				return nil, ErrAuthCredentialsRequired
			}

			authOptions = auth.Settings{
				Credentials: &auth.Credentials{
					Login:    req.Docker.Auth.Credentials.Login,
					Password: req.Docker.Auth.Credentials.Password,
				},
			}
		}

		serviceReq.Components.Docker = &runtime.DockerOptions{
			ForUserID: req.Docker.ClientUserId,

			Container: containerOptions,
			Auth:      authOptions,
		}
	}

	if req.Network != nil {
		serviceReq.Components.Network = &network.Options{}

		if req.Network.Tailscale != nil {
			serviceReq.Components.Network.Tailscale = &tailscale.Settings{
				ClientID: req.Network.Tailscale.ClientId,
				AuthKey:  req.Network.Tailscale.AuthKey,
			}
		}

		if req.Network.Piko != nil {
			serviceReq.Components.Network.Piko = &piko.Settings{
				AuthKey: req.Network.Piko.AuthKey,
			}

			for _, e := range req.Network.Piko.Endpoints {
				serviceEndpoint := piko.EndpointSettings{
					PortID: e.TemplatePort,
					Name:   e.Name,
				}

				serviceReq.Components.Network.Piko.Endpoints = append(
					serviceReq.Components.Network.Piko.Endpoints, serviceEndpoint,
				)
			}
		}
	}

	err := h.srv.ChangeMode(ctx, serviceReq)

	if err != nil {
		log.Infof("[HANDLER] change mode failed: %s", err)
	}

	return &model.ChangeModeResp{}, err
}
