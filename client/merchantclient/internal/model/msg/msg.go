package msg

import "encoding/json"

type Status string

const (
	Ok    Status = "ok"
	Error Status = "err"
)

type MSG struct {
	// Type for different events in package event
	Type string `json:"type"`
	// Meta for store different ids
	Meta Meta `json:"meta"`
	// Content for different events in package content
	Content json.RawMessage `json:"content"`
}

type Meta struct {
	Status    string `json:"status,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	MessageID string `json:"message_id,omitempty"`
	Err       *Err   `json:"err,omitempty"`
}

type Err struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"msg,omitempty"`
}
