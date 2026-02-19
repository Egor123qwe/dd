package rest

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	docs "gitlab.roy9.ru/roy9/backend/core/healthcheckservice/cmd/docs"

	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/config"
	servHandler "gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/handler/server"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/middleware"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/storage"
)

type RestServer interface {
	Start(ctx context.Context)
	Stop(ctx context.Context)
}

type server struct {
	httpServer *http.Server
	log        slog.Logger
}

func New(cfg config.Config, log slog.Logger, storage storage.Storage) RestServer {

	handler := servHandler.New(cfg, log, storage)

	router := configureRouter(handler, cfg, log)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.HttpConfig.Host, cfg.HttpConfig.Port),
		Handler: router,
	}

	return server{
		httpServer: httpServer,
		log:        log,
	}
}

func configureRouter(h servHandler.Handler, cfg config.Config, log slog.Logger) *gin.Engine {
	router := gin.Default()

	docs.SwaggerInfo.BasePath = "/"

	middlerware := middleware.New(cfg, log)

	apiRouters := router.Group("api/v1")
	apiRouters.Use(middlerware.FetchUserMiddleware())

	{
		apiRouters.GET("status/rent/client", h.HandleClientStatus)
		apiRouters.GET("status/session", h.HandleMerchantSessions)
		apiRouters.GET("status/rent/merchant/:session_id", h.HandleMerchantStatus)
		apiRouters.GET("status/rent/client/:session_id", h.HandleClientStatusBySession)
		apiRouters.GET("status/session/:session_id", h.HandleMerchantSession)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	return router
}

func (s server) Start(ctx context.Context) {

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.Error("failed to start HTTP server: ", err.Error(), nil)
		}
	}()
}

func (s server) Stop(ctx context.Context) {
	const op = "server.Shutdown"

	s.log.Info("%s : Shutdown server at %s", op, s.httpServer.Addr)
	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Fatal("%w : Shutting down server: %w", op, err.Error())
	}
}
