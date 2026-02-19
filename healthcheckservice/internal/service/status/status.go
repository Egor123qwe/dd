package status

import (
	"context"
	"errors"
	"log/slog"

	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/domain/model"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/handler/model/message"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/storage"
)

var (
	ErrMerchantNoSessions = errors.New("no sessions")
	ErrClienttNoRents     = errors.New("no rent")
	ErrInternalServer     = errors.New("internal server")
)

type Status interface {
	RentMerchant(ctx context.Context, msg message.MerchantRent) error
	RentClient(ctx context.Context, msg message.ClientRent) error
	SetStatusMerchant(ctx context.Context, msg message.MerchantRent) error
	GetClientRent(ctx context.Context, msg message.ClientMessage) ([]model.Client, error)
	GetClientRentBySession(ctx context.Context, msg message.ClientMessage) (model.Client, error)
	GetMerchantRent(ctx context.Context, msg message.MerchantMessage) (model.Merchant, error)
	SessionExpiredMerchant(ctx context.Context, msg message.FullMessage) error
	MerchantSessions(ctx context.Context, userID string) ([]string, error)
	DetailMerchantSession(ctx context.Context, sessionID, userID string) (model.Session, error)
}

type service struct {
	storage storage.Storage
	log     slog.Logger
}

func New(log slog.Logger, storage storage.Storage) Status {
	return service{
		storage: storage,
		log:     log,
	}
}
