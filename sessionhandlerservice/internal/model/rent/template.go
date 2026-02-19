package rent

type PortType string

const (
	HTTP PortType = "http"
	TCP  PortType = "tcp"
	UDP  PortType = "udp"
)

type Template struct {
	ID string

	Title            string
	Description      string
	ShortDescription string
	Version          string

	Type string

	ImageName string
	ImageTag  string
	Ports     []Port
	Envs      []Env
	Volumes   []string
	UseGPU    bool
}

type Env struct {
	Key   string
	Value string
	Type  string
}

type Port struct {
	Auth  bool
	Port  int
	Type  string
	Title string
}

type TemplateSettings struct {
	Template       Template
	Authentication Authentication
}

type Authentication struct {
	Login    string
	Password string
}
