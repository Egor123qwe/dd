package session

import "gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/msg/content/session/settings"

type MerchantRentStartReq struct {
	RequestID string `json:"request_id"`
	SessionID string `json:"session_id"`
	ClientID  string `json:"client_id"`

	Settings settings.Settings `json:"settings"`
}
