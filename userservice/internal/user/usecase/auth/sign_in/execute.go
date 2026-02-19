package sign_in

import (
	"context"
	"fmt"
	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/entity"
	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/service/perm"
	"github.com/go-faster/errors"
	"time"
)

const (
	maxFailedSignInAttempts = 5
	signInFailedAttemptsTTL = 10 * time.Minute
)

var ErrTooManyAttempts = errors.New("To many failed attempts. Try again in 10 minutes")

type Req struct {
	Email    string
	Password string
}

type Resp struct {
	UserID       int
	AccessToken  string
	RefreshToken string
}

func (u usecase) SignIn(ctx context.Context, req Req) (Resp, error) {
	user, err := u.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, errs.ErrResourceNotFound) {
			return Resp{}, errs.ErrUnauthorized
		}

		fmt.Println(err.Error())
		return Resp{}, err
	}

	userDTO := user.ToDTO()

	if err := u.hasher.Compare(userDTO.HashPassword, req.Password); err != nil {
		return Resp{}, errs.ErrUnauthorized
	}

	roles, err := u.rolesRepo.GetListByIDs(ctx, userDTO.RoleIDs...)
	if err != nil {
		return Resp{}, fmt.Errorf("failed to get roles: %w", err)
	}

	if !userDTO.IsPasswordUpdated {
		filteredRoles := make([]entity.Role, 0, len(roles))
		for _, role := range roles {
			roleDTO := role.ToDTO()
			if roleDTO.Name != "SUI" {
				filteredRoles = append(filteredRoles, role)
			}
		}

		roles = filteredRoles
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

	result := Resp{
		UserID:       userDTO.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return result, nil
}
