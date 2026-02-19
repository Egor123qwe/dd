package role

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/entity"
	"github.com/Interpuls/ifc2-service-farm/pkg/transactor/transaction"
)

func (r repo) Create(ctx context.Context, req entity.Role) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	executor := transaction.SelectExecutor(ctx, r.db)

	query := r.db.Rebind(`INSERT INTO roles (name, description, permissions, is_default, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?) RETURNING id`)

	roleDTO := req.ToDTO()
	now := time.Now().UTC()

	var description sql.NullString
	if roleDTO.Description != nil && *roleDTO.Description != "" {
		description = sql.NullString{
			String: *roleDTO.Description,
			Valid:  true,
		}
	}

	var id int
	err := executor.QueryRowxContext(
		ctx,
		query,
		roleDTO.Name,
		description,
		roleDTO.EncodedPermission,
		roleDTO.IsDefault,
		now,
		now,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create_user role: %w", err)
	}

	return id, nil
}
