package rent

type MerchantMode string

const (
	ProxyMode MerchantMode = "proxy"
	P2PMode   MerchantMode = "p2p"
)

type Authentication struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type SettingsReq struct {
	Mode       MerchantMode `json:"mode"`
	TemplateID string       `json:"template_id"`
}
