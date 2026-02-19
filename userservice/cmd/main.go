package main

import (
	"bytes"
	"context"
	"github.com/Interpuls/ifc2-service-farm/config"
	"github.com/Interpuls/ifc2-service-farm/internal/app"
	"github.com/joho/godotenv"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Interpuls/ifc2-service-farm/pkg/logger"
	"github.com/spf13/viper"
)

func init() {
	_ = godotenv.Load()

	viper.SetConfigType("yaml")
	if err := viper.ReadConfig(bytes.NewBuffer(config.GetConfigData())); err != nil {
		panic(err)
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))
	// Set env prefix to avoid conflicts with system variables like USER
	viper.SetEnvPrefix("IFC2")
	viper.AutomaticEnv()

	if err := logger.Init(config.NewLogger()); err != nil {
		panic(err)
	}
}

func main() {
	log := logger.Get()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	srv, err := app.New(ctx, config.New())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init app")
	}

	log.Info().Msg("app initialized successfully")

	ctx = logger.WrapToContext(ctx, logger.Get())

	if err := srv.Run(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed in app running")
	}

	log.Info().Msg("app stopped")
}
