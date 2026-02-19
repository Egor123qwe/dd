package get_role_list

import (
	"context"
	"fmt"

	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/entity"
	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/service/perm"
	"github.com/Interpuls/ifc2-service-farm/pkg/permission"
)

func (u usecase) GetRoleList(ctx context.Context, permit permission.Permit) ([]entity.Role, error) {
	roles, err := u.roleRepo.GetList(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}

	currentUser, err := u.userRepo.GetByID(ctx, permit.GetActorId())
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	currentUserRoles, err := u.roleRepo.GetListByIDs(ctx, currentUser.ToDTO().RoleIDs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user roles: %w", err)
	}

	currentUserPerms := perm.GetPermForUser(currentUser, currentUserRoles...)

	var filteredRoles []entity.Role

	for _, role := range roles {
		if currentUserPerms.HasLessOrEqualPermissions(role.ToDTO().EncodedPermission) {
			filteredRoles = append(filteredRoles, role)
		}
	}

	return filteredRoles, nil
}
