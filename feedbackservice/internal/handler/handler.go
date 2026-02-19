package handler

import (
	"net/http"

	"gitlab.roy9.ru/roy9/backend/core/feedbackservice/docs"
	"gitlab.roy9.ru/roy9/backend/core/feedbackservice/internal/handler/feedback"
	"gitlab.roy9.ru/roy9/backend/core/feedbackservice/internal/handler/middleware"
	svcfeedback "gitlab.roy9.ru/roy9/backend/core/feedbackservice/internal/service/feedback"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Register(feedbackService *svcfeedback.Service, authMiddleware *middleware.Middleware) http.Handler {
	router := gin.Default()

	docs.SwaggerInfo.BasePath = "/api/v1/feedback"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	feedbackHandler := feedback.New(feedbackService)

	apiGroup := router.Group("api/v1/feedback")
	apiGroup.Use(authMiddleware.Validate())

	apiGroup.POST("/score/rental", feedbackHandler.Create)
	apiGroup.POST("/request/local", feedbackHandler.CreateFeedbackLocal)
	apiGroup.POST("/request/partnership", feedbackHandler.CreateFeedbackPartnership)

	return router
}
