package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/entity"
	"github.com/Interpuls/ifc2-service-farm/pkg/transactor/transaction"
	"github.com/Interpuls/ifc2-service-farm/pkg/util"
)

func (r repo) GetByEmail(ctx context.Context, email string) (entity.User, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	executor := transaction.SelectExecutor(ctx, r.db)

	query := r.db.Rebind(`
		SELECT 
			id, 
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
		FROM users 
		WHERE email = ?`)

	var userDTO entity.UserDTO
	var zipCode, phone sql.NullString
	var languageID sql.NullInt64
	var createdAtStr, updatedAtStr *string

	err := executor.QueryRowxContext(ctx, query, email).Scan(
		&userDTO.ID,
		&userDTO.Username,
		&userDTO.Email,
		&userDTO.HashPassword,
		&userDTO.Name,
		&zipCode,
		&phone,
		&languageID,
		&userDTO.IsPasswordUpdated,
		&createdAtStr,
		&updatedAtStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, fmt.Errorf("%w: user with email %s not found", errs.ErrResourceNotFound, email)
		}
		return entity.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	if zipCode.Valid {
		userDTO.ZipCode = &zipCode.String
	}
	if phone.Valid {
		userDTO.Phone = &phone.String
	}
	if languageID.Valid {
		langID := int(languageID.Int64)
		userDTO.LanguageID = &langID
	}

	if createdAtStr != nil {
		createdAt, err := util.ParseMySQLDateTime(*createdAtStr)
		if err != nil {
			return entity.User{}, err
		}
		userDTO.CreatedAt = &createdAt
	}

	if updatedAtStr != nil {
		updatedAt, err := util.ParseMySQLDateTime(*updatedAtStr)
		if err != nil {
			return entity.User{}, err
		}
		userDTO.UpdatedAt = &updatedAt
	}

	roles, err := r.getUserRoleIDs(ctx, userDTO.ID)
	if err != nil {
		return entity.User{}, fmt.Errorf("failed to get user roles: %w", err)
	}

	userDTO.RoleIDs = roles

	user, err := entity.NewUser(
		userDTO.ID,
		userDTO.RoleIDs,
		userDTO.HashPassword,
		userDTO.Username,
		userDTO.Email,
		userDTO.Name,
		userDTO.ZipCode,
		userDTO.Phone,
		userDTO.LanguageID,
		userDTO.IsPasswordUpdated,
		userDTO.CreatedAt,
		userDTO.UpdatedAt,
	)
	if err != nil {
		return entity.User{}, fmt.Errorf("failed to create user entity: %w", err)
	}

	return user, nil
}
