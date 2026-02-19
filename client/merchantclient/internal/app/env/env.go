package env

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/config"
)

var EnvsNotFound = errors.New("envs not found")

const (
	CheatMode          = "CHEAT_MODE"
	AuthTokenKey       = "ROY9_AUTH_TOKEN"
	RdPortKey          = "ROY9_RD_PORT"
	SmConnectionUrlKey = "ROY9_CONNECTION_URL"
	LogLevelKey        = "ROY9_CLIENT_VERBOSE"
)

type Params struct {
	CheatMode *bool

	AuthToken *string

	RdPort          *int
	SmConnectionUrl *string

	LogLevel *string
}

func Parse() (Params, error) {
	params := Params{}
	godotenv.Load()

	if param := os.Getenv(CheatMode); param != "" {
		param = strings.ToLower(strings.Trim(param, " "))

		if param == "true" || param == "false" {
			flag := param == "true"

			params.CheatMode = &flag
		}
	}

	if param := os.Getenv(AuthTokenKey); param != "" {
		params.AuthToken = &param
	}

	if param := os.Getenv(RdPortKey); param != "" {
		port, err := strconv.Atoi(param)
		if err == nil {
			return params, fmt.Errorf("invalid port: %s", param)
		}

		params.RdPort = &port
	}

	if param := os.Getenv(SmConnectionUrlKey); param != "" {
		params.SmConnectionUrl = &param
	}

	if param := os.Getenv(LogLevelKey); param != "" {
		params.LogLevel = &param
	}

	return params, nil
}

func (p Params) RefreshConfig() {
	if p.CheatMode != nil {
		viper.Set(config.CheatModeKey, *p.CheatMode)
	}

	if p.AuthToken != nil {
		viper.Set(config.AuthTokenKey, *p.AuthToken)
	}

	if p.RdPort != nil {
		viper.Set(config.RdPortKey, *p.RdPort)
	}

	if p.SmConnectionUrl != nil {
		viper.Set(config.SmConnectionUrlKey, *p.SmConnectionUrl)
	}

	if p.LogLevel != nil {
		viper.Set(config.LogLevelKey, *p.LogLevel)
	}
}
