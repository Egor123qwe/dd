package balance

import (
	"context"

	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
)

func (r *repo) Withdraw(ctx context.Context, userID int, amount float64) error {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	query := r.db.Rebind(`
		UPDATE user_balance SET balance = balance - ? WHERE user_id = ? AND balance >= ?
	`)
	result, err := r.db.ExecContext(ctx, query, amount, userID, amount)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errs.ErrInsufficientBalance
	}
	return nil
}
