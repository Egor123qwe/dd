package role

import (
	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	ginExtend "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/gin_extend"
	authMiddleware "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/middleware/auth"
	respProcessor "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/resp_processor"
	"github.com/gin-gonic/gin"
)

func (h *handler) newGetRoleListHandler() gin.HandlerFunc {
	return ginExtend.NewHandler(respProcessor.JsonRespSender)(
		func(c *gin.Context) (any, []string, error) {
			permit := authMiddleware.GetPermitFromContext(c)
			if permit == nil {
				return nil, nil, errs.ErrUnauthorized
			}

			roles, err := h.getRoleListUC.GetRoleList(c.Request.Context(), *permit)
			if err != nil {
				return nil, nil, err
			}

			resp, err := h.converter.ToGetListResp(permit.GetActorId(), roles)
			if err != nil {
				return nil, nil, err
			}
			return resp, nil, nil
		},
	)
}
