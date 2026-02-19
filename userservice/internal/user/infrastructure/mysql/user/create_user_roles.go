package user

import (
	"context"
	"fmt"
	"github.com/Interpuls/ifc2-service-farm/pkg/transactor/transaction"
)

func (r repo) createUserRoles(ctx context.Context, userID int, roleIDs []int) error {
	executor := transaction.SelectExecutor(ctx, r.db)

	query := r.db.Rebind("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)")

	for _, roleID := range roleIDs {
		if _, err := executor.ExecContext(ctx, query, userID, roleID); err != nil {
			return fmt.Errorf("failed to insert user_role: %w", err)
		}
	}

	return nil
}
