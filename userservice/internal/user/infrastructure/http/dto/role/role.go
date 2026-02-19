package role

type CreateRoleReq struct {
	Name        string  `json:"name" binding:"required,min=1,max=50"`
	Description *string `json:"description,omitempty" binding:"omitempty,min=0,max=255"`

	PermissionsNames []string `json:"permissions_names" binding:"required"`
}

type CreateRoleResp struct {
	ID int `json:"id"`
}

type GetRolesResp struct {
	Roles []GetRoleResp `json:"roles"`
}

type GetRoleResp struct {
	ID               int      `json:"id"`
	Name             string   `json:"name"`
	Description      *string  `json:"description"`
	PermissionsNames []string `json:"permissions_names"`
	IsDefault        bool     `json:"is_default"`
	CreatedAt        string   `json:"created_at"`
	UpdatedAt        *string  `json:"updated_at,omitempty"`
}

type UpdateRoleReq struct {
	Name             string   `json:"name" binding:"required,min=1,max=50"`
	Description      *string  `json:"description,omitempty" binding:"omitempty,min=0,max=255"`
	PermissionsNames []string `json:"permissions_names" binding:"required"`
}
