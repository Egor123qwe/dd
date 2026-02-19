package get_user

import (
	"context"
	"fmt"

	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/entity"
	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/service/perm"
	"github.com/Interpuls/ifc2-service-farm/pkg/permission"
)

func (u usecase) GetUser(ctx context.Context, permit permission.Permit, id int) (entity.User, error) {
	if !permit.HasPermission(permission.UsersManagementRead) && !(id == permit.GetActorId()) {
		return entity.User{}, errs.ErrForbidden
	}

	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return entity.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	if permit.HasPermission(permission.UsersManagementRead) {
		currentUser, err := u.userRepo.GetByID(ctx, permit.GetActorId())
		if err != nil {
			return entity.User{}, fmt.Errorf("failed to get current user: %w", err)
		}

		currentUserRoles, err := u.rolesRepo.GetListByIDs(ctx, currentUser.ToDTO().RoleIDs...)
		if err != nil {
			return entity.User{}, fmt.Errorf("failed to get current user roles: %w", err)
		}

		targetUserRoles, err := u.rolesRepo.GetListByIDs(ctx, user.ToDTO().RoleIDs...)
		if err != nil {
			return entity.User{}, fmt.Errorf("failed to get target user roles: %w", err)
		}

		currentUserPerms := perm.GetPermForUser(currentUser, currentUserRoles...)
		targetUserPerms := perm.GetPermForUser(user, targetUserRoles...)

		if !currentUserPerms.HasLessOrEqualPermissions(targetUserPerms) {
			return entity.User{}, errs.ErrForbidden
		}
	}

	return user, nil
}
