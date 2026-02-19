package handler

import (
	"log/slog"
	"net/http"

	docs "gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/docs"
	"gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/config"
	"gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/internal/handler/category"
	"gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/internal/handler/hardware"
	"gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/internal/handler/session"
	"gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/internal/handler/tariff"

	"go.opentelemetry.io/otel/trace"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Router(
	logger slog.Logger,
	cfg config.Config,
	tariffService tariff.TariffService,
	sessionService session.SessionService,
	categoryService category.CategoryService,
	hardwareService hardware.HardwareService,
	tracer trace.Tracer,
) http.Handler {
	router := gin.Default()

	tariffHanlder := tariff.New(logger, tariffService, tracer)
	sessionHandler := session.New(logger, sessionService, tracer)
	categoryHandler := category.New(logger, categoryService, tracer)
	hardwareHandler := hardware.New(logger, hardwareService, tracer)

	docs.SwaggerInfo.BasePath = "/api/v1"

	v1 := router.Group("api/v1")
	{
		v1.GET("exchange/category/tariff", tariffHanlder.List)

		v1.GET("exchange/category/gpu", categoryHandler.GPUDict)

		v1.GET("session/:session_id", sessionHandler.ReceiveByID)
		v1.GET("exchange/category/tariff/:tariff_id/session", sessionHandler.ListByTariffID)
		v1.GET("exchange/category/gpu/:gpu_id/session", sessionHandler.ListByGPUID)

		v1.GET("hardware/gpu/:gpu_id", hardwareHandler.GPUByID)
		v1.GET("hardware/gpu/", hardwareHandler.GPUList)
	}

	router.GET("swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
