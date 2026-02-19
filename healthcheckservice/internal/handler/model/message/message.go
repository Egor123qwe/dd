package message

type status string

const (
	EventTypeSessionUpdated    = "session-status-updated"
	EventTypeClientStartRent   = "client-start-rent"
	EventTypeMerchantStartRent = "merchant-start-rent"

	RentStatusRunnuing status = "running"
	RentStatusPending  status = "pending"
	RentStatusStopped  status = "stopped"
)

type FullMessage struct {
	Type    string  `json:"type"`
	Meta    Meta    `json:"meta"`
	Content Content `json:"content"`
}

type Meta struct {
	Status string     `json:"status"`
	Conn   Connection `json:"conn"`
}

type Connection struct {
	SessionID    string `json:"session_id"`
	Type         string `json:"type"`
	ConnectionID string `json:"conn_id"`
}

type Content struct {
	SessionID string `json:"session_id"`
	RequestID string `json:"request_id"`
}

type Settings struct {
	Mode string `json:"mode"`
}

type ClientMessage struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
}

type RentResponse struct {
	Content ClientContent `json:"content"`
}

type LeaseResponse struct {
	Content MerchantContent `json:"content"`
}
