package get_balance

import (
	"context"

	"github.com/Interpuls/ifc2-service-farm/internal/user/usecase/port"
	"github.com/Interpuls/ifc2-service-farm/pkg/permission"
)

type Usecase interface {
	GetBalance(ctx context.Context, permit permission.Permit, userID int) (float64, error)
}

type usecase struct {
	balanceRepo port.BalanceRepo
}

func New(balanceRepo port.BalanceRepo) Usecase {
	return &usecase{balanceRepo: balanceRepo}
}
