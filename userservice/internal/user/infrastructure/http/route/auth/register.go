package auth

import (
	"fmt"

	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	dto "github.com/Interpuls/ifc2-service-farm/internal/user/infrastructure/http/dto/auth"
	ginExtend "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/gin_extend"
	respProcessor "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/resp_processor"
	"github.com/gin-gonic/gin"
)

func (h *handler) newRegisterHandler() gin.HandlerFunc {
	return ginExtend.NewHandler(respProcessor.JsonRespSender)(
		func(c *gin.Context) (any, []string, error) {
			var req dto.RegisterReq

			if err := c.ShouldBindJSON(&req); err != nil {
				h.log.Err(err).Msg("failed to parse json request")
				return nil, nil, fmt.Errorf("%w: %w", errs.ErrInvalidRequest, err)
			}

			if err := h.validator.ValidateStruct(req); err != nil {
				return nil, nil, err
			}

			resp, err := h.registerUC.Register(c.Request.Context(), req.Convert())
			if err != nil {
				h.log.Err(err).Str("email", req.Email).Msg("failed to register")
				return nil, nil, err
			}

			return dto.ToRegisterResp(resp), nil, nil
		},
	)
}
