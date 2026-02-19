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
	"github.com/Masterminds/squirrel"
)

func (r repo) GetListByIDs(ctx context.Context, roleIDs ...int) ([]entity.Role, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	executor := transaction.SelectExecutor(ctx, r.db)

	if len(roleIDs) == 0 {
		return nil, nil
	}

	queryBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select("id", "name", "description", "permissions", "is_default", "created_at", "updated_at").
		From("roles").
		Where(squirrel.Eq{"id": roleIDs})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	rows, err := executor.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles by IDs: %w", err)
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

	if len(roles) != len(roleIDs) {
		return nil, fmt.Errorf("failed to get all roles")
	}

	return roles, nil
}
