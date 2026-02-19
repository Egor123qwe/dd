package balance

import (
	"context"
	"database/sql"
)

func (r *repo) GetBalance(ctx context.Context, userID int) (float64, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	query := r.db.Rebind(`SELECT balance FROM user_balance WHERE user_id = ?`)
	var balance float64
	err := r.db.QueryRowxContext(ctx, query, userID).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return balance, nil
}
