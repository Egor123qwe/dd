package permissions

import (
	ginExtend "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/gin_extend"
	respProcessor "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/resp_processor"

	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	dto "github.com/Interpuls/ifc2-service-farm/internal/user/infrastructure/http/dto/permissions"
	authMiddleware "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/middleware/auth"
	"github.com/gin-gonic/gin"
)

func (h handler) newGetMyPermissionListHandler() gin.HandlerFunc {
	return ginExtend.NewHandler(respProcessor.JsonRespSender)(
		func(c *gin.Context) (any, []string, error) {
			permit := authMiddleware.GetPermitFromContext(c)
			if permit == nil {
				return nil, nil, errs.ErrUnauthorized
			}

			list := permit.GetExternalPermissions()

			return dto.ToGetListResp(list), nil, nil
		},
	)
}
