package delete_user

import (
	"context"
	"fmt"

	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/service/perm"
	"github.com/Interpuls/ifc2-service-farm/pkg/permission"
)

func (u usecase) DeleteUser(ctx context.Context, permit permission.Permit, id int) error {
	if !permit.HasPermission(permission.UsersManagementWrite) && !(id == permit.GetActorId()) {
		return errs.ErrForbidden
	}

	targetUser, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get target user: %w", err)
	}

	currentUser, err := u.userRepo.GetByID(ctx, permit.GetActorId())
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	currentUserRoles, err := u.rolesRepo.GetListByIDs(ctx, currentUser.ToDTO().RoleIDs...)
	if err != nil {
		return fmt.Errorf("failed to get current user roles: %w", err)
	}

	targetUserRoles, err := u.rolesRepo.GetListByIDs(ctx, targetUser.ToDTO().RoleIDs...)
	if err != nil {
		return fmt.Errorf("failed to get target user roles: %w", err)
	}

	currentUserPerms := perm.GetPermForUser(currentUser, currentUserRoles...)
	targetUserPerms := perm.GetPermForUser(targetUser, targetUserRoles...)

	if !currentUserPerms.HasLessOrEqualPermissions(targetUserPerms) {
		return errs.ErrForbidden
	}

	if err := u.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}
