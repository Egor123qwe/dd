package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gitlab.roy9.ru/roy9/backend/core/feedbackservice/internal/config"
	"gitlab.roy9.ru/roy9/backend/core/feedbackservice/internal/domain/user"
)

type Service struct {
	client http.Client
	cfg    config.AuthConfig
}

func New(client http.Client, cfg config.AuthConfig) *Service {
	svc := &Service{
		client: client,
		cfg:    cfg,
	}

	return svc
}

func (s *Service) Validate(ctx context.Context, token string) (user.User, error) {
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s%s", s.cfg.URL, "/api/v1/auth/validate"),
		nil,
	)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	if err != nil {
		return user.User{}, user.ErrCreateRequest
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return user.User{}, user.ErrSendRequest
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return user.User{}, user.ErrReadReqBody
	}

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("%s", string(body))

		switch resp.StatusCode {
		case http.StatusForbidden:
			return user.User{}, err

		default:
			return user.User{}, err
		}
	}

	var usr user.User

	if err := json.Unmarshal(body, &usr); err != nil {
		return user.User{}, user.ErrUnmarshalRespBody
	}

	return usr, nil
}
