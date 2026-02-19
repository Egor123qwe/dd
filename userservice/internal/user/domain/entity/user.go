package entity

import (
	"time"

	"github.com/Interpuls/ifc2-service-farm/pkg/util"
)

type User struct {
	id                int
	roleIDs           []int
	hashPassword      string
	username          string
	email             string
	name              string
	zipCode           *string
	phone             *string
	languageID        *int
	isPasswordUpdated bool
	createdAt         *time.Time
	updatedAt         *time.Time
}

func NewUser(
	id int,
	roleIDs []int,
	hashPassword string,
	username string,
	email string,
	name string,
	zipCode *string,
	phone *string,
	languageID *int,
	isPasswordUpdated bool,
	createdAt *time.Time,
	updatedAt *time.Time,
) (User, error) {
	entity := User{
		id:                id,
		roleIDs:           roleIDs,
		hashPassword:      hashPassword,
		username:          username,
		email:             email,
		name:              name,
		zipCode:           zipCode,
		phone:             phone,
		languageID:        languageID,
		isPasswordUpdated: isPasswordUpdated,
		createdAt:         createdAt,
		updatedAt:         updatedAt,
	}

	return entity, nil
}

func (u *User) ApplyUpdate(
	roleIDs *[]int,
	hashPassword *string,
	username *string,
	email *string,
	name *string,
	zipCode **string,
	phone **string,
	languageID **int,
	isPasswordUpdated *bool,
) error {
	util.UpdateIfNotNil(&u.roleIDs, roleIDs)
	util.UpdateIfNotNil(&u.hashPassword, hashPassword)
	util.UpdateIfNotNil(&u.username, username)
	util.UpdateIfNotNil(&u.email, email)
	util.UpdateIfNotNil(&u.name, name)
	util.UpdateIfNotNil(&u.zipCode, zipCode)
	util.UpdateIfNotNil(&u.phone, phone)
	util.UpdateIfNotNil(&u.languageID, languageID)
	util.UpdateIfNotNil(&u.isPasswordUpdated, isPasswordUpdated)

	return nil
}

type UserDTO struct {
	ID                int
	RoleIDs           []int
	HashPassword      string
	Username          string
	Email             string
	Name              string
	ZipCode           *string
	Phone             *string
	LanguageID        *int
	IsPasswordUpdated bool
	CreatedAt         *time.Time
	UpdatedAt         *time.Time
}

func (u *User) ToDTO() UserDTO {
	dto := UserDTO{
		ID:                u.id,
		RoleIDs:           u.roleIDs,
		HashPassword:      u.hashPassword,
		Username:          u.username,
		Email:             u.email,
		Name:              u.name,
		ZipCode:           u.zipCode,
		Phone:             u.phone,
		LanguageID:        u.languageID,
		IsPasswordUpdated: u.isPasswordUpdated,
		CreatedAt:         u.createdAt,
		UpdatedAt:         u.updatedAt,
	}

	return dto
}
