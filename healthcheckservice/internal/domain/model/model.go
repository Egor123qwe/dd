package model

import (
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/handler/model/message"
)

const (
	ParseLayout = "2006-01-02 15:04:05.000000"
)

type Rent struct {
	SessionID string `json:"session" db:"session_id"`
	ID        string `json:"request_id"`
	CreatedAt string `json:"created_at" db:"created_at"`
	DeletedAt string `json:"deleted_at" db:"deleted_at"`
	ClientID  string `json:"client_id" db:"client_id"`
	Status    string `json:"status"`
}

type Merchant struct {
	RequestID string                    `json:"request_id,omitempty" db:"request_id"`
	SessionID string                    `json:"session_id,omitempty" db:"session_id"`
	ClientID  string                    `json:"client_id,omitempty" db:"client_id"`
	Settings  *message.MerchantSettings `json:"settings,omitempty"`
	Status    string                    `json:"status,omitempty" db:"status"`
	CreatedAt string                    `json:"created_at,omitempty" db:"created_at"`
}

type Client struct {
	RequestID      string                  `json:"request_id" db:"request_id"`
	SessionID      string                  `json:"session_id" db:"session_id"`
	MerchantUserID string                  `json:"merchant_user_id,omitempty"`
	Settings       *message.ClientSettings `json:"settings,omitempty"`
	Status         string                  `json:"status" db:"status"`
	CreatedAt      string                  `json:"created_at" db:"created_at"`
	Paid           bool                    `json:"paid" db:"paid"`
	TemplateID     string                  `json:"template_id" db:"template_id"`
}

type User struct {
	UserID string `json:"user_id"`
}
