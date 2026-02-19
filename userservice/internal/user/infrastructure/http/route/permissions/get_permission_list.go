package permissions

import (
	dto "github.com/Interpuls/ifc2-service-farm/internal/user/infrastructure/http/dto/permissions"
	ginExtend "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/gin_extend"
	respProcessor "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/resp_processor"
	"github.com/Interpuls/ifc2-service-farm/pkg/permission"
	"github.com/gin-gonic/gin"
)

func (h handler) newGetPermissionListHandler() gin.HandlerFunc {
	return ginExtend.NewHandler(respProcessor.JsonRespSender)(
		func(c *gin.Context) (any, []string, error) {
			list := permission.GetListOfExternal()

			return dto.ToGetListResp(list), nil, nil
		},
	)
}
