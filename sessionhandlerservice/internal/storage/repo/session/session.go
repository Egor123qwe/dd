package session

import (
	"context"

	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model/session"
)

type Session interface {
	GetMerchantIDs(ctx context.Context, sessionID string) (session.Merchant, error)
	GetSessionIDByMerchant(ctx context.Context, merchantID string) (string, error)
	GetTotalPrice(ctx context.Context, sessionID string) (float64, error)
}
