package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	docs "gitlab.roy9.ru/roy9/backend/core/historyservice/cmd/docs"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	
	"gitlab.roy9.ru/roy9/backend/core/historyservice/internal/config"
	handler "gitlab.roy9.ru/roy9/backend/core/historyservice/internal/handler"
	"gitlab.roy9.ru/roy9/backend/core/historyservice/internal/middleware"
)

type Server interface {
	Start()
	Stop(ctx context.Context)
	Wait()
}

type server struct {
	wg         *sync.WaitGroup
	httpServer *http.Server
	log        *slog.Logger
}

func NewServer(cfg *config.Config, log *slog.Logger) Server {

	router := configureRouter(cfg, log)

	httpServer := http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.HttpConfig.Host, cfg.HttpConfig.Port),
		Handler: router,
	}

	return &server{
		httpServer: &httpServer,
		log:        log,
		wg:         &sync.WaitGroup{},
	}
}

func configureRouter(cfg *config.Config, log *slog.Logger) *gin.Engine {
	router := gin.Default()

	h := handler.New(cfg, log)

	mid := middleware.NewMiddleware(cfg, log)

	docs.SwaggerInfo.BasePath = "/"

	apiRouters := router.Group("api/v1")
	apiRouters.Use(mid.FetchUserMiddleware())

	{
		apiRouters.GET("/user/history/rent", h.History().Rent)
		apiRouters.GET("/user/history/lease", h.History().Lease)
	}

	// Админ: все аренды (без проверки пользователя — доступ закрыт на стороне gateway/userservice).
	adminRouters := router.Group("api/v1/admin")
	adminRouters.GET("/history/rents", h.History().AdminRents)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	
	return router
}

func (s server) Start() {
	const op = "server.Start"

	s.log.Info("%s : Starting server at %s", op, s.httpServer.Addr)
	s.wg.Add(1)

	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("listen error: ", err.Error())
	}
}

func (s server) Stop(ctx context.Context) {
	const op = "server.Stop"
	
	defer s.wg.Done()
	s.log.Info("%s : Shutdown server at %s", op, s.httpServer.Addr)

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Fatal("%w : Shutting down server: %w", op, err.Error())
	}
}

func (s server) Wait() {
	s.wg.Wait()
}
