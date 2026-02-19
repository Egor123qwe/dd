package balance_withdraw

import (
	"context"
	"errors"

	"github.com/Interpuls/ifc2-service-farm/internal/user/usecase/port"
	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	"github.com/Interpuls/ifc2-service-farm/pkg/permission"
)

type Usecase interface {
	Withdraw(ctx context.Context, permit permission.Permit, userID int, amount float64) error
}

type usecase struct {
	balanceRepo port.BalanceRepo
}

func New(balanceRepo port.BalanceRepo) Usecase {
	return &usecase{balanceRepo: balanceRepo}
}

func (u *usecase) Withdraw(ctx context.Context, permit permission.Permit, userID int, amount float64) error {
	if permit.GetActorId() != userID {
		return errs.ErrForbidden
	}
	if amount <= 0 {
		return errs.ErrInvalidRequest
	}
	err := u.balanceRepo.Withdraw(ctx, userID, amount)
	if err != nil && errors.Is(err, errs.ErrInsufficientBalance) {
		return errs.ErrInsufficientBalance
	}
	return err
}
