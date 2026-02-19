package content

import "gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/handler/model/msg/content/rent"

type InitSessionReq struct {
	SessionID string `json:"session_id"`

	Settings rent.SettingsReq `json:"settings"`
}

type MerchantRentStartReq struct {
	RequestID string `json:"request_id"`
	SessionID string `json:"session_id"`
	ClientID  string `json:"client_id"`

	Settings rent.SettingsMerchant `json:"settings"`
}

// ClientRentStartReq used to send request to start session for client
type ClientRentStartReq struct {
	RequestID string `json:"request_id"`
	SessionID string `json:"session_id"`

	Settings rent.SettingsClient `json:"settings"`
}

type WatchRentReq struct {
	RequestID string `json:"request_id"`
}
