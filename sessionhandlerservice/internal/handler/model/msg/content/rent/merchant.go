package rent

import (
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model/rent"
)

type SettingsMerchant struct {
	Mode     MerchantMode     `json:"mode"`
	Template TemplateMerchant `json:"template"`
	Network  NetworkMerchant  `json:"network"`
}

type TemplateMerchant struct {
	ID string `json:"id"`

	Title            string `json:"title"`
	Description      string `json:"description"`
	ShortDescription string `json:"short_description"`
	Version          string `json:"version"`

	Type string `json:"type"`

	ImageName string `json:"image_name"`
	ImageTag  string `json:"image_tag"`

	Ports   []TemplatePortMerchant `json:"ports"`
	Envs    []TemplateEnvsMerchant `json:"envs"`
	Volumes []string               `json:"volumes"`
	UseGPU  bool                   `json:"use_gpu"`

	// Minimum requirements for provider (to filter who can run this template)
	MinCPU                int32    `json:"min_cpu,omitempty"`
	MinRAMBytes            uint64   `json:"min_ram_bytes,omitempty"`
	MinStorageBytes        uint64   `json:"min_storage_bytes,omitempty"`
	MinVolumeStorageBytes  []uint64 `json:"min_volume_storage_bytes,omitempty"`

	Authentication Authentication `json:"authentication"`
}

type TemplateEnvsMerchant struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

type TemplatePortMerchant struct {
	Auth  bool   `json:"auth"`
	Port  int    `json:"port"`
	Type  string `json:"type"`
	Title string `json:"title"`
}

type NetworkMerchant struct {
	Piko      *PikoMerchant      `json:"piko,omitempty"`
	Tailscale *TailscaleMerchant `json:"tailscale,omitempty"`
}

type PikoMerchant struct {
	AuthKey   string                 `json:"auth_key"`
	Endpoints []PikoEndpointMerchant `json:"endpoints"`
}

type PikoEndpointMerchant struct {
	TemplatePort int    `json:"template_port"`
	Endpoint     string `json:"name"`
}

type TailscaleMerchant struct {
	AuthKey string `json:"auth_key"`
}

func ConvertToMerchantSettings(req rent.Settings) SettingsMerchant {
	result := SettingsMerchant{
		Mode: MerchantMode(req.Mode),

		Template: TemplateMerchant{
			ID: req.Template.Template.ID,

			Title:            req.Template.Template.Title,
			Description:      req.Template.Template.Description,
			ShortDescription: req.Template.Template.ShortDescription,
			Version:          req.Template.Template.Version,

			Type: req.Template.Template.Type,

			ImageName: req.Template.Template.ImageName,
			ImageTag:  req.Template.Template.ImageTag,

			Volumes: req.Template.Template.Volumes,
			UseGPU:  req.Template.Template.UseGPU,

			MinCPU:               req.Template.Template.MinCPU,
			MinRAMBytes:          req.Template.Template.MinRAMBytes,
			MinStorageBytes:      req.Template.Template.MinStorageBytes,
			MinVolumeStorageBytes: req.Template.Template.MinVolumeStorageBytes,

			Authentication: Authentication{
				Login:    req.Template.Authentication.Login,
				Password: req.Template.Authentication.Password,
			},
		},

		Network: NetworkMerchant{},
	}

	for _, env := range req.Template.Template.Envs {
		result.Template.Envs = append(result.Template.Envs, TemplateEnvsMerchant{
			Key:   env.Key,
			Value: env.Value,
			Type:  env.Type,
		})
	}

	for _, port := range req.Template.Template.Ports {
		result.Template.Ports = append(result.Template.Ports, TemplatePortMerchant{
			Auth:  port.Auth,
			Port:  port.Port,
			Type:  port.Type,
			Title: port.Title,
		})
	}

	if req.Network.Tailscale != nil {
		result.Network.Tailscale = &TailscaleMerchant{
			AuthKey: req.Network.Tailscale.MerchantAuthKey,
		}
	}

	if req.Network.Piko != nil {
		result.Network.Piko = &PikoMerchant{
			AuthKey: req.Network.Piko.AuthKey,
		}

		for _, endpoint := range req.Network.Piko.Endpoints {
			result.Network.Piko.Endpoints = append(result.Network.Piko.Endpoints, PikoEndpointMerchant{
				TemplatePort: endpoint.TemplatePort,
				Endpoint:     endpoint.Endpoint,
			})
		}
	}

	return result
}
