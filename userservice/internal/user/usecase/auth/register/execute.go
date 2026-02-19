package register

import (
	"context"
	"errors"
	"fmt"

	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/entity"
	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/service/perm"
	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
)

type Req struct {
	Email    string
	Password string
	Username string
	Name     string
	ZipCode  *string
	Phone    *string
}

type Resp struct {
	UserID       int
	AccessToken  string
	RefreshToken string
}

func (u usecase) Register(ctx context.Context, req Req) (Resp, error) {
	_, err := u.userRepo.GetByUsername(ctx, req.Username)
	if err == nil {
		return Resp{}, fmt.Errorf("%w: username already taken", errs.ErrInvalidRequest)
	}
	if !errors.Is(err, errs.ErrResourceNotFound) {
		return Resp{}, fmt.Errorf("failed to check username: %w", err)
	}

	_, err = u.userRepo.GetByEmail(ctx, req.Email)
	if err == nil {
		return Resp{}, fmt.Errorf("%w: email already taken", errs.ErrInvalidRequest)
	}
	if !errors.Is(err, errs.ErrResourceNotFound) {
		return Resp{}, fmt.Errorf("failed to check email: %w", err)
	}

	roleID := u.cfg.DefaultRoleID
	if roleID <= 0 {
		roleID = 1
	}
	_, err = u.rolesRepo.GetByID(ctx, roleID)
	if err != nil {
		return Resp{}, fmt.Errorf("default role not found: %w", err)
	}

	hashPassword, err := u.hasher.Hash(req.Password)
	if err != nil {
		return Resp{}, fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := entity.NewUser(
		0,
		[]int{roleID},
		hashPassword,
		req.Username,
		req.Email,
		req.Name,
		req.ZipCode,
		req.Phone,
		nil,
		false,
		nil,
		nil,
	)
	if err != nil {
		return Resp{}, err
	}

	id, err := u.userRepo.Create(ctx, user)
	if err != nil {
		return Resp{}, fmt.Errorf("failed to create user: %w", err)
	}

	user, err = u.userRepo.GetByID(ctx, id)
	if err != nil {
		return Resp{}, fmt.Errorf("failed to get created user: %w", err)
	}

	roles, err := u.rolesRepo.GetListByIDs(ctx, user.ToDTO().RoleIDs...)
	if err != nil {
		return Resp{}, fmt.Errorf("failed to get roles: %w", err)
	}

	p := perm.GetPermForUser(user, roles...)
	accessToken, err := u.jwt.Generate(int64(id), uint64(p), u.cfg.AtExp)
	if err != nil {
		return Resp{}, fmt.Errorf("failed to generate access token: %w", err)
	}
	refreshToken, err := u.jwt.Generate(int64(id), uint64(p), u.cfg.RtExp)
	if err != nil {
		return Resp{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return Resp{
		UserID:       id,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
