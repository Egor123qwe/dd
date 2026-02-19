package middleware

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.roy9.ru/roy9/backend/core/historyservice/internal/config"
	"gitlab.roy9.ru/roy9/backend/core/historyservice/internal/domain/model"
)

var (
	ErrUnauthorized   = errors.New("unauthorized")
	ErrInternalServer = errors.New("internal server error")
	ErrBadToken       = errors.New("bad token")
)

type Middleware struct {
	httpClient *http.Client
	cfg        *config.Config
	log        *slog.Logger
}

func NewMiddleware(cfg *config.Config, log *slog.Logger) *Middleware {
	httpClient := &http.Client{
		Timeout: cfg.User.Timeout,
	}

	return &Middleware{
		httpClient: httpClient,
		cfg:        cfg,
		log:        log,
	}
}

func (m *Middleware) FetchUserMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var user model.User
		token := ctx.GetHeader("Authorization")
		if token == "" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": ErrBadToken.Error()})
			ctx.Abort()

			return
		}

		req, err := http.NewRequest(http.MethodPost, m.cfg.User.UserPath, nil)
		if err != nil {
			m.log.Error("fetch user request failed", err.Error(), nil)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServer.Error()})
			ctx.Abort()

			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", token)

		res, err := m.httpClient.Do(req)
		if err != nil {
			m.log.Error("Unauthorized", err.Error(), nil)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServer.Error()})
			ctx.Abort()

			return
		}

		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": ErrBadToken.Error()})
			ctx.Abort()

			return
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServer.Error()})
			ctx.Abort()

			return
		}

		if err := json.Unmarshal(body, &user); err != nil {
			m.log.Error("failed to unmarshal response", err.Error(), nil)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServer.Error()})
			ctx.Abort()

			return
		}

		m.log.Info("authenticated successfully")
		ctx.Set("UserID", user.UserID)

		ctx.Next()
	}
}
