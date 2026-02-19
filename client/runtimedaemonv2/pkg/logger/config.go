package logger

import "github.com/spf13/viper"

type config struct {
	Level string

	ToStderr bool

	ToFile    bool
	Fn        string
	MaxSizeMb int
	MaxFiles  int

	ToSentry  bool
	SentryDSN string
}

func newConfig() config {
	return config{
		Level: viper.GetString("logger.level"),

		ToStderr: viper.GetBool("logger.to_stderr"),

		ToFile:    viper.GetBool("logger.to_file"),
		Fn:        viper.GetString("logger.fn"),
		MaxSizeMb: viper.GetInt("logger.max_size_mb"),
		MaxFiles:  viper.GetInt("logger.max_files"),

		ToSentry:  viper.GetBool("logger.to_sentry"),
		SentryDSN: viper.GetString("logger.sentry_dsn"),
	}
}
