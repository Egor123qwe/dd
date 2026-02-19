package piko

type Status int32

const (
	Running Status = iota
	Stopped
	Error
)

var StatusMap = map[Status]string{
	Running: "Running",
	Stopped: "Stopped",
	Error:   "Error",
}

type Settings struct {
	Endpoints []EndpointSettings
	AuthKey   string
}

type ConnectReq struct {
	Endpoints []Endpoint
	AuthKey   string
}

type State struct {
	Status    Status
	StatusMsg string

	Endpoints []EndpointInfo
}

type Info struct {
	Availability bool
	Version      string
}
