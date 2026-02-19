package flag

import (
	"flag"

	"github.com/spf13/viper"
	"gitlab.roy9.ru/roy9/backend/clientside/merchantclient/config"
)

const (
	cheatModeFlag = "cheat-mode"

	tokenFlag = "token"

	rdPortFlag          = "rd-port"
	smConnectionUrlFlag = "connection-url"

	logLevelFlag = "verbose"
)

type Flags struct {
	CheatMode *bool

	TokenFlag *string

	RDPort *int
	SmURL  *string

	LogLevel *string
}

func Parse() Flags {
	flags := Flags{
		CheatMode: flag.Bool(
			cheatModeFlag,
			viper.GetBool(config.CheatModeKey),
			"cheat-mode",
		),

		TokenFlag: flag.String(
			tokenFlag,
			viper.GetString(config.AuthTokenKey),
			"example token",
		),

		RDPort: flag.Int(
			rdPortFlag,
			viper.GetInt(config.RdPortKey),
			"runtimedaemon port",
		),

		SmURL: flag.String(
			smConnectionUrlFlag,
			viper.GetString(config.SmConnectionUrlKey),
			"state machine url",
		),

		LogLevel: flag.String(
			logLevelFlag,
			viper.GetString(config.LogLevelKey),
			"log level. Allowed: [DEBUG, INFO, NOTICE, WARNING, ERROR, CRITICAL]",
		),
	}

	flag.Parse()

	return flags
}

func (f Flags) RefreshConfig() {
	viper.Set(config.CheatModeKey, *f.CheatMode)
	viper.Set(config.AuthTokenKey, *f.TokenFlag)
	viper.Set(config.RdPortKey, f.RDPort)
	viper.Set(config.SmConnectionUrlKey, f.SmURL)
	viper.Set(config.LogLevelKey, f.LogLevel)
}
