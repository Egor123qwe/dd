package event

type (
	ExpireType string
	IDType     string
)

const (
	ExpiredSession     ExpireType = "expired-session"
	ExpiredClient      ExpireType = "expired-client"
	ExpiredRequest     ExpireType = "expired-request"
	ExpiredPaidRequest ExpireType = "expired-paid-request"
	ExpiredDeal        ExpireType = "expired-deal"

	SessionIDType     IDType = "session_id"
	ClientIDType      IDType = "client_user_id"
	RequestIDType     IDType = "request_id"
	PaidRequestIDType IDType = "paid_request_id"
	DealIDType        IDType = "deal_id"
)

var EventTypeMap = map[IDType]ExpireType{
	SessionIDType:     ExpiredSession,
	ClientIDType:      ExpiredClient,
	RequestIDType:     ExpiredRequest,
	PaidRequestIDType: ExpiredPaidRequest,
	DealIDType:        ExpiredDeal,
}
