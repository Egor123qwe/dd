package user

import (
	"context"
	"fmt"
	"time"

	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/entity"
	"github.com/Interpuls/ifc2-service-farm/pkg/transactor/transaction"
)

func (r repo) Create(ctx context.Context, req entity.User) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	if transaction.FromContext(ctx) == nil {
		tx, err := r.db.BeginTxx(ctx, nil)
		if err != nil {
			return 0, fmt.Errorf("failed to start transaction: %w", err)
		}

		transaction.WrapToContext(ctx, transaction.New(tx))
	}

	executor := transaction.SelectExecutor(ctx, r.db)

	query := r.db.Rebind(`
		INSERT INTO users (
			username, 
			email, 
			hash_password, 
			name, 
			zip_code, 
			phone,
			language_id,
			is_password_updated,
			created_at,
			updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		) RETURNING id`)

	userDTO := req.ToDTO()
	now := time.Now().UTC()

	var id int
	err := executor.QueryRowxContext(ctx, query,
		userDTO.Username,
		userDTO.Email,
		userDTO.HashPassword,
		userDTO.Name,
		userDTO.ZipCode,
		userDTO.Phone,
		userDTO.LanguageID,
		false, // is_password_updated is always false for new users
		now,
		now,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create_user user: %w", err)
	}

	if err := r.createUserRoles(ctx, id, userDTO.RoleIDs); err != nil {
		return 0, err
	}

	return id, nil
}
