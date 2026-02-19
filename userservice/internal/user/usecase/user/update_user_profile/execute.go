package update_user_profile

import (
	"context"
	"fmt"

	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/service/perm"
	"github.com/Interpuls/ifc2-service-farm/pkg/permission"
)

type Req struct {
	ID         int
	RoleIDs    *[]int
	Username   *string
	Email      *string
	Name       *string
	ZipCode    **string
	Phone      **string
	LanguageID **int
}

func (u usecase) UpdateUserProfile(ctx context.Context, permit permission.Permit, req Req) error {
	if !permit.HasPermission(permission.UsersManagementWrite) && !(req.ID == permit.GetActorId()) {
		return errs.ErrForbidden
	}

	user, err := u.userRepo.GetByID(ctx, req.ID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	currentUser, err := u.userRepo.GetByID(ctx, permit.GetActorId())
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	currentUserRoles, err := u.rolesRepo.GetListByIDs(ctx, currentUser.ToDTO().RoleIDs...)
	if err != nil {
		return fmt.Errorf("failed to get current user roles: %w", err)
	}

	targetUserRoles, err := u.rolesRepo.GetListByIDs(ctx, user.ToDTO().RoleIDs...)
	if err != nil {
		return fmt.Errorf("failed to get target user roles: %w", err)
	}

	currentUserPerms := perm.GetPermForUser(currentUser, currentUserRoles...)
	targetUserPerms := perm.GetPermForUser(user, targetUserRoles...)

	if !currentUserPerms.HasLessOrEqualPermissions(targetUserPerms) {
		return errs.ErrForbidden
	}

	if req.RoleIDs != nil {
		roles, err := u.rolesRepo.GetListByIDs(ctx, *req.RoleIDs...)
		if err != nil {
			return fmt.Errorf("failed to get roles: %w", err)
		}

		// check if user has all permissions to create_user user with requested roles
		if !permit.HasAllPermissions(permission.DecodePermissions(perm.GetPermForUser(user, roles...))...) {
			return errs.ErrForbidden
		}
	}

	err = user.ApplyUpdate(
		req.RoleIDs,
		nil, // hashPassword - not updated in profile update
		req.Username,
		req.Email,
		req.Name,
		req.ZipCode,
		req.Phone,
		req.LanguageID,
		nil, // isPasswordUpdated - not changed in profile update
	)

	if err != nil {
		return fmt.Errorf("failed to apply update: %w", err)
	}

	if err := u.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}
