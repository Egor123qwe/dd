package rent

import "gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model/rent"

type SettingsClient struct {
	Mode     MerchantMode   `json:"mode"`
	Template TemplateClient `json:"template"`
	Network  NetworkClient  `json:"network"`
}

type TemplateClient struct {
	ShortDescription string         `json:"short_description"`
	Authentication   Authentication `json:"authentication"`
}

type NetworkClient struct {
	Piko      *PikoClient      `json:"piko,omitempty"`
	Tailscale *TailscaleClient `json:"tailscale,omitempty"`
}

type PikoClient struct {
	Endpoints []PikoEndpointClient `json:"endpoints"`
}

type PikoEndpointClient struct {
	Title string `json:"title"`
	Type  string `json:"type"`

	Endpoint string `json:"name"`
	Link     string `json:"link"`
}

type TailscaleClient struct {
	AuthKey string `json:"auth_key"`
}

func ConvertToClientSettings(req rent.Settings) SettingsClient {
	result := SettingsClient{
		Mode: MerchantMode(req.Mode),

		Template: TemplateClient{
			ShortDescription: req.Template.Template.ShortDescription,

			Authentication: Authentication{
				Login:    req.Template.Authentication.Login,
				Password: req.Template.Authentication.Password,
			},
		},

		Network: NetworkClient{},
	}

	if req.Network.Tailscale != nil {
		result.Network.Tailscale = &TailscaleClient{
			AuthKey: req.Network.Tailscale.ClientAuthKey,
		}
	}

	if req.Network.Piko != nil {
		result.Network.Piko = &PikoClient{}

		for _, endpoint := range req.Network.Piko.Endpoints {
			result.Network.Piko.Endpoints = append(result.Network.Piko.Endpoints, PikoEndpointClient{
				Title: endpoint.Title,
				Type:  endpoint.Type,

				Endpoint: endpoint.Endpoint,
				Link:     endpoint.Link,
			})
		}
	}

	return result
}
