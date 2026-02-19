package sharep2p

import (
	"errors"

	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/domain/hardware"
	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/domain/message"
)

type InitMerchantRequest struct {
	Type    string                     `json:"type"`
	Meta    message.Meta               `json:"meta"`
	Content ContentInitMerchantRequest `json:"content"`
}

type MetaInitMerchantRequest struct {
	ConnectionID string `json:"connection_id"`
}

type ContentInitMerchantRequest struct {
	Hardware hardware.Session   `json:"hardware"`
	Schedule Schedule           `json:"schedule"`
	Prepull  []hardware.Prepull `json:"prepull"`
	NodeName string             `json:"node_name"`
}

type Schedule struct {
	Type string `json:"type"`
}

type InitMerchantResponse struct {
	Type    string                      `json:"type"`
	Meta    message.Meta                `json:"meta"`
	Content ContentInitMerchantResponse `json:"content"`
}

type ContentInitMerchantResponse struct {
	SessionID string  `json:"session_id"`
	Tier      string  `json:"tier"`
	Price     float32 `json:"price"`
}

type ReadyMerchantRequest struct {
	Type    string               `json:"type"`
	Meta    message.Meta         `json:"meta"`
	Content ContentReadyMerchant `json:"content"`
}

type ContentReadyMerchant struct {
	SessionID string `json:"session_id,omitempty"`
}

type ReadyMerchantResponse struct {
	Type    string               `json:"type"`
	Meta    message.Meta         `json:"meta"`
	Content ContentReadyMerchant `json:"content"`
}

type KeepAlive struct {
	Type    string       `json:"type"`
	Meta    message.Meta `json:"meta"`
	Content struct{}     `json:"content"`
}

type Stop struct {
	Type    string               `json:"type"`
	Meta    message.Meta         `json:"meta"`
	Content ContentReadyMerchant `json:"content"`
}

var (
	ErrSessionNotFound           = errors.New("session not found")
	ErrInavlidRequest            = errors.New("invalid request")
	ErrCanNotCreateSession       = errors.New("can not create session")
	ErrCanNotSetSessionStatus    = errors.New("can not set session status")
	ErrCanNotDeleteSession       = errors.New("can not delete session")
	ErrUnidentifiedHardware      = errors.New("unidentified hardware: not all device components are recognized; please contact support to add your hardware to the price list")
)
