package perm

import (
	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/entity"
	"github.com/Interpuls/ifc2-service-farm/internal/user/domain/vo"
	"github.com/Interpuls/ifc2-service-farm/pkg/permission"
)

func GetPermForUser(user entity.User, roles ...entity.Role) permission.EncodedPermission {
	perms := make([]permission.EncodedPermission, 0, len(roles))

	for _, role := range roles {
		// ignore service role for users without password update (security reasons)
		if !user.ToDTO().IsPasswordUpdated && role.ToDTO().Name == vo.Service.Value() {
			continue
		}

		perms = append(perms, role.ToDTO().EncodedPermission)
	}

	return permission.SumPermissions(perms...)
}
