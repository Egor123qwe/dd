package rent

import (
	"context"
	"fmt"
	"strings"

	"github.com/dchest/uniuri"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model/rent"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/util/jwt"
)

func (s service) network(ctx context.Context, mode rent.MerchantMode, ports []rent.Port) (rent.NetworkSettings, error) {
	var result rent.NetworkSettings

	switch mode {
	case rent.P2PMode:
		tailscale, err := s.tailscale()
		if err != nil {
			return rent.NetworkSettings{}, err
		}

		result.Tailscale = &tailscale

	case rent.ProxyMode:
		piko, err := s.piko(ports)
		if err != nil {
			return rent.NetworkSettings{}, err
		}

		result.Piko = &piko
	}

	return result, nil
}

func (s service) piko(ports []rent.Port) (rent.Piko, error) {
	token, err := jwt.Generate(s.config.pikoTokenExp, s.config.pikoSecretKey)
	if err != nil {
		return rent.Piko{}, err
	}

	result := rent.Piko{
		AuthKey: token,
	}

	for _, port := range ports {
		// now we only support http
		if rent.PortType(port.Type) != rent.HTTP && rent.PortType(port.Type) != rent.TCP {
			continue
		}

		endpoint := strings.ToLower(uniuri.NewLen(20))

		result.Endpoints = append(result.Endpoints, rent.PikoEndpoint{
			Title: port.Title,
			Type:  port.Type,

			TemplatePort: port.Port,
			Endpoint:     endpoint,
			Link:         fmt.Sprintf(s.config.linkTemplate, endpoint),
		})
	}

	return result, nil
}

func (s service) tailscale() (rent.Tailscale, error) {
	result := rent.Tailscale{
		ClientAuthKey:   "TODO",
		MerchantAuthKey: "TODO",
	}

	return result, nil
}
