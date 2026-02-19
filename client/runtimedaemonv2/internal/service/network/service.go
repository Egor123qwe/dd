package network

import (
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/template"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network/piko"
)

func (s service) getPikoEndpoints(requestedEndpoints []piko.EndpointSettings, appPorts []network.ActivePort) []piko.Endpoint {
	result := make([]piko.Endpoint, 0)

	for _, endpoint := range requestedEndpoints {
		for _, port := range appPorts {
			if endpoint.PortID == port.PortID && s.availableProxyProtocol(port.Protocol) {
				port := piko.Endpoint{
					Settings: endpoint,
					Port:     port.Port,
					Protocol: port.Protocol,
				}

				result = append(result, port)
			}
		}
	}

	return result
}

func (s service) availableProxyProtocol(protocol template.Protocol) bool {
	return protocol == template.HTTP || protocol == template.TCP
}
