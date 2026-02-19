package user

import (
	jsonField "github.com/Interpuls/ifc2-service-farm/pkg/json_field"
)

type CreateReq struct {
	Username string  `json:"username" validate:"required,min=3,max=50,alphanum"`
	RoleIDs  []int   `json:"role_ids" validate:"required"`
	Email    string  `json:"email" validate:"required,email,max=254"`
	Password string  `json:"password" validate:"required,min=8,password"`
	Name     string  `json:"name" validate:"required,min=1,max=100"`
	ZipCode  *string `json:"zip_code,omitempty" validate:"omitempty,max=20"`
	Phone    *string `json:"phone,omitempty" validate:"omitempty,max=20"`
}

type CreateResp struct {
	ID int `json:"id"`
}

type GetResp struct {
	ID                            int     `json:"id"`
	RoleIDs                       []int   `json:"role_ids"`
	Username                      string  `json:"username"`
	Email                         string  `json:"email"`
	Name                          string  `json:"name"`
	ZipCode                       *string `json:"zip_code,omitempty"`
	Phone                         *string `json:"phone,omitempty"`
	LanguageID     *int    `json:"language_id"`
	CreatedAt      *string `json:"created_at,omitempty"`
	UpdatedAt      *string `json:"updated_at,omitempty"`
}

type UpdateProfileReq struct {
	RoleIDs                       *[]int           `json:"role_ids,omitempty"`
	Username                      *string          `json:"username,omitempty" validate:"omitempty,min=3,max=50,alphanum"`
	Email                         *string          `json:"email,omitempty" validate:"omitempty,email,max=254"`
	Name                          *string          `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	ZipCode                       jsonField.String `json:"zip_code" validate:"omitempty,max=20"`
	Phone                         jsonField.String `json:"phone" validate:"omitempty,max=20"`
	LanguageID jsonField.Int `json:"language_id" validate:"omitempty"`
}

type UpdatePasswordReq struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,password"`
}

type GetListResp struct {
	Users []GetResp `json:"users"`
}

type DeleteReq struct {
	ID int `json:"id" validate:"required"`
}
