package main

import (
	"bytes"
	"context"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/spf13/viper"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/config"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/app"
	httpServer "gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/app/server/http"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/service/runtimedaemon"
	exit "gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/context"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/logger"
)

const startTimeout = 20 * time.Second

func init() {
	viper.SetConfigType("yaml")
	if err := viper.ReadConfig(bytes.NewBuffer(config.Data)); err != nil {
		panic(err)
	}

	logger.Init()
}

func main() {
	ctx, cancel := exit.WithSignal(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	rd, err := runtimedaemon.New()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		_ = rd.Serve(ctx)
	}()

	rdWaitCtx, rdCancel := context.WithTimeout(ctx, startTimeout)
	defer rdCancel()

	if err := rd.WaitEnabled(rdWaitCtx); err != nil {
		log.Fatalf("runtime daemon not ready: %v", err)
	}

	backend := app.NewWebBackend(rd)
	defer backend.Disconnect()

	web := httpServer.New(backend)
	if err := web.Serve(ctx); err != nil && ctx.Err() == nil {
		log.Fatalf("http server: %v", err)
	}
}
