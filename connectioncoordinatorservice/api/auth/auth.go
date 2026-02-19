package auth

import (
	"fmt"

	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/api/auth/model"
)

type Auth interface {
	ValidateToken(token string) (model.User, error)
}

type auth struct {
	config config
}

func New() Auth {
	config := newConfig()

	return &auth{
		config: config,
	}
}

func (a auth) withURL(endpoint string) string { return fmt.Sprintf("%s%s", a.config.URL, endpoint) }
