package content

import "time"

type SessionStatus string

const (
	RunningStatus SessionStatus = "running"
	PendingStatus SessionStatus = "pending"
	StoppedStatus SessionStatus = "stopped"
)

type MerchantRentStatus string

const (
	RunningMerchantRentStatus MerchantRentStatus = "running"
	ErrorMerchantRentStatus   MerchantRentStatus = "error"
)

type Initiator string

const (
	ClientInitiator   Initiator = "client"
	MerchantInitiator Initiator = "merchant"
	ServerInitiator   Initiator = "server"

	UnknownInitiator Initiator = "unknown"
)

type SessionStatusResp struct {
	RequestID string `json:"request_id"`
	SessionID string `json:"session_id"`

	Status    SessionStatus `json:"status"`
	StatusMsg string        `json:"status_msg"`

	CreatedAt *time.Time `json:"created_at,omitempty"`

	// who was initiator of status update (used in case of session stop)
	Initiator Initiator `json:"initiator,omitempty"`
}

type RentRequestStatusUpdatedReq struct {
	RequestID string `json:"request_id"`

	Status    MerchantRentStatus `json:"status"`
	StatusMsg string             `json:"status_msg"`
}
