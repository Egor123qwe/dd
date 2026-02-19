package role

import (
	"fmt"
	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	dto "github.com/Interpuls/ifc2-service-farm/internal/user/infrastructure/http/dto/role"
	ginExtend "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/gin_extend"
	authMiddleware "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/middleware/auth"
	respProcessor "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/resp_processor"
	"github.com/gin-gonic/gin"
)

func (h *handler) newCreateRoleHandler() gin.HandlerFunc {
	return ginExtend.NewHandler(respProcessor.JsonRespSender)(
		func(c *gin.Context) (any, []string, error) {
			permit := authMiddleware.GetPermitFromContext(c)
			if permit == nil {
				return nil, nil, errs.ErrUnauthorized
			}

			var req dto.CreateRoleReq

			if err := c.ShouldBindJSON(&req); err != nil {
				return nil, nil, fmt.Errorf("%w: %w", errs.ErrInvalidRequest, err)
			}

			if err := h.validator.ValidateStruct(req); err != nil {
				return nil, nil, err
			}

			id, err := h.createRoleUC.CreateRole(c.Request.Context(), *permit, h.converter.Convert(req))
			if err != nil {
				return nil, nil, err
			}

			return h.converter.ToCreateResp(id), nil, nil
		},
	)
}
