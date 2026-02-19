package app

import (
	"context"
	userModule "github.com/Interpuls/ifc2-service-farm/internal/user"

	"github.com/Interpuls/ifc2-service-farm/config"
	"github.com/Interpuls/ifc2-service-farm/pkg/jwt"
	"github.com/Interpuls/ifc2-service-farm/pkg/logger"
	"github.com/Interpuls/ifc2-service-farm/pkg/server"
	httpServer "github.com/Interpuls/ifc2-service-farm/pkg/server/http"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type App struct {
	httpServer server.Server
	grpcServer server.Server

	log zerolog.Logger
}

func New(ctx context.Context, cfg config.Config) (App, error) {
	log := logger.Get()

	router := gin.Default()
	rootRouter := router.Group("/api")

	jwtSrv := jwt.New(cfg.Auth.Secret)

	user, err := userModule.Init(
		cfg.User,
		jwtSrv,
	)
	if err != nil {
		return App{}, err
	}

	user.AssignHttpHandler(rootRouter, jwtSrv)

	app := App{
		httpServer: httpServer.New(router, cfg.Http),
		//grpcServer: grpcServer.New(_, cfg.Grpc),

		log: log,
	}

	return app, nil
}

func (a App) Run(ctx context.Context) error {
	serveGr := server.NewServeGroup(a.httpServer)

	return serveGr.Serve(ctx)
}
