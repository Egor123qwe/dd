package event

type EventType string

const (
	ShareP2PInit  EventType = "share-p2p-init"
	ShareP2PReady EventType = "share-p2p-ready"
	KeepAlive     EventType = "keep-alive"
	Expired       EventType = "expired-session"
	ShareP2PStop  EventType = "share-p2p-stop"
)
