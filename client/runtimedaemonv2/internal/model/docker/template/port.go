package template

type Protocol int32

const (
	HTTP Protocol = iota
	TCP
	UDP
)

type Port struct {
	Port     string // will be used as portID (should be equal to local container port)
	Title    string
	Protocol Protocol

	AuthAvailable bool
}
