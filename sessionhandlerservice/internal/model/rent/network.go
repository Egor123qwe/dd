package rent

type NetworkSettings struct {
	Piko      *Piko
	Tailscale *Tailscale
}

type Piko struct {
	AuthKey   string
	Endpoints []PikoEndpoint
}

type PikoEndpoint struct {
	Title string
	Type  string

	TemplatePort int
	Endpoint     string
	Link         string
}

type Tailscale struct {
	ClientAuthKey   string
	MerchantAuthKey string
}
