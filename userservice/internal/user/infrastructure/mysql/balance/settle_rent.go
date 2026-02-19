package balance

import (
	"context"

	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
)

// SettleRent в одной транзакции: списание cost с клиента, начисление cost*merchantRate продавцу.
func (r *repo) SettleRent(ctx context.Context, clientUserID, merchantUserID int, cost float64, merchantRate float64) error {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	withdrawQuery := r.db.Rebind(`
		UPDATE user_balance SET balance = balance - ? WHERE user_id = ? AND balance >= ?
	`)
	result, err := tx.ExecContext(ctx, withdrawQuery, cost, clientUserID, cost)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errs.ErrInsufficientBalance
	}

	merchantAmount := cost * merchantRate
	topUpQuery := r.db.Rebind(`
		INSERT INTO user_balance (user_id, balance) VALUES (?, ?)
		ON CONFLICT (user_id) DO UPDATE SET balance = user_balance.balance + EXCLUDED.balance
	`)
	_, err = tx.ExecContext(ctx, topUpQuery, merchantUserID, merchantAmount)
	if err != nil {
		return err
	}

	return tx.Commit()
}
