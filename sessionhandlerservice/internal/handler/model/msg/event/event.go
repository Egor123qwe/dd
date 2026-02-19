package event

type Event string

// Events from client to server
const (
	StartSessionEvent Event = "start-session"
	StopSessionEvent  Event = "stop-session"

	// RentRequestStatusUpdatedEvent is used for session status responses
	RentRequestStatusUpdatedEvent Event = "rent-request-status-updated"
)

// Events from server to client
const (
	// ErrorEvent is used for error responses
	ErrorEvent Event = "error"

	// MerchantStartRentEvent is used for start session by merchant
	MerchantStartRentEvent Event = "merchant-start-rent"

	// ClientStartRentEvent is used for start session by client (send after merchant start session)
	ClientStartRentEvent Event = "client-start-rent"

	// SessionStatusUpdatedEvent is used for session status responses
	SessionStatusUpdatedEvent Event = "session-status-updated"
)

// Events from server to other services
const (
	// InitRent is used for init session (starting tracking of request_id)
	InitRent Event = "init-rent"
)

// Events from ttl-service
const (
	ExpiredRentEvent    Event = "expired-request"
	ExpiredClientEvent  Event = "expired-client"
	ExpiredSessionEvent Event = "expired-session"
)

// Events from resource-pull-Service
const (
	ShareP2PStop Event = "share-p2p-stop"
)
