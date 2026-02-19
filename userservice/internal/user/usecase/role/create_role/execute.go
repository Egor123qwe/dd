package create_role

import (
	"context"
	"fmt"
	"time"

	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/entity"
	"github.com/Interpuls/ifc2-service-farm/pkg/permission"
)

type Req struct {
	Name             string
	Description      string
	PermissionsNames []string
}

func (u usecase) CreateRole(ctx context.Context, permit permission.Permit, req Req) (int, error) {
	if !permit.HasPermission(permission.UserRolesManagementWrite) {
		return 0, errs.ErrForbidden
	}

	perms, err := permission.GetPermissionsByNames(req.PermissionsNames...)
	if err != nil {
		return 0, fmt.Errorf("failed to get permissions: %w", err)
	}

	if !permit.HasAllPermissions(perms...) {
		return 0, errs.ErrForbidden
	}

	for _, perm := range perms {
		if !perm.IsExternal() {
			return 0, fmt.Errorf("%w: permission '%s' is not external (cannot set)", errs.ErrInvalidRequest, perm.String())
		}
	}

	encodedPerm, err := permission.EncodePermissions(perms...)
	if err != nil {
		return 0, fmt.Errorf("failed to encode permissions: %w", err)
	}

	role := entity.NewRole(
		0,
		req.Name,
		req.Description,
		encodedPerm,
		false,
		time.Now().UTC(),
		nil,
	)

	id, err := u.roleRepo.Create(ctx, role)
	if err != nil {
		return 0, fmt.Errorf("failed to create_user role: %w", err)
	}

	return id, nil
}
