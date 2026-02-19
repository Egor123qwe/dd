package content

type StopSessionReq struct {
	RequestID string `json:"request_id"`
	Reason    string `json:"reason,omitempty"`
}

type ShareP2PStopContent struct {
	SessionID string `json:"session_id"`
}
