package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.roy9.ru/roy9/backend/core/feedbackservice/internal/domain/user"
	"gitlab.roy9.ru/roy9/backend/core/feedbackservice/internal/handler/response"
)

var (
	ErrValidateToken = errors.New("failed to validate token")
)

type Middleware struct {
	authService AuthService
}

type AuthService interface {
	Validate(ctx context.Context, token string) (user.User, error)
}

func New(authService AuthService) *Middleware {
	middleware := &Middleware{
		authService: authService,
	}

	return middleware
}

func (m *Middleware) Validate() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		usr, err := m.authService.Validate(context.Background(), token)
		if err != nil {
			response.NewError(c, http.StatusUnauthorized, ErrValidateToken.Error())
			c.Abort()

			return
		}

		c.Set("userID", usr.ID)

		c.Next()
	}
}
