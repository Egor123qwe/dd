package rent

import (
	"context"
	"database/sql"
	"errors"
)

func (r repo) GetClientRents(ctx context.Context, clientID string) ([]string, error) {
	result := make([]string, 0)

	query := `SELECT id 
			  FROM rent 
			  WHERE client_id = $1 AND deleted_at IS NULL`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		clientID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRentNotFound
		}

		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var requestID string

		if err := rows.Scan(
			&requestID,
		); err != nil {
			return nil, err
		}

		result = append(result, requestID)
	}

	return result, nil
}
