package port

import (
	"context"

	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/entity"
)

type UserRepo interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error

	Create(ctx context.Context, req entity.User) (int, error)

	GetByID(ctx context.Context, userID int) (entity.User, error)
	GetByEmail(ctx context.Context, email string) (entity.User, error)
	GetByUsername(ctx context.Context, username string) (entity.User, error)
	GetList(ctx context.Context) ([]entity.User, error)

	Update(ctx context.Context, req entity.User) error

	Delete(ctx context.Context, userID int) error
}
