package settings

type Mode string

const (
	ProxyMode Mode = "proxy"
	P2PMode   Mode = "p2p"
)

type PortType string

const (
	HTTP PortType = "http"
	TCP  PortType = "tcp"
	UDP  PortType = "udp"
)

type EnvType string

const (
	SSHKey       EnvType = "ssh_key"
	Username     EnvType = "username"
	Password     EnvType = "password"
	HfToken      EnvType = "hf_token"
	CivitaiToken EnvType = "civitai_token"
	Other        EnvType = "other"
)

type Settings struct {
	Mode     Mode     `json:"mode"`
	Template Template `json:"template"`
	Network  Network  `json:"network"`
}

type Template struct {
	ID string `json:"id"`

	Title            string `json:"title"`
	Description      string `json:"description"`
	ShortDescription string `json:"short_description"`
	Version          string `json:"version"`

	Type string `json:"type"`

	ImageName string `json:"image_name"`
	ImageTag  string `json:"image_tag"`

	Ports   []TemplatePort `json:"ports"`
	Envs    []TemplateEnvs `json:"envs"`
	Volumes []string       `json:"volumes"`
	UseGPU  bool           `json:"use_gpu"`

	Authentication Authentication `json:"authentication"`
}

type TemplateEnvs struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

type TemplatePort struct {
	Auth  bool   `json:"auth"`
	Port  int    `json:"port"`
	Type  string `json:"type"`
	Title string `json:"title"`
}

type Network struct {
	Piko      *Piko      `json:"piko,omitempty"`
	Tailscale *Tailscale `json:"tailscale,omitempty"`
}

type Piko struct {
	AuthKey   string         `json:"auth_key"`
	Endpoints []PikoEndpoint `json:"endpoints"`
}

type PikoEndpoint struct {
	TemplatePort int    `json:"template_port"`
	Endpoint     string `json:"name"`
}

type Tailscale struct {
	AuthKey string `json:"auth_key"`
}

type Authentication struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
