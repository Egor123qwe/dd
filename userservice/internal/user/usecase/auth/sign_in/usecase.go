package sign_in

import (
	"context"
	"time"

	"github.com/Interpuls/ifc2-service-farm/internal/user/usecase/port"
	"github.com/Interpuls/ifc2-service-farm/pkg/hasher"
	"github.com/Interpuls/ifc2-service-farm/pkg/jwt"
)

type Usecase interface {
	SignIn(ctx context.Context, req Req) (Resp, error)
}

type usecase struct {
	userRepo  port.UserRepo
	rolesRepo port.RoleRepo

	hasher hasher.Hasher
	jwt    jwt.JWT

	cfg Config
}

type Config struct {
	AtExp time.Duration
	RtExp time.Duration
}

func New(
	userRepo port.UserRepo,
	rolesRepo port.RoleRepo,
	jwt jwt.JWT,
	cfg Config,
) Usecase {
	u := usecase{
		userRepo:  userRepo,
		rolesRepo: rolesRepo,

		hasher: hasher.New(),
		jwt:    jwt,

		cfg: cfg,
	}

	return u
}
