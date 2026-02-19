package get_balance

import (
	"context"

	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	"github.com/Interpuls/ifc2-service-farm/pkg/permission"
)

func (u *usecase) GetBalance(ctx context.Context, permit permission.Permit, userID int) (float64, error) {
	if permit.GetActorId() != userID {
		return 0, errs.ErrForbidden
	}
	return u.balanceRepo.GetBalance(ctx, userID)
}
