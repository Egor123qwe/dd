package user

import (
	"strconv"

	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	ginExtend "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/gin_extend"
	authMiddleware "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/middleware/auth"
	respProcessor "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/resp_processor"
	"github.com/gin-gonic/gin"
)

func (h *handler) newGetUserHandler() gin.HandlerFunc {
	return ginExtend.NewHandler(respProcessor.JsonRespSender)(
		func(c *gin.Context) (any, []string, error) {
			permit := authMiddleware.GetPermitFromContext(c)
			if permit == nil {
				return nil, nil, errs.ErrUnauthorized
			}

			userIDStr := c.Param("userId")

			id, err := strconv.Atoi(userIDStr)
			if err != nil {
				return nil, nil, err
			}

			user, err := h.getUserUC.GetUser(c.Request.Context(), *permit, id)
			if err != nil {
				return nil, nil, err
			}

			resp, err := h.converter.ToGetResp(permit.GetActorId(), user)
			if err != nil {
				return nil, nil, err
			}

			return resp, nil, nil
		},
	)
}
