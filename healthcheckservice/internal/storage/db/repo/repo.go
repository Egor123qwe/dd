package repo

import (
	"context"
	"errors"

	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/domain/model"
)

var (
	ErrRentNotFound        = errors.New("rent record is empty or not found")
	ErrSessionDoesNotExist = errors.New("session does not exist")
	ErrPrepullDoesNotExist = errors.New("prepull does not exist")
)

type Rent interface {
	Rent(ctx context.Context, requestID string) (model.Rent, error)
	ActiveRentsByClientID(ctx context.Context, clientID string) ([]model.Rent, error)
	ActiveRentByMerchantSession(ctx context.Context, sessionID, userID string) (model.Rent, error)
	MerchantUserID(ctx context.Context, sessionID string) (string, error)
	MerchantExist(ctx context.Context, sessionID, userID string) (bool, error)
	MerchantSessions(ctx context.Context, userID string) ([]string, error)
	PaidFlagForClient(ctx context.Context, rentID string) (bool, error)
	TemplateIDForClient(ctx context.Context, rentID string) (string, error)
	Session(ctx context.Context, sessionID, userID string) (model.Session, error)
}
