package event

type Event string

// Events from session-handler-service
const (
	// MerchantStartRentEvent is used for start session by merchant
	MerchantStartRentEvent Event = "merchant-start-rent"

	// SessionStatusUpdatedEvent is used for session status responses
	SessionStatusUpdatedEvent Event = "session-status-updated"

	RentRequestStatusUpdatedEvent Event = "rent-request-status-updated"

	StopSession Event = "stop-session"
)

// Events in resource-pull-service
const (
	ShareP2PInit  Event = "share-p2p-init"
	ShareP2PReady Event = "share-p2p-ready"

	ShareP2PStop   Event = "share-p2p-stop"
	ExpiredSession Event = "expired-session"

	KeepAlive Event = "keepalive"
)
