package settle_rent

import (
	"context"

	"github.com/Interpuls/ifc2-service-farm/internal/user/usecase/port"
)

const DefaultMerchantRate = 0.9

type Usecase interface {
	SettleRent(ctx context.Context, clientUserID, merchantUserID int, cost float64, merchantRate float64) error
}

type usecase struct {
	balanceRepo port.BalanceRepo
}

func New(balanceRepo port.BalanceRepo) Usecase {
	return &usecase{balanceRepo: balanceRepo}
}

func (u *usecase) SettleRent(ctx context.Context, clientUserID, merchantUserID int, cost float64, merchantRate float64) error {
	if merchantRate <= 0 || merchantRate > 1 {
		merchantRate = DefaultMerchantRate
	}
	return u.balanceRepo.SettleRent(ctx, clientUserID, merchantUserID, cost, merchantRate)
}
