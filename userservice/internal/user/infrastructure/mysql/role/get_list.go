package role

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/entity"
	"github.com/Interpuls/ifc2-service-farm/pkg/permission"
	"github.com/Interpuls/ifc2-service-farm/pkg/transactor/transaction"
	"github.com/Interpuls/ifc2-service-farm/pkg/util"
)

func (r repo) GetList(ctx context.Context) ([]entity.Role, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	executor := transaction.SelectExecutor(ctx, r.db)

	query := `SELECT id, name, description, permissions, is_default, created_at, updated_at FROM roles ORDER BY id`

	rows, err := executor.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}
	defer rows.Close()

	var roles []entity.Role

	for rows.Next() {
		var id int
		var name string
		var description sql.NullString
		var encodedPermission permission.EncodedPermission
		var isDefault bool
		var createdAtStr string
		var updatedAtStr sql.NullString

		err := rows.Scan(
			&id,
			&name,
			&description,
			&encodedPermission,
			&isDefault,
			&createdAtStr,
			&updatedAtStr,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}

		var descriptionValue string
		if description.Valid {
			descriptionValue = description.String
		}

		createdAt, err := util.ParseMySQLDateTime(createdAtStr)
		if err != nil {
			return nil, err
		}

		var updatedAt *time.Time
		if updatedAtStr.Valid {
			parsed, err := util.ParseMySQLDateTime(updatedAtStr.String)
			if err != nil {
				return nil, err
			}
			updatedAt = &parsed
		}

		role := entity.NewRole(id, name, descriptionValue, encodedPermission, isDefault, createdAt, updatedAt)
		roles = append(roles, role)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate roles: %w", err)
	}

	return roles, nil
}
