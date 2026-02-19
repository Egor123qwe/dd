package user

import (
	"strconv"

	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	ginExtend "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/gin_extend"
	authMiddleware "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/middleware/auth"
	respProcessor "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/resp_processor"
	"github.com/gin-gonic/gin"
)

func (h *handler) newDeleteUserHandler() gin.HandlerFunc {
	return ginExtend.NewHandler(respProcessor.JsonRespSender)(
		func(c *gin.Context) (any, []string, error) {
			permit := authMiddleware.GetPermitFromContext(c)
			if permit == nil {
				return nil, nil, errs.ErrUnauthorized
			}

			userIDStr := c.Param("userId")

			id, err := strconv.Atoi(userIDStr)
			if err != nil {
				h.log.Err(err).Str("userId", userIDStr).Msg("failed to parse user id")
				return nil, nil, err
			}

			err = h.deleteUserUC.DeleteUser(c.Request.Context(), *permit, id)
			if err != nil {
				h.log.Err(err).Int("user_id", id).Msg("failed to delete user")
				return nil, nil, err
			}

			return nil, nil, nil
		},
	)
}
