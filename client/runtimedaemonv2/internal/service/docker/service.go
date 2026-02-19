package docker

import (
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/auth"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/template"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/docker/api"
)

func (s service) isProxyPortsCorrect(current []string, needed []string) bool {
	if len(current) != len(needed) {
		return false
	}

	for i := 0; i < len(current); i++ {
		if current[i] != needed[i] {
			return false
		}
	}

	return true
}

func (s service) filterContainerPortsByProxability(containerPorts []api.Port, templatesPorts []template.Port) []string {
	var result []string

	for _, tp := range templatesPorts {
		exist := false

		for _, cp := range containerPorts {
			if cp.Local == tp.Port {
				if tp.AuthAvailable && s.protocolAuthAvailable(tp.Protocol) {
					result = append(result, cp.Host)
				}

				exist = true
				break
			}
		}

		if !exist {
			log.Errorf(
				"invalid template or container issue: port %s not found in container", tp.Port,
			)
		}
	}

	return result
}

func (s service) containerPortsFromProxyPorts(ports []auth.Port) []string {
	result := make([]string, 0, len(ports))

	for _, port := range ports {
		result = append(result, port.InPort)
	}

	return result
}

func (s service) protocolAuthAvailable(protocol template.Protocol) bool {
	return protocol == template.HTTP || protocol == template.TCP
}
