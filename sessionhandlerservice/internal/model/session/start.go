package session

import (
	"time"

	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model/rent"
)

type InitReq struct {
	Client    Client
	SessionID string

	Settings rent.RequestSettings
}

type InitResp struct {
	Merchant  Merchant
	RequestID string

	CreatedAt time.Time
	Settings  rent.Settings
}

type StartReq struct {
	RequestID string
}

type StartResp struct {
	Client    Client
	SessionID string
}
