package user

import (
	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	ginExtend "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/gin_extend"
	authMiddleware "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/middleware/auth"
	respProcessor "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/resp_processor"
	"github.com/gin-gonic/gin"
)

func (h *handler) newGetUserListHandler() gin.HandlerFunc {
	return ginExtend.NewHandler(respProcessor.JsonRespSender)(
		func(c *gin.Context) (any, []string, error) {
			permit := authMiddleware.GetPermitFromContext(c)
			if permit == nil {
				return nil, nil, errs.ErrUnauthorized
			}

			users, err := h.getUserListUC.GetUserList(c.Request.Context(), *permit)
			if err != nil {

				return nil, nil, err
			}

			resp, err := h.converter.ToGetListResp(permit.GetActorId(), users)
			if err != nil {
				return nil, nil, err
			}
			
			return resp, nil, nil
		},
	)
}
