package user

import (
	"context"
	"time"

	"github.com/Interpuls/ifc2-service-farm/pkg/transactor"
	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/entity"
	"github.com/jmoiron/sqlx"
)

const (
	requestTimeout = 10 * time.Second
)

type Repo interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error

	Create(ctx context.Context, req entity.User) (int, error)

	GetByID(ctx context.Context, userID int) (entity.User, error)
	GetByEmail(ctx context.Context, email string) (entity.User, error)
	GetByUsername(ctx context.Context, username string) (entity.User, error)
	GetList(ctx context.Context) ([]entity.User, error)

	Update(ctx context.Context, req entity.User) error

	Delete(ctx context.Context, userID int) error
}

type repo struct {
	db         *sqlx.DB
	transactor transactor.Transactor
}

func New(db *sqlx.DB) Repo {
	r := repo{
		db:         db,
		transactor: transactor.New(db),
	}

	return r
}

func (r repo) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.transactor.WithTransaction(ctx, fn)
}
