package auth

import (
	"fmt"
	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	dto "github.com/Interpuls/ifc2-service-farm/internal/user/infrastructure/http/dto/auth"
	ginExtend "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/gin_extend"
	respProcessor "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/resp_processor"
	"github.com/gin-gonic/gin"
)

func (h *handler) newRefreshTokenHandler() gin.HandlerFunc {
	return ginExtend.NewHandler(respProcessor.JsonRespSender)(
		func(c *gin.Context) (any, []string, error) {
			var req dto.RefreshReq

			if err := c.ShouldBindJSON(&req); err != nil {
				return nil, nil, fmt.Errorf("%w: %w", errs.ErrInvalidRequest, err)
			}

			if err := h.validator.ValidateStruct(req); err != nil {
				return nil, nil, err
			}

			resp, err := h.refreshTokenUC.RefreshToken(c.Request.Context(), req.Convert())
			if err != nil {
				return nil, nil, err
			}

			return dto.ToRefreshResp(resp), nil, nil
		},
	)
}
