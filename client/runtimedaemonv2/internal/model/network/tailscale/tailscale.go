package tailscale

type Status int32

const (
	Running Status = iota
	Stopped
)

var StatusMap = map[Status]string{
	Running: "Running",
	Stopped: "Stopped",
}

type Settings struct {
	ClientID string
	AuthKey  string
}

type State struct {
	Status    Status
	StatusMsg string

	Connection ConnectionState
}

type Info struct {
	Availability bool
	Version      string
}

type ConnectionState struct {
	IPs           IP
	PeerHostnames []string
}

type IP struct {
	IpV4 string
	IpV6 string
}
