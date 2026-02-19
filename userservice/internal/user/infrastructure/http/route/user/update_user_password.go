package user

import (
	"fmt"
	"strconv"

	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	dto "github.com/Interpuls/ifc2-service-farm/internal/user/infrastructure/http/dto/user"
	ginExtend "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/gin_extend"
	authMiddleware "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/middleware/auth"
	respProcessor "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/resp_processor"
	"github.com/gin-gonic/gin"
)

func (h *handler) newUpdateUserPasswordHandler() gin.HandlerFunc {
	return ginExtend.NewHandler(respProcessor.JsonRespSender)(
		func(c *gin.Context) (any, []string, error) {
			permit := authMiddleware.GetPermitFromContext(c)
			if permit == nil {
				return nil, nil, errs.ErrUnauthorized
			}

			userIDStr := c.Param("userId")
			userID, err := strconv.Atoi(userIDStr)
			if err != nil {
				h.log.Err(err).Str("userId", userIDStr).Msg("failed to parse user id")
				return nil, nil, fmt.Errorf("%w: %w", errs.ErrInvalidRequest, err)
			}

			var req dto.UpdatePasswordReq

			if err := c.ShouldBindJSON(&req); err != nil {
				h.log.Err(err).Msg("failed to parse json request")
				return nil, nil, fmt.Errorf("%w: %w", errs.ErrInvalidRequest, err)
			}

			if err := h.validator.ValidateStruct(req); err != nil {
				return nil, nil, err
			}

			err = h.updateUserPasswordUC.UpdateUserPassword(c.Request.Context(), *permit, req.Convert(userID))
			if err != nil {
				h.log.Err(err).Int("user_id", userID).Msg("failed to update user password")
				return nil, nil, err
			}

			return nil, nil, nil
		},
	)
}
