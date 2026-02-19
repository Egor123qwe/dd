package status

import (
	"context"
	"errors"

	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/domain/model"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/handler/model/message"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/storage/db/repo"
)

func (s service) RentMerchant(ctx context.Context, msg message.MerchantRent) error {
	userID, err := s.storage.DB().Rent().MerchantUserID(ctx, msg.Content.SessionID)
	if err != nil {
		return err
	}

	keyUserID := string(msg.Content.RequestID) + string(msg.Content.SessionID)

	_, err = s.storage.Cache().Repo().Set(ctx, keyUserID, userID)
	if err != nil {
		return err
	}

	key := userID + string(msg.Content.SessionID)

	err = s.storage.Cache().Repo().SetOrUpdateMerchant(ctx, key, msg)
	if err != nil {
		return err
	}

	return nil

}

func (s service) SetStatusMerchant(ctx context.Context, msg message.MerchantRent) error {
	keyUserID := string(msg.Content.RequestID) + string(msg.Content.SessionID)

	userID, err := s.storage.Cache().Repo().Get(ctx, keyUserID)
	if err != nil {
		return err
	}

	key := userID + string(msg.Content.SessionID)

	switch msg.Content.Status {

	case string(message.RentStatusStopped):
		return s.storage.Cache().Repo().DeleteMerchant(ctx, key, keyUserID)

	default:
		err = s.storage.Cache().Repo().SetOrUpdateMerchant(ctx, key, msg)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s service) GetMerchantRent(ctx context.Context, msg message.MerchantMessage) (model.Merchant, error) {
	rent, err := s.storage.DB().Rent().ActiveRentByMerchantSession(ctx, msg.SessionID, msg.UserID)
	if err != nil {
		if errors.Is(err, repo.ErrRentNotFound) {
			return model.Merchant{}, ErrMerchantNoSessions
		}
		return model.Merchant{}, err
	}

	return model.Merchant{
		RequestID: rent.ID,
		SessionID: rent.SessionID,
		ClientID:  rent.ClientID,
		Status:    rent.Status,
		CreatedAt: rent.CreatedAt,
	}, nil
}

func (s service) SessionExpiredMerchant(ctx context.Context, msg message.FullMessage) error {
	userID, err := s.storage.DB().Rent().MerchantUserID(ctx, msg.Meta.Conn.SessionID)
	if err != nil {
		return err
	}

	key := userID + string(msg.Meta.Conn.SessionID)

	err = s.storage.Cache().Repo().Delete(ctx, key)
	if err != nil {
		return err
	}
	return nil
}

func (s service) MerchantSessions(ctx context.Context, userID string) ([]string, error) {
	sessions, err := s.storage.DB().Rent().MerchantSessions(ctx, userID)
	if err != nil {
		s.log.Error("failed to get merchant sessions", err.Error(), nil)

		return nil, err
	}

	return sessions, nil
}

func (s service) DetailMerchantSession(ctx context.Context, sessionID, userID string) (model.Session, error) {
	session, err := s.storage.DB().Rent().Session(ctx, sessionID, userID)
	if err != nil {
		s.log.Error("failed to get session", err.Error(), nil)

		return model.Session{}, err
	}

	return session, nil
}
