package main

import (
	"bytes"
	"context"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/config"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/app"
	httpServer "gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/app/server/http"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/internal/service/runtimedaemon"
	exit "gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/context"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/pkg/logger"
)

func init() {
	viper.SetConfigType("yaml")
	if err := viper.ReadConfig(bytes.NewBuffer(config.Data)); err != nil {
		panic(err)
	}
	_ = godotenv.Load() // текущая директория (для запуска из IDE)
	loadEnvFromExecDir() // затем папка с бинарником (приоритет для make build-mac/build-win)
	if host := os.Getenv(config.EnvBackendHost); host != "" {
		host = strings.TrimSpace(host)
		replaceLocalhostInURLs(host)
	}
	logger.Init()
}

// loadEnvFromExecDir загружает .env из папки, где лежит исполняемый файл (для билда из make build-mac/build-win).
func loadEnvFromExecDir() {
	if execPath, err := os.Executable(); err == nil {
		envPath := filepath.Join(filepath.Dir(execPath), ".env")
		_ = godotenv.Load(envPath)
	}
}

// replaceLocalhostInURLs подставляет хост из ROY9_BACKEND_HOST вместо localhost в auth.service_url, status_check.url и state_machine.connection_url.
func replaceLocalhostInURLs(host string) {
	for _, key := range []string{config.AuthServiceURLKey, config.StatusCheckURLKey, config.SmConnectionUrlKey} {
		url := viper.GetString(key)
		if url != "" {
			viper.Set(key, strings.Replace(url, "localhost", host, 1))
		}
	}
}

func main() {
	ctx, cancel := exit.WithSignal(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Логируем используемый backend host (из auth.service_url)
	if u, err := url.Parse(viper.GetString(config.AuthServiceURLKey)); err == nil && u.Host != "" {
		logging.MustGetLogger("merchantclient").Infof("backend host: %s", u.Host)
	} else {
		logging.MustGetLogger("merchantclient").Info("backend host: localhost (default)")
	}

	rd, err := runtimedaemon.New()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		_ = rd.Serve(ctx)
	}()

	// Не ждём runtimedaemon/Docker — страница открывается сразу, статус Docker проверяется в /api/status
	backend := app.NewWebBackend(rd)
	defer backend.Disconnect()

	web := httpServer.New(backend)
	if err := web.Serve(ctx); err != nil && ctx.Err() == nil {
		log.Fatalf("http server: %v", err)
	}
}
