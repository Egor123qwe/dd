package auth

import (
	"fmt"
	"github.com/Interpuls/ifc2-service-farm/pkg/errs"
	ginExtend "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/gin_extend"
	respProcessor "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/resp_processor"
	"github.com/Interpuls/ifc2-service-farm/pkg/jwt"
	"github.com/Interpuls/ifc2-service-farm/pkg/permission"
	"github.com/gin-gonic/gin"
)

type Middleware struct {
	jwt       jwt.JWT
	debugAuth *DebugAuth
}

type DebugAuth struct {
	DebugPermit permission.Permit
}

func New(
	jwt jwt.JWT,
	debugAuth *DebugAuth,
) Middleware {
	m := Middleware{
		jwt:       jwt,
		debugAuth: debugAuth,
	}

	return m
}

func (m Middleware) NewAuthMiddleware() gin.HandlerFunc {
	return ginExtend.NewHandler(respProcessor.MiddlewareProcessor)(
		func(c *gin.Context) (any, []string, error) {
			// for debug
			if m.debugAuth != nil {
				c.Set(PermitCtxKey, m.debugAuth.DebugPermit)
				return nil, nil, nil
			}

			permit, err := m.getPermit(c)
			if err != nil {
				return nil, nil, fmt.Errorf("%w: %w", errs.ErrUnauthorized, err)
			}

			c.Set(PermitCtxKey, permit)

			return nil, nil, nil
		},
	)
}
