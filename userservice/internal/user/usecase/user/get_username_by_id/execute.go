package get_username_by_id

import (
	"context"
	"errors"

	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
)

func (u usecase) GetUsernameByID(ctx context.Context, id int) (string, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrResourceNotFound) {
			return "", nil
		}
		return "", err
	}
	return user.ToDTO().Username, nil
}
