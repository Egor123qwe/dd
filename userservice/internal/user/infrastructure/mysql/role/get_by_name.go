package role

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/entity"
	"github.com/Interpuls/ifc2-service-farm/pkg/permission"
	"github.com/Interpuls/ifc2-service-farm/pkg/transactor/transaction"
	"github.com/Interpuls/ifc2-service-farm/pkg/util"
)

func (r repo) GetByName(ctx context.Context, name string) (entity.Role, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	executor := transaction.SelectExecutor(ctx, r.db)

	query := r.db.Rebind(`SELECT id, name, description, permissions, is_default, created_at, updated_at FROM roles WHERE name = ?`)

	var id int
	var roleName string
	var description sql.NullString
	var encodedPermission permission.EncodedPermission
	var isDefault bool
	var createdAtStr string
	var updatedAtStr sql.NullString

	err := executor.QueryRowxContext(ctx, query, name).Scan(
		&id,
		&roleName,
		&description,
		&encodedPermission,
		&isDefault,
		&createdAtStr,
		&updatedAtStr,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Role{}, fmt.Errorf("%w: role with name %s not found", errs.ErrResourceNotFound, name)
		}

		return entity.Role{}, fmt.Errorf("failed to get role: %w", err)
	}

	var descriptionValue string
	if description.Valid {
		descriptionValue = description.String
	}

	createdAt, err := util.ParseMySQLDateTime(createdAtStr)
	if err != nil {
		return entity.Role{}, err
	}

	var updatedAt *time.Time
	if updatedAtStr.Valid {
		parsed, err := util.ParseMySQLDateTime(updatedAtStr.String)
		if err != nil {
			return entity.Role{}, err
		}
		updatedAt = &parsed
	}

	role := entity.NewRole(id, roleName, descriptionValue, encodedPermission, isDefault, createdAt, updatedAt)

	return role, nil
}
