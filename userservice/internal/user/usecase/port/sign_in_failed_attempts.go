package port

import "time"

type SignInFailedAttempt struct {
	CreatedAt time.Time `json:"created_at"`
}

type SignInFailedAttemptsRepo interface {
	Add(userID int, attempt SignInFailedAttempt) error
	GetList(userID int) ([]SignInFailedAttempt, error)
	Delete(userID int) error
}
