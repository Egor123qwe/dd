package user

import (
	"context"
	"fmt"
	"github.com/Interpuls/ifc2-service-farm/pkg/transactor/transaction"
)

func (r repo) getUserRoleIDs(ctx context.Context, userID int) ([]int, error) {
	query := r.db.Rebind("SELECT role_id FROM user_roles WHERE user_id = ?")

	executor := transaction.SelectExecutor(ctx, r.db)

	rows, err := executor.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user_roles: %w", err)
	}

	defer rows.Close()

	var roles []int

	for rows.Next() {
		var roleID int

		if err := rows.Scan(&roleID); err != nil {
			return nil, err
		}

		roles = append(roles, roleID)
	}

	return roles, nil
}
