package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/viper"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/app"
	"gitlab.roy9.ru/roy9/backend/statemachine/connectioncoordinatorservice/internal/service/util/logger"
)

var (
	// used to disable connection auth and use mock generator of UserID
	debug = false
)

func init() {
	viper.AddConfigPath("config")
	viper.SetConfigName("config")

	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		logFatal(err)
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))
	viper.AutomaticEnv()

	logger.Init()
}

func main() {
	srv, err := app.New(debug)
	if err != nil {
		logFatal(err)
	}

	if err := srv.Start(context.Background()); err != nil {
		logFatal(err)
	}
}

// logFatal used to save container in running state (to change configs)
// because of pipeline
func logFatal(err error) {
	log.Println(err.Error())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Fatal(err)
}
