package status

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/domain/model"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/handler/model/message"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/storage/cache/repo"
)

// GetClientRent возвращает активные аренды клиента только из БД (без Redis).
func (s service) GetClientRent(ctx context.Context, msg message.ClientMessage) ([]model.Client, error) {
	clientID := msg.UserID
	rents, err := s.storage.DB().Rent().ActiveRentsByClientID(ctx, clientID)
	if err != nil {
		return nil, err
	}
	if len(rents) == 0 {
		return nil, ErrClienttNoRents
	}
	out := make([]model.Client, 0, len(rents))
	for _, r := range rents {
		merchantUserID, _ := s.storage.DB().Rent().MerchantUserID(ctx, r.SessionID)
		client := model.Client{
			RequestID:      r.ID,
			SessionID:      r.SessionID,
			MerchantUserID: merchantUserID,
			Status:         r.Status,
			CreatedAt:      r.CreatedAt,
		}
		creds, err := s.clientCredentials(ctx, client)
		if err != nil {
			client.TemplateID = ""
			client.Paid = false
		} else {
			client = creds
		}
		out = append(out, client)
	}
	return out, nil
}

func (s service) RentClient(ctx context.Context, msg message.ClientRent) error {
	key := string(msg.Meta.Connection.UserID)

	switch msg.Content.Status {

	case string(message.RentStatusPending):
		rent, err := s.storage.DB().Rent().Rent(ctx, msg.Content.RequestID)
		if err != nil {
			return err
		}
		msg.Content.CreatedAt = rent.CreatedAt

	case string(message.RentStatusStopped):
		return s.storage.Cache().Repo().DeleteClientRent(ctx, key, msg.Content.SessionID)
	}

	err := s.storage.Cache().Repo().SetOrUpdateClient(ctx, key, msg)
	if err != nil {
		return err
	}

	return nil
}

func (s service) GetClientRentBySession(ctx context.Context, msg message.ClientMessage) (model.Client, error) {
	key := string(msg.UserID)

	rent, err := s.storage.Cache().Repo().GetClientRentBySession(ctx, key, msg.SessionID)

	if err != nil {
		if errors.Is(err, repo.ErrNoRent) || errors.Is(err, redis.Nil) {
			return model.Client{}, ErrClienttNoRents
		}

		return model.Client{}, err
	}

	rent, err = s.clientCredentials(ctx, rent)
	if err != nil {
		return model.Client{}, err
	}

	return rent, nil
}

func (s service) clientCredentials(ctx context.Context, rent model.Client) (model.Client, error) {
	template_id, err := s.storage.DB().Rent().TemplateIDForClient(ctx, rent.RequestID)
	if err != nil {
		return model.Client{}, err
	}

	paid, err := s.storage.DB().Rent().PaidFlagForClient(ctx, rent.RequestID)
	if err != nil {
		return model.Client{}, err
	}

	rent.TemplateID = template_id
	rent.Paid = paid

	return rent, nil
}
