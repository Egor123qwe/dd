package delete_role

import (
	"context"

	"github.com/Interpuls/ifc2-service-farm/internal/user/usecase/port"
	"github.com/Interpuls/ifc2-service-farm/pkg/permission"
)

type Usecase interface {
	DeleteRole(ctx context.Context, permit permission.Permit, roleID int) error
}

type usecase struct {
	roleRepo port.RoleRepo
	userRepo port.UserRepo
}

func New(
	roleRepo port.RoleRepo,
	userRepo port.UserRepo,
) Usecase {
	u := usecase{
		roleRepo: roleRepo,
		userRepo: userRepo,
	}

	return u
}
