package get_user_list

import (
	"context"
	"fmt"

	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/entity"
	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/service/perm"
	"github.com/Interpuls/ifc2-service-farm/pkg/permission"
)

func (u usecase) GetUserList(ctx context.Context, permit permission.Permit) ([]entity.User, error) {
	if !permit.HasPermission(permission.UsersManagementRead) {
		return nil, errs.ErrForbidden
	}

	users, err := u.userRepo.GetList(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	currentUser, err := u.userRepo.GetByID(ctx, permit.GetActorId())
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	currentUserRoles, err := u.rolesRepo.GetListByIDs(ctx, currentUser.ToDTO().RoleIDs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user roles: %w", err)
	}

	currentUserPerms := perm.GetPermForUser(currentUser, currentUserRoles...)

	var filteredUsers []entity.User
	for _, user := range users {
		targetUserRoles, err := u.rolesRepo.GetListByIDs(ctx, user.ToDTO().RoleIDs...)
		if err != nil {
			return nil, fmt.Errorf("failed to get target user roles: %w", err)
		}

		targetUserPerms := perm.GetPermForUser(user, targetUserRoles...)

		if currentUserPerms.HasLessOrEqualPermissions(targetUserPerms) {
			filteredUsers = append(filteredUsers, user)
		}
	}

	return filteredUsers, nil
}
