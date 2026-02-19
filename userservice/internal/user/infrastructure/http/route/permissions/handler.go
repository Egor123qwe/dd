package permissions

import (
	authMiddleware "github.com/Interpuls/ifc2-service-farm/pkg/handler/http/middleware/auth"
	"github.com/Interpuls/ifc2-service-farm/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type handler struct {
	log zerolog.Logger

	middleware authMiddleware.Middleware
}

func New(
	router *gin.RouterGroup,
	middleware authMiddleware.Middleware,
) {
	h := handler{
		log: logger.Named("PermissionsHandler"),

		middleware: middleware,
	}

	router.GET("/", h.newGetPermissionListHandler())

	meGroup := router.Group("/me")
	meGroup.Use(h.middleware.NewAuthMiddleware())

	meGroup.GET("/", h.newGetMyPermissionListHandler())
}
