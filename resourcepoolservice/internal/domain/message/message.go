package message

import (
	"encoding/json"
)

type FullMessage struct {
	Type    string          `json:"type"`
	Meta    Meta            `json:"meta"`
	Content json.RawMessage `json:"content"`
}

type Meta struct {
	Status    string     `json:"status,omitempty"`
	SessionID string     `json:"session_id,omitempty"`
	MessageID string     `json:"message_id,omitempty"`
	Conn      Connection `json:"conn,omitempty"`
	Err       *Err        `json:"err,omitempty"`
}

type Connection struct {
	ConnectionID string `json:"conn_id,omitempty"`
	UserID       string `json:"user_id,omitempty"`
	SessionID    string `json:"session_id,omitempty"`
	Type         string `json:"type"`
}

type Err struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"msg,omitempty"`
}

type PingPongMessage struct {
	Type    string   `json:"type"`
	Meta    Meta     `json:"meta"`
	Content struct{} `json:"content"`
}
