package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/op/go-logging"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/api/auth"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/model"
	authModel "gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/model/auth"
)

var log = logging.MustGetLogger("auth")

type Auth interface {
	Auth(w http.ResponseWriter, r *http.Request) (string, error)
}

type service struct {
	AuthAPI auth.Auth
}

func New(authAPI auth.Auth, debug bool) Auth {
	if debug {
		return &mockService{}
	}

	return &service{
		AuthAPI: authAPI,
	}
}

func (a service) Auth(w http.ResponseWriter, r *http.Request) (string, error) {
	token, err := a.parseToken(r)
	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	log.Debugf("parsed token: %v", token)

	user, err := a.AuthAPI.ValidateToken(token.Value)
	if err != nil {
		return "", fmt.Errorf("failed to authenticate user: %w", err)
	}

	log.Debugf("authenticated user: %v", user.ID)
	return user.ID, nil
}

func (a service) parseToken(r *http.Request) (authModel.Token, error) {
	headers := r.Header
	// for app tokens
	appAuthHeader := headers.Get("Authorization")
	// for no-app tokens (merchant_client service)
	clientAuthHeader := headers.Get("Client-Authorization")
	// for browser WebSocket (cannot set headers): query param
	queryToken := r.URL.Query().Get("access_token")

	var value string
	var tokenType authModel.TokenType

	switch {
	case clientAuthHeader != "":
		tokenType = authModel.NoAppTokenType
		parts := strings.SplitN(clientAuthHeader, " ", 2)
		if len(parts) < 2 {
			return authModel.Token{}, model.ErrInvalidAuth
		}
		value = parts[1]
	case appAuthHeader != "":
		tokenType = authModel.BasicTokenType
		parts := strings.SplitN(appAuthHeader, " ", 2)
		if len(parts) < 2 {
			return authModel.Token{}, model.ErrInvalidAuth
		}
		value = parts[1]
	case queryToken != "":
		tokenType = authModel.BasicTokenType
		value = queryToken
	default:
		return authModel.Token{}, model.ErrMissingAuth
	}

	return authModel.Token{Value: value, Type: tokenType}, nil
}
