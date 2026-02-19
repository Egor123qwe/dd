package main

import (
	"bytes"
	"log"
	_ "net/http/pprof"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/spf13/viper"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/config"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/app"
	loggeropts "gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/service/util/logger"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/pkg/logger"
)

const (
	sentryFlashTimeout = 15 * time.Second
)

func init() {
	viper.SetConfigType("yaml")
	if err := viper.ReadConfig(bytes.NewBuffer(config.Data)); err != nil {
		log.Fatal(err)
	}

	log.SetOutput(&bytes.Buffer{})

	logOpts, err := loggeropts.NewLoggerOpts()
	if err != nil {
		log.Fatal(err)
	}

	if err := logger.Init(logOpts); err != nil {
		log.Fatal(err)
	}
}

func main() {
	log := logger.NewLogger("main", logger.DefaultWithSentry())

	defer sentry.Flush(sentryFlashTimeout)

	server, err := app.New()
	if err != nil {
		log.Fatal(err, logger.DefaultWithSentry())
	}

	if err := server.Run(); err != nil {
		log.Fatal(err, logger.DefaultWithSentry())
	}
}
