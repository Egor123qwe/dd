package role

import (
	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/entity"
	createRoleUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/role/create_role"
	updateRoleUsecase "github.com/Interpuls/ifc2-service-farm/internal/user/usecase/role/update_role"
	"github.com/Interpuls/ifc2-service-farm/pkg/permission"
	"github.com/Interpuls/ifc2-service-farm/pkg/util"
)

type Converter struct {
}

func NewConverter() Converter {
	return Converter{}
}

func (c Converter) ToCreateResp(roleID int) CreateRoleResp {
	return CreateRoleResp{
		ID: roleID,
	}
}

func (c Converter) ToGetListResp(requesterID int, roles []entity.Role) (GetRolesResp, error) {

	response := GetRolesResp{
		Roles: make([]GetRoleResp, 0, len(roles)),
	}

	for _, role := range roles {
		roleDTO := role.ToDTO()

		permissions := permission.DecodePermissions(roleDTO.EncodedPermission)

		getRoleResp := GetRoleResp{
			ID:               roleDTO.ID,
			Name:             roleDTO.Name,
			Description:      roleDTO.Description,
			PermissionsNames: make([]string, 0, len(permissions)),
			IsDefault:        roleDTO.IsDefault,
		}

		for _, p := range permissions {
			getRoleResp.PermissionsNames = append(getRoleResp.PermissionsNames, p.String())
		}

		getRoleResp.CreatedAt = roleDTO.CreatedAt.String()

		if roleDTO.UpdatedAt != nil {
			getRoleResp.UpdatedAt = util.Ptr(roleDTO.UpdatedAt.String())
		}

		response.Roles = append(response.Roles, getRoleResp)
	}

	return response, nil
}

func (c Converter) Convert(req CreateRoleReq) createRoleUsecase.Req {
	var description string
	if req.Description != nil {
		description = *req.Description
	}

	return createRoleUsecase.Req{
		Name:             req.Name,
		Description:      description,
		PermissionsNames: req.PermissionsNames,
	}
}

func (c Converter) ConvertUpdate(req UpdateRoleReq, roleID int) updateRoleUsecase.Req {
	var description string
	if req.Description != nil {
		description = *req.Description
	}

	return updateRoleUsecase.Req{
		ID:               roleID,
		Name:             req.Name,
		Description:      description,
		PermissionsNames: req.PermissionsNames,
	}
}
