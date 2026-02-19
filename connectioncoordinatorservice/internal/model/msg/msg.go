package msg

import "encoding/json"

const ConnFieldJSON = "conn"

type IdType string

const (
	// ConnectionID used for sending message to specific connection
	ConnectionID IdType = "conn_id"

	// UserID used for sending message to all connections of specific user
	UserID IdType = "user_id"

	SessionID IdType = "session_id"

	// AllID used for sending message to all connections of all users
	AllID IdType = "all_ids"
)

type MSG struct {
	Data []byte
}

type Full struct {
	Type    string          `json:"type"`
	Meta    json.RawMessage `json:"meta"`
	Content json.RawMessage `json:"content"`
}

type Meta struct {
	Conn      Connection `json:"conn,omitempty"`
	SessionID string     `json:"session_id,omitempty"`
}

type Connection struct {
	UserID    string `json:"user_id,omitempty"`
	ConnID    string `json:"conn_id,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	Type      IdType `json:"type"`
}

type ErrorResp struct {
	Error string `json:"error"`
}

type InitResponse struct {
	Type    string          `json:"type"`
	Meta    Meta            `json:"meta"`
	Content json.RawMessage `json:"content"`
}
