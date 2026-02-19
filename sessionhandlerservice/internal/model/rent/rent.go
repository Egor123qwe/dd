package rent

import "time"

type Status string

const (
	PendingRentStatus  Status = "pending"
	StartedRentStatus  Status = "started"
	FinishedRentStatus Status = "finished"
)

type MerchantMode string

const (
	ProxyMode MerchantMode = "proxy"
	P2PMode   MerchantMode = "p2p"
)

type Rent struct {
	ID        string
	SessionID string
	Status    Status
	ClientId  string
	CreatedAt time.Time
	DeletedAt time.Time
}

type RequestSettings struct {
	Mode       MerchantMode
	TemplateID string
}

type Settings struct {
	Mode     MerchantMode
	Network  NetworkSettings
	Template TemplateSettings
}
