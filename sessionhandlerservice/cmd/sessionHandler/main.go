package main

import (
	"context"
	"log"
	"strings"

	"github.com/spf13/viper"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/app"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/util/logger"
)

func init() {
	viper.AddConfigPath("config")
	viper.SetConfigName("config")

	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))
	viper.AutomaticEnv()

	logger.Init()
}

func main() {
	srv, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := srv.Start(context.Background()); err != nil {
		log.Fatal(err)
	}
}
