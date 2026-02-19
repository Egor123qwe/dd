package message

type ClientRent struct {
	Type    string        `json:"type"`
	Meta    ClientMeta    `json:"meta,omitempty"`
	Content ClientContent `json:"content"`
}

type ClientMeta struct {
	MessageID  string           `json:"message_id"`
	Status     string           `json:"status"`
	Connection ClientConnection `json:"conn"`
}

type ClientConnection struct {
	UserID string `json:"user_id"`
	ConnID string `json:"conn_id"`
}

type ClientContent struct {
	RequestID     string          `json:"request_id"`
	SessionID     string          `json:"session_id"`
	Settings      *ClientSettings `json:"settings,omitempty"`
	Status        string          `json:"status"`
	StatusMessage string          `json:"status_message,omitempty"`
	CreatedAt     string          `json:"created_at"`
}

type ClientSettings struct {
	Mode     string          `json:"mode,omitempty"`
	Template *ClientTemplate `json:"template,omitempty"`
	Network  *ClientNetwork  `json:"network,omitempty"`
}

type ClientTemplate struct {
	Authentication   *ClientAuthentication `json:"authentication,omitempty"`
	ShortDescription string                `json:"short_description,omitempty"`
}

type ClientAuthentication struct {
	Login    string `json:"login,omitempty"`
	Password string `json:"password,omitempty"`
}

type ClientNetwork struct {
	Piko      *ClientPiko      `json:"piko,omitempty"`
	Tailscale *ClientTailscale `json:"tailscale,omitempty"`
}

type ClientPiko struct {
	Endpoints []ClientEnpoints `json:"endpoints,omitempty"`
}

type ClientEnpoints struct {
	Title string `json:"title,omitempty"`
	Type  string `json:"type,omitempty"`
	Name  string `json:"name,omitempty"`
	Link  string `json:"link,omitempty"`
}

type ClientTailscale struct {
	AuthKey string `json:"auth_key,omitempty"`
}
