package message

type Message struct {
	Type    string  `json:"type"`
	Meta    Meta    `json:"meta"`
	Content Content `json:"content"`
}

type Meta struct {
	Status    string     `json:"status,omitempty"`
	Err       Err        `json:"err,omitempty"`
	Conn      Connection `json:"conn,omitempty"`
	SessionID string     `json:"session_id,omitempty"`
}

type Connection struct {
	ConnectionID string `json:"conn_id,omitempty"`
	UserID       string `json:"user_id,omitempty"`
	SessionID    string `json:"session_id,omitempty"`
	Type         string `json:"type,omitempty"`
}

type Err struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type Content struct {
	RequestID string `json:"request_id,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	DealID    string `json:"deal_id,omitempty"`
	Reason    string `json:"reason,omitempty"`
}
