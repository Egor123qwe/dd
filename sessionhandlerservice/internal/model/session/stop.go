package session

type StopReq struct {
	RequestID string
	Reason    string
}

type StopResp struct {
	Client   Client
	Merchant Merchant

	SessionID string
}
