package user

import (
	"context"
	"fmt"
	"github.com/Interpuls/ifc2-service-farm/pkg/transactor/transaction"
)

func (r repo) deleteUserRoles(ctx context.Context, userID int) error {
	executor := transaction.SelectExecutor(ctx, r.db)

	query := r.db.Rebind("DELETE FROM user_roles WHERE user_id = ?")

	if _, err := executor.ExecContext(ctx, query, userID); err != nil {
		return fmt.Errorf("failed to delete user_roles: %w", err)
	}

	return nil
}
