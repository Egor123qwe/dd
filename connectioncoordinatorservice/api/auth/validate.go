package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/api/auth/model"
	errorsList "gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/model"
)

func (a auth) ValidateToken(token string) (model.User, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		a.withURL("/api/v1/machine/token/validate"),
		nil,
	)
	if err != nil {
		return model.User{}, fmt.Errorf("%w: %w", errorsList.ErrFailedCreateRequest, err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return model.User{}, fmt.Errorf("%w: %w", errorsList.ErrFailedSendRequest, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.User{}, fmt.Errorf("failed read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		errMsg := errors.New(string(body))
		switch resp.StatusCode {
		case http.StatusForbidden:
			return model.User{}, fmt.Errorf("%w: %w", errorsList.ErrBadClaims, errMsg)
		case http.StatusInternalServerError:
			return model.User{}, fmt.Errorf("%w: %w", errorsList.ErrInternalServerError, errMsg)
		}
		return model.User{}, fmt.Errorf("failed validate token: %w", errMsg)
	}

	var user model.User
	if err := json.Unmarshal(body, &user); err != nil {
		return model.User{}, fmt.Errorf("%w: %w", errorsList.ErrFailedToCreateResponse, err)
	}
	return user, nil
}
