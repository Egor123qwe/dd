package main

import (
	"bytes"
	"flag"
	"log"
	_ "net/http/pprof"
	"strings"
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

var backendHost string

func init() {
	flag.StringVar(&backendHost, "backend-host", "", "IP/host вместо localhost для piko (server_url, link_template)")
	flag.Parse()

	viper.SetConfigType("yaml")
	if err := viper.ReadConfig(bytes.NewBuffer(config.Data)); err != nil {
		log.Fatal(err)
	}

	if host := strings.TrimSpace(backendHost); host != "" {
		for _, key := range []string{"network.piko.server_url", "network.piko.link_template"} {
			v := viper.GetString(key)
			if v != "" {
				viper.Set(key, strings.Replace(v, "localhost", host, 1))
			}
		}
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

	if backendHost != "" {
		log.Infof("piko backend host: %s (server_url=%s)", backendHost, viper.GetString("network.piko.server_url"))
	} else {
		log.Info("piko backend host: localhost (default)")
	}

	server, err := app.New()
	if err != nil {
		log.Fatal(err, logger.DefaultWithSentry())
	}

	if err := server.Run(); err != nil {
		log.Fatal(err, logger.DefaultWithSentry())
	}
}
