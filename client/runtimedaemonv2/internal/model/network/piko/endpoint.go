package piko

import "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/docker/template"

type Endpoint struct {
	Settings EndpointSettings
	Port     string
	Protocol template.Protocol
}

type EndpointSettings struct {
	PortID string // its like port id here (to know for which port (from template) is this endpoint)
	Name   string
}

type EndpointInfo struct {
	Protocol template.Protocol
	Endpoint Endpoint
	Link     string
}
