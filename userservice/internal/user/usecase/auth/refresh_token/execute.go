package refresh_token

import (
	"context"
	"errors"
	"fmt"
	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/service/perm"

	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
)

type Req struct {
	RefreshToken string
}

type Resp struct {
	RefreshToken string
	AccessToken  string
}

func (u usecase) RefreshToken(ctx context.Context, req Req) (Resp, error) {
	claims, err := u.jwt.Validate(req.RefreshToken)
	if err != nil {
		return Resp{}, errs.ErrUnauthorized
	}

	user, err := u.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, errs.ErrResourceNotFound) {
			return Resp{}, errs.ErrUnauthorized
		}

		return Resp{}, fmt.Errorf("failed to find user: %w", err)
	}

	userDTO := user.ToDTO()

	roles, err := u.rolesRepo.GetListByIDs(ctx, userDTO.RoleIDs...)
	if err != nil {
		return Resp{}, fmt.Errorf("failed to get roles: %w", err)
	}

	perm := perm.GetPermForUser(user, roles...)

	accessToken, err := u.jwt.Generate(int64(userDTO.ID), uint64(perm), u.cfg.AtExp)
	if err != nil {
		return Resp{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := u.jwt.Generate(int64(userDTO.ID), uint64(perm), u.cfg.RtExp)
	if err != nil {
		return Resp{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	resp := Resp{
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
	}

	return resp, nil
}
