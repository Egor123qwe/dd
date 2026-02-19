package role

import (
	"fmt"
	"strconv"

	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	dto "github.com/Interpuls/ifc2-service-farm/internal/user/infrastructure/http/dto/role"
	ginExtend "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/gin_extend"
	authMiddleware "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/middleware/auth"
	respProcessor "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/resp_processor"
	"github.com/gin-gonic/gin"
)

func (h *handler) newUpdateRoleHandler() gin.HandlerFunc {
	return ginExtend.NewHandler(respProcessor.JsonRespSender)(
		func(c *gin.Context) (any, []string, error) {
			permit := authMiddleware.GetPermitFromContext(c)
			if permit == nil {
				return nil, nil, errs.ErrUnauthorized
			}

			roleIDStr := c.Param("roleId")
			roleID, err := strconv.Atoi(roleIDStr)
			if err != nil {
				return nil, nil, fmt.Errorf("%w: %w", errs.ErrInvalidRequest, err)
			}

			var req dto.UpdateRoleReq

			if err := c.ShouldBindJSON(&req); err != nil {
				return nil, nil, fmt.Errorf("%w: %w", errs.ErrInvalidRequest, err)
			}

			if err := h.validator.ValidateStruct(req); err != nil {
				return nil, nil, err
			}

			err = h.updateRoleUC.UpdateRole(c.Request.Context(), *permit, h.converter.ConvertUpdate(req, roleID))
			if err != nil {
				return nil, nil, err
			}

			return nil, nil, nil
		},
	)
}
