package rent

import (
	"context"
	"database/sql"
	"errors"
)

func (r repo) GetMerchantRent(ctx context.Context, sessionID string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	var requestID string

	query := `SELECT 
    				id
			  FROM rent WHERE session_id = $1 AND deleted_at IS NULL`

	err := r.db.QueryRowContext(
		ctx,
		query,
		sessionID,
	).Scan(
		&requestID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrRentNotFound
		}

		return "", err
	}

	return requestID, nil
}

func (r repo) ChangeMerchantStatus(ctx context.Context, sessionID string, status string) error {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	query := `UPDATE session SET status = $1 WHERE id = $2`

	res, err := r.db.ExecContext(ctx, query,
		status,
		sessionID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrMerchantNotFound
	}

	return nil
}
