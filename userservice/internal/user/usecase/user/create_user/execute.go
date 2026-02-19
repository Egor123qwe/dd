package create_user

import (
	"context"
	"errors"
	"fmt"
	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/service/perm"

	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/entity"
	"github.com/Interpuls/ifc2-service-farm/pkg/permission"
)

type Req struct {
	Username string
	RoleIDs  []int
	Email    string
	Password string
	Name     string
	ZipCode  *string
	Phone    *string
}

func (u usecase) CreateUser(ctx context.Context, permit permission.Permit, req Req) (int, error) {
	if !permit.HasPermission(permission.UsersManagementWrite) {
		return 0, errs.ErrForbidden
	}

	// Check if username already exists
	_, err := u.userRepo.GetByUsername(ctx, req.Username)
	if err == nil {
		return 0, fmt.Errorf("%w: user with username '%s' already exists", errs.ErrInvalidRequest, req.Username)
	}
	if !errors.Is(err, errs.ErrResourceNotFound) {
		return 0, fmt.Errorf("failed to check username: %w", err)
	}

	// Check if email already exists
	_, err = u.userRepo.GetByEmail(ctx, req.Email)
	if err == nil {
		return 0, fmt.Errorf("%w: user with email '%s' already exists", errs.ErrInvalidRequest, req.Email)
	}
	if !errors.Is(err, errs.ErrResourceNotFound) {
		return 0, fmt.Errorf("failed to check email: %w", err)
	}

	roles, err := u.rolesRepo.GetListByIDs(ctx, req.RoleIDs...)
	if err != nil {
		return 0, fmt.Errorf("failed to get roles: %w", err)
	}

	userCreator, err := u.userRepo.GetByID(ctx, permit.GetActorId())
	if err != nil {
		return 0, fmt.Errorf("failed to get user: %w", err)
	}

	// check if user has all permissions to create_user user with requested roles
	if !permit.HasAllPermissions(permission.DecodePermissions(perm.GetPermForUser(userCreator, roles...))...) {
		return 0, errs.ErrForbidden
	}

	hashPassword, err := u.hasher.Hash(req.Password)
	if err != nil {
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := entity.NewUser(
		0,
		req.RoleIDs,
		hashPassword,
		req.Username,
		req.Email,
		req.Name,
		req.ZipCode,
		req.Phone,
		nil,   // LanguageID
		false, // IsPasswordUpdated - always false for new users
		nil,   // CreatedAt
		nil,   // UpdatedAt
	)
	if err != nil {
		return 0, err
	}

	id, err := u.userRepo.Create(ctx, user)
	if err != nil {
		return 0, fmt.Errorf("failed to create_user user: %w", err)
	}

	return id, nil
}
