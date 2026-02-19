package user

import "errors"

type User struct {
	ID string `json:"user_id"`
}

var (
	ErrCreateRequest     = errors.New("failed create request to validate token")
	ErrSendRequest       = errors.New("failed to send request to validate token")
	ErrReadReqBody       = errors.New("failed to read response body")
	ErrUnmarshalRespBody = errors.New("failed to unmarshal response body")
)
