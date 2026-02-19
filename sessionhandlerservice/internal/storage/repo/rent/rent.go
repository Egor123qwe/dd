package rent

import (
	"context"

	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model/rent"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/pkg/sqlt/transaction"
)

type Rent interface {
	WithTransaction(ctx context.Context) (Rent, transaction.Service, error)

	Create(ctx context.Context, rent rent.Rent, configuration rent.Settings) (rent.Rent, error)
	Get(ctx context.Context, id string) (rent.Rent, error)
	Stop(ctx context.Context, id string, reason string, cost float64) error

	ChangeStatus(ctx context.Context, id string, status rent.Status) error
	ChangeMerchantStatus(ctx context.Context, sessionID string, status string) error

	GetSettings(ctx context.Context, requestID string) (rent.Settings, error)

	GetClientRents(ctx context.Context, clientID string) ([]string, error)
	GetMerchantRent(ctx context.Context, sessionID string) (string, error)
}
