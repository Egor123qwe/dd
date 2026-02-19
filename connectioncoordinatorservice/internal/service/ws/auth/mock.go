package auth

import (
	"net/http"

	"github.com/google/uuid"
)

type mockService struct{}

func (a mockService) Auth(w http.ResponseWriter, r *http.Request) (string, error) {
	return uuid.New().String(), nil
}
