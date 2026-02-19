package role

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

	Create(ctx context.Context, req entity.Role) (int, error)

	GetList(ctx context.Context) ([]entity.Role, error)
	GetByID(ctx context.Context, id int) (entity.Role, error)
	GetByName(ctx context.Context, name string) (entity.Role, error)
	GetListByIDs(ctx context.Context, ids ...int) ([]entity.Role, error)

	Update(ctx context.Context, role entity.Role) error

	Delete(ctx context.Context, roleID int) error
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
