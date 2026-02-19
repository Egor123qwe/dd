package user

import (
	"context"
	"fmt"
	"time"

	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/entity"
	"github.com/Interpuls/ifc2-service-farm/pkg/transactor/transaction"
)

func (r repo) Update(ctx context.Context, req entity.User) error {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	if transaction.FromContext(ctx) == nil {
		tx, err := r.db.BeginTxx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		transaction.WrapToContext(ctx, transaction.New(tx))
	}

	executor := transaction.SelectExecutor(ctx, r.db)

	query := r.db.Rebind(`
		UPDATE users SET 
			username = ?,
			email = ?,
			hash_password = ?,
			name = ?,
			zip_code = ?,
			phone = ?,
			language_id = ?,
			is_password_updated = ?,
			updated_at = ?
		WHERE id = ?`)

	userDTO := req.ToDTO()
	now := time.Now().UTC()

	_, err := executor.ExecContext(ctx, query,
		userDTO.Username,
		userDTO.Email,
		userDTO.HashPassword,
		userDTO.Name,
		userDTO.ZipCode,
		userDTO.Phone,
		userDTO.LanguageID,
		userDTO.IsPasswordUpdated,
		now,
		userDTO.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if err := r.deleteUserRoles(ctx, userDTO.ID); err != nil {
		return err
	}

	if err := r.createUserRoles(ctx, userDTO.ID, userDTO.RoleIDs); err != nil {
		return err
	}

	return nil
}
