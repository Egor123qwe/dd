package delete_user

import (
	"context"

	"github.com/Interpuls/ifc2-service-farm/internal/user/usecase/port"
	"github.com/Interpuls/ifc2-service-farm/pkg/hasher"
	"github.com/Interpuls/ifc2-service-farm/pkg/permission"
	"github.com/go-playground/validator/v10"
)

type Usecase interface {
	DeleteUser(ctx context.Context, permit permission.Permit, id int) error
}

type usecase struct {
	userRepo  port.UserRepo
	rolesRepo port.RoleRepo

	hasher    hasher.Hasher
	validator *validator.Validate
}

func New(
	userRepo port.UserRepo,
	rolesRepo port.RoleRepo,
) Usecase {
	u := usecase{
		userRepo:  userRepo,
		rolesRepo: rolesRepo,

		hasher:    hasher.New(),
		validator: validator.New(),
	}

	return u
}
