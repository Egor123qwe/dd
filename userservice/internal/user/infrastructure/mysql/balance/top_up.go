package balance

import (
	"context"
)

func (r *repo) TopUp(ctx context.Context, userID int, amount float64) error {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	query := r.db.Rebind(`
		INSERT INTO user_balance (user_id, balance) VALUES (?, ?)
		ON CONFLICT (user_id) DO UPDATE SET balance = user_balance.balance + EXCLUDED.balance
	`)
	_, err := r.db.ExecContext(ctx, query, userID, amount)
	return err
}
