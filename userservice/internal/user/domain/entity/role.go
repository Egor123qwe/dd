package entity

import (
	"time"

	"github.com/Interpuls/ifc2-service-farm/pkg/permission"
)

type Role struct {
	id          int
	name        string
	description string

	encodedPermission permission.EncodedPermission
	isDefault         bool
	createdAt         time.Time
	updatedAt         *time.Time
}

type RoleDTO struct {
	ID          int
	Name        string
	Description *string

	EncodedPermission permission.EncodedPermission
	IsDefault         bool
	CreatedAt         time.Time
	UpdatedAt         *time.Time
}

func NewRole(
	id int,
	name string,
	description string,
	encodedPermission permission.EncodedPermission,
	isDefault bool,
	createdAt time.Time,
	updatedAt *time.Time,
) Role {
	entity := Role{
		id:                id,
		name:              name,
		description:       description,
		encodedPermission: encodedPermission,
		isDefault:         isDefault,
		createdAt:         createdAt,
		updatedAt:         updatedAt,
	}

	return entity
}

func (r *Role) ToDTO() RoleDTO {
	var description *string
	if r.description != "" {
		description = &r.description
	}

	dto := RoleDTO{
		ID:                r.id,
		Name:              r.name,
		Description:       description,
		EncodedPermission: r.encodedPermission,
		IsDefault:         r.isDefault,
		CreatedAt:         r.createdAt,
		UpdatedAt:         r.updatedAt,
	}

	return dto
}
