package message

type MerchantRent struct {
	Type    string          `json:"type"`
	Meta    MerchantMeta    `json:"meta"`
	Content MerchantContent `json:"content"`
}

type MerchantMeta struct {
	Status string             `json:"status"`
	Conn   MerchantConnection `json:"conn"`
}
type MerchantConnection struct {
	ConnID string `json:"conn_id"`
	Type   string `json:"type"`
}
type MerchantContent struct {
	RequestID string            `json:"request_id"`
	SessionID string            `json:"session_id"`
	Settings  *MerchantSettings `json:"settings"`
	ClientID  string            `json:"client_id"`
	Status    string            `json:"status"`
}

type MerchantSettings struct {
	Mode     string            `json:"mode"`
	Template *MerchantTemplate `json:"template"`
	Network  *MerchantNetwork  `json:"network"`
}
type MerchantTemplate struct {
	ID               string                  `json:"id"`
	Title            string                  `json:"title"`
	Description      string                  `json:"description"`
	ShortDescription string                  `json:"short_description"`
	ImageName        string                  `json:"image_name"`
	ImageTag         string                  `json:"image_tag"`
	Version          string                  `json:"version"`
	Ports            []MerchantPorts         `json:"ports"`
	Envs             *[]MerchantEnv           `json:"envs"`
	Volumes          []string                `json:"volumes"`
	UseGPU           bool                    `json:"use_gpu"`
	Authentication   *MerchantAuthentication `json:"authentication"`
}

type MerchantEnv struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type MerchantPorts struct {
	Auth  bool   `json:"auth"`
	Port  int    `json:"port"`
	Type  string `json:"type"`
	Title string `json:"title"`
}

type MerchantAuthentication struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type MerchantNetwork struct {
	Piko      *MerchantPiko      `json:"piko,omitempty"`
	Tailscale *MerchantTailscale `json:"tailscale,omitempty"`
}

type MerchantPiko struct {
	AuthKey  string              `json:"auth_key"`
	Endpoint []MerchantEndpoints `json:"endpoints"`
}

type MerchantEndpoints struct {
	Name         string `json:"name"`
	TemplatePort int    `json:"template_port"`
}

type MerchantTailscale struct {
	AuthKey string `json:"auth_key"`
}

type MerchantMessage struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
}
