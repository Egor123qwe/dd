package shareP2P

import (
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/model/msg/content/hardware"
)

type Schedule struct {
	Type string `json:"type"`
}

type InitMerchantReq struct {
	Hardware hardware.Session   `json:"hardware"`
	Schedule Schedule           `json:"schedule"`
	Prepull  []hardware.Prepull `json:"prepull"`
	NodeName string             `json:"node_name"`
}

type InitMerchantResp struct {
	SessionID string  `json:"session_id"`
	Tier      string  `json:"tier"`
	Price     float32 `json:"price"`
}
