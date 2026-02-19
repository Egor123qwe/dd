package update_role

import (
	"context"
	"fmt"

	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/entity"
	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/service/perm"
	"github.com/Interpuls/ifc2-service-farm/pkg/permission"
)

type Req struct {
	ID               int
	Name             string
	Description      string
	PermissionsNames []string
}

func (u usecase) UpdateRole(ctx context.Context, permit permission.Permit, req Req) error {
	if !permit.HasPermission(permission.UserRolesManagementWrite) {
		return errs.ErrForbidden
	}

	perms, err := permission.GetPermissionsByNames(req.PermissionsNames...)
	if err != nil {
		return fmt.Errorf("failed to get permissions: %w", err)
	}

	if !permit.HasAllPermissions(perms...) {
		return errs.ErrForbidden
	}

	encodedPerm, err := permission.EncodePermissions(perms...)
	if err != nil {
		return fmt.Errorf("failed to encode permissions: %w", err)
	}

	existingRole, err := u.roleRepo.GetByID(ctx, req.ID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	existingRoleDTO := existingRole.ToDTO()
	if existingRoleDTO.IsDefault {
		return fmt.Errorf("%w:cannot update default role", errs.ErrForbidden)
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

	if !currentUserPerms.HasLessOrEqualPermissions(encodedPerm) {
		return errs.ErrForbidden
	}

	role := entity.NewRole(
		req.ID,
		req.Name,
		req.Description,
		encodedPerm,
		existingRoleDTO.IsDefault,
		existingRoleDTO.CreatedAt,
		existingRoleDTO.UpdatedAt,
	)

	if err := u.roleRepo.Update(ctx, role); err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	return nil
}
