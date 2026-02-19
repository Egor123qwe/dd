package balance

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

const requestTimeout = 10 * time.Second

type Repo interface {
	GetBalance(ctx context.Context, userID int) (float64, error)
	TopUp(ctx context.Context, userID int, amount float64) error
	Withdraw(ctx context.Context, userID int, amount float64) error
	SettleRent(ctx context.Context, clientUserID, merchantUserID int, cost float64, merchantRate float64) error
}

type repo struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) Repo {
	return &repo{db: db}
}
