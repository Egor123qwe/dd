package get_username_by_id

import (
	"context"

	"github.com/Interpuls/ifc2-service-farm/internal/user/usecase/port"
)

type Usecase interface {
	GetUsernameByID(ctx context.Context, id int) (string, error)
}

type usecase struct {
	userRepo port.UserRepo
}

func New(userRepo port.UserRepo) Usecase {
	return &usecase{userRepo: userRepo}
}
