package delete_role

import (
	"context"
	"fmt"

	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/service/perm"
	"github.com/Interpuls/ifc2-service-farm/pkg/permission"
)

func (u usecase) DeleteRole(ctx context.Context, permit permission.Permit, roleID int) error {
	if !permit.HasPermission(permission.UserRolesManagementWrite) {
		return errs.ErrForbidden
	}

	role, err := u.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	if role.ToDTO().IsDefault {
		return fmt.Errorf("%w:cannot delete default role", errs.ErrForbidden)
	}

	currentUser, err := u.userRepo.GetByID(ctx, permit.GetActorId())
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	currentUserRoles, err := u.roleRepo.GetListByIDs(ctx, currentUser.ToDTO().RoleIDs...)
	if err != nil {
		return fmt.Errorf("failed to get current user roles: %w", err)
	}

	currentUserPerms := perm.GetPermForUser(currentUser, currentUserRoles...)

	if !currentUserPerms.HasLessOrEqualPermissions(role.ToDTO().EncodedPermission) {
		return errs.ErrForbidden
	}

	if err := u.roleRepo.Delete(ctx, roleID); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}
