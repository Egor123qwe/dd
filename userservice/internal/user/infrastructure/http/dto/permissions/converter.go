package permissions

import "github.com/Interpuls/ifc2-service-farm/pkg/permission"

func ToGetListResp(perms []permission.Permission) GetListResp {
	resp := make(GetListResp, 0, len(perms))

	for _, perm := range perms {
		resp = append(resp, perm.String())
	}

	return resp
}
