package session

type MerchantStatus string

const (
	PendingMerchantStatus  MerchantStatus = "pending"
	ReadyMerchantStatus    MerchantStatus = "ready"
	ReservedMerchantStatus MerchantStatus = "reserved"
)
