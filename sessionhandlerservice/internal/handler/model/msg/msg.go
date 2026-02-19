package msg

import "encoding/json"

type Status string

const (
	Ok    Status = "ok"
	Error Status = "err"
)

type IdType string

const (
	// ConnectionID used for sending message to specific connection
	ConnectionID IdType = "conn_id"

	// UserID used for sending message to all connections of specific user
	UserID IdType = "user_id"

	// AllID used for sending message to all connections of all users
	AllID IdType = "all_ids"
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
	Status    string     `json:"status,omitempty"`
	MessageID string     `json:"message_id,omitempty"`
	SessionID string     `json:"session_id,omitempty"`
	Conn      Connection `json:"conn,omitempty"`
	Err       *Err       `json:"err,omitempty"`
}

type Connection struct {
	UserID string `json:"user_id,omitempty"`
	ConnID string `json:"conn_id,omitempty"`
	Type   IdType `json:"type"`
}

type Err struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"msg,omitempty"`
}
