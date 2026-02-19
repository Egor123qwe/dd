package user

import (
	"context"
	"fmt"

	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	"github.com/Interpuls/ifc2-service-farm/pkg/transactor/transaction"
)

func (r repo) Delete(ctx context.Context, userID int) error {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	executor := transaction.SelectExecutor(ctx, r.db)

	query := r.db.Rebind(`DELETE FROM users WHERE id = ?`)

	result, err := executor.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w: user with id %d not found", errs.ErrResourceNotFound, userID)
	}

	return nil
}
