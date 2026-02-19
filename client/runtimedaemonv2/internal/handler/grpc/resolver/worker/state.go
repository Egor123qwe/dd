package worker

import (
	"context"
	"fmt"
	"time"

	model "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/handler/grpc/generate"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/hardware"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/runtime"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/runtime/health"
)

const (
	streamTimeout       = 5 * time.Second
	stateReceiveTimeout = 20 * time.Second
)

func (h Handler) GetStateStream(req *model.RuntimeStateStreamReq, sender model.Worker_GetStateStreamServer) error {
	for {
		stateCtx, _ := context.WithTimeout(context.Background(), stateReceiveTimeout)

		state, err := h.srv.State(stateCtx)
		if err != nil {
			log.Errorf("failed to get state: %s", err)

			state.Health = health.Health{
				Status:    health.Error,
				StatusMsg: fmt.Errorf("failed to get state: %w", err).Error(),
			}
		}

		// Send health
		if err := sender.Send(convertStateToHandlerType(state)); err != nil {
			return err
		}

		<-time.After(streamTimeout)
	}
}

func convertStateToHandlerType(state health.State) *model.RuntimeState {
	resp := &model.RuntimeState{
		Mode: model.Mode(state.Worker.Mode),

		Status:    model.RuntimeState_Status(state.Health.Status),
		StatusMsg: state.Health.StatusMsg,
	}

	// Docker state
	resp.Container = &model.DockerState{
		Status:    model.DockerState_Status(state.Components.Docker.Health.Status),
		StatusMsg: state.Components.Docker.Health.StatusMsg,
	}

	resp.Container.Container = &model.ContainerState{
		TemplateID: state.Components.Docker.Container.UsedTemplateID,
		Status:     model.ContainerState_Status(state.Components.Docker.Container.State.Status),
		StatusMsg:  state.Components.Docker.Container.State.StatusMsg,

		SharedVolume: &model.ContainerState_SharedVolume{
			Enabled: state.Components.Docker.Container.SharedVolume.Enabled,
		},
	}

	for _, p := range state.Components.Docker.Container.State.Ports {
		port := &model.ContainerState_Port{
			LocalPort: p.Local,
			HostPort:  p.Host,
		}

		resp.Container.Container.Ports = append(resp.Container.Container.Ports, port)
	}

	resp.Container.Auth = &model.ContainerAuthState{
		Enabled: state.Components.Docker.Auth.Enabled,
	}

	for _, p := range state.Components.Docker.Auth.Ports {
		port := &model.ContainerAuthState_Port{
			InPort:  p.InPort,
			OutPort: p.OutPort,
		}

		resp.Container.Auth.Ports = append(resp.Container.Auth.Ports, port)
	}

	// Network state
	// Tailscale state
	resp.Network = &model.NetworkState{
		Status:    model.NetworkState_Status(state.Components.Network.Health.Status),
		StatusMsg: state.Components.Network.Health.StatusMsg,

		Tailscale: &model.TailscaleState{
			Hosted:        state.Worker.Mode == runtime.ModeHostP2PVM,
			PeerHostnames: state.Components.Network.ConnectionState.Tailscale.Connection.PeerHostnames,

			Status:    model.TailscaleState_Status(state.Components.Network.ConnectionState.Tailscale.Status),
			StatusMsg: state.Components.Network.ConnectionState.Tailscale.StatusMsg,
		},
	}

	resp.Network.Tailscale.Ips = append(resp.Network.Tailscale.Ips, &model.TailscaleState_Ip{
		Ipv4: state.Components.Network.ConnectionState.Tailscale.Connection.IPs.IpV4,
		Ipv6: state.Components.Network.ConnectionState.Tailscale.Connection.IPs.IpV6,
	})

	// Piko state
	resp.Network.Piko = &model.PikoState{
		Status:    model.PikoState_Status(state.Components.Network.ConnectionState.Piko.Status),
		StatusMsg: state.Components.Network.ConnectionState.Piko.StatusMsg,
	}

	for _, e := range state.Components.Network.ConnectionState.Piko.Endpoints {
		respEndpoint := &model.PikoState_Endpoint{
			TemplatePort: e.Endpoint.Settings.PortID,
			ProxiedPort:  e.Endpoint.Port,
			Name:         e.Endpoint.Settings.Name,
			Link:         e.Link,
			Protocol:     model.PortProtocol(e.Protocol),
		}

		resp.Network.Piko.Endpoints = append(
			resp.Network.Piko.Endpoints, respEndpoint,
		)
	}

	// Iptables state
	resp.Network.Iptables = &model.IptablesState{
		Configured: state.Components.Network.ConnectionState.Iptables.Configured,
	}

	// General network state
	for _, p := range state.AppNetwork.ActivePorts {
		exposedPort := &model.NetworkState_Port{
			Title:    p.Title,
			Port:     p.Port,
			Protocol: model.PortProtocol(p.Protocol),
		}

		resp.Network.ExposedPorts = append(resp.Network.ExposedPorts, exposedPort)
	}

	resp.System = convertHardwareToHandlerType(state.System)

	return resp
}

func convertHardwareToHandlerType(hardware hardware.Info) *model.SystemInfo {
	result := &model.SystemInfo{}

	result = &model.SystemInfo{
		Ram: &model.RAM{
			TotalMem: float32(hardware.RAM.TotalMem),
			UsedMem:  float32(hardware.RAM.UsedMem),
			FreeMem:  float32(hardware.RAM.FreeMem),
		},

		Cpu: &model.CPU{
			Name:       hardware.CPU.Name,
			CoresCount: hardware.CPU.CoresCount,
		},

		Network: &model.Network{
			Ping:     float32(hardware.Network.Ping),
			Download: float32(hardware.Network.Download),
			Upload:   float32(hardware.Network.Upload),
		},

		Gpu: &model.GPU{
			NvidiaDriver: &model.GPU_NvidiaDriver{
				Installed: hardware.GPU.NvidiaDriver.Installed,
				Version:   hardware.GPU.NvidiaDriver.Version,
			},
		},

		Storage: &model.Storage{
			TotalMem: float32(hardware.Storage.TotalMem),
			UsedMem:  float32(hardware.Storage.UsedMem),
			FreeMem:  float32(hardware.Storage.FreeMem),
		},
	}

	for _, disk := range hardware.Storage.Types {
		result.Storage.Types = append(result.Storage.Types, model.Storage_DiskType(disk))
	}

	for _, gpu := range hardware.GPU.Cards {
		card := &model.GPU_Card{
			IsNvidia: gpu.IsNvidia,
			Name:     gpu.Name,
		}

		if gpu.VMem != nil {
			card.Mem = &model.GPU_Card_VMem{
				Total: float32(gpu.VMem.Total),
				Used:  float32(gpu.VMem.Used),
				Free:  float32(gpu.VMem.Free),
			}
		}

		result.Gpu.Cards = append(result.Gpu.Cards, card)
	}

	return result
}
