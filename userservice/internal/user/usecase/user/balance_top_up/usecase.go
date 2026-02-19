package balance_top_up

import (
	"context"

	"github.com/Interpuls/ifc2-service-farm/internal/user/usecase/port"
	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	"github.com/Interpuls/ifc2-service-farm/pkg/permission"
)

type Usecase interface {
	TopUp(ctx context.Context, permit permission.Permit, userID int, amount float64) error
}

type usecase struct {
	balanceRepo port.BalanceRepo
}

func New(balanceRepo port.BalanceRepo) Usecase {
	return &usecase{balanceRepo: balanceRepo}
}

func (u *usecase) TopUp(ctx context.Context, permit permission.Permit, userID int, amount float64) error {
	if permit.GetActorId() != userID {
		return errs.ErrForbidden
	}
	if amount <= 0 {
		return errs.ErrInvalidRequest
	}
	return u.balanceRepo.TopUp(ctx, userID, amount)
}
