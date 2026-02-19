package balance

import (
	"context"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	balancerepo "gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/repo/balance"
)

var ErrInsufficientBalance = errors.New("insufficient balance")

const timeout = 10 * time.Second

type repo struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) balancerepo.Balance {
	return &repo{db: db}
}

// SettleRent в одной транзакции: списание cost с клиента, начисление cost*merchantRate продавцу.
// Таблица user_balance (user_id BIGINT/INT, balance DOUBLE PRECISION/NUMERIC), PRIMARY KEY или UNIQUE(user_id).
func (r *repo) SettleRent(ctx context.Context, clientUserID, merchantUserID int, cost, merchantRate float64) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx,
		`UPDATE user_balance SET balance = balance - $1 WHERE user_id = $2 AND balance >= $1`,
		cost, clientUserID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrInsufficientBalance
	}

	merchantAmount := cost * merchantRate
	_, err = tx.ExecContext(ctx,
		`INSERT INTO user_balance (user_id, balance) VALUES ($1, $2)
		 ON CONFLICT (user_id) DO UPDATE SET balance = user_balance.balance + EXCLUDED.balance`,
		merchantUserID, merchantAmount)
	if err != nil {
		return err
	}

	return tx.Commit()
}
