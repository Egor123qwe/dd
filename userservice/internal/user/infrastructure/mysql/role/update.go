package role

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/entity"
	"github.com/Interpuls/ifc2-service-farm/pkg/transactor/transaction"
)

func (r repo) Update(ctx context.Context, role entity.Role) error {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	executor := transaction.SelectExecutor(ctx, r.db)

	query := r.db.Rebind(`UPDATE roles SET name = ?, description = ?, permissions = ?, updated_at = ? WHERE id = ?`)

	roleDTO := role.ToDTO()
	now := time.Now().UTC()

	var description sql.NullString
	if roleDTO.Description != nil && *roleDTO.Description != "" {
		description = sql.NullString{
			String: *roleDTO.Description,
			Valid:  true,
		}
	}

	_, err := executor.ExecContext(
		ctx,
		query,
		roleDTO.Name,
		description,
		roleDTO.EncodedPermission,
		now,
		roleDTO.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	return nil
}
