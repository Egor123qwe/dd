package update_user_password

import (
	"context"
	"errors"
	"fmt"
	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/service/perm"
	"github.com/Interpuls/ifc2-service-farm/pkg/permission"
)

var (
	ErrWrongPassword = errors.New("Current password is incorrect.")
)

type Req struct {
	ID              int
	CurrentPassword string
	NewPassword     string
}

func (u usecase) UpdateUserPassword(ctx context.Context, permit permission.Permit, req Req) error {
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

	if err := u.hasher.Compare(user.ToDTO().HashPassword, req.CurrentPassword); err != nil {
		return fmt.Errorf("%w: %w", errs.ErrInvalidRequest, ErrWrongPassword)
	}

	hashPassword, err := u.hasher.Hash(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	isPasswordUpdated := true
	err = user.ApplyUpdate(
		nil,
		&hashPassword,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&isPasswordUpdated,
	)

	if err != nil {
		return fmt.Errorf("failed to apply update: %w", err)
	}

	if err := u.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}
