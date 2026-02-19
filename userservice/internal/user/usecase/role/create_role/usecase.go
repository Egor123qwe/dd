package create_role

import (
	"context"
	"github.com/Interpuls/ifc2-service-farm/internal/user/usecase/port"
	"github.com/Interpuls/ifc2-service-farm/pkg/permission"
)

type Usecase interface {
	CreateRole(ctx context.Context, permit permission.Permit, req Req) (int, error)
}

type usecase struct {
	roleRepo port.RoleRepo
}

func New(
	roleRepo port.RoleRepo,
) Usecase {
	u := usecase{
		roleRepo: roleRepo,
	}

	return u
}
