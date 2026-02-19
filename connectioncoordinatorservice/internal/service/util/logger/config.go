package logger

import "github.com/spf13/viper"

type config struct {
	Level     string
	ToFile    bool
	ToStderr  bool
	Fn        string
	MaxSizeMb int
	MaxFiles  int
}

func newConfig() config {
	return config{
		Level:     viper.GetString("logger.level"),
		ToFile:    viper.GetBool("logger.to_file"),
		ToStderr:  viper.GetBool("logger.to_stderr"),
		Fn:        viper.GetString("logger.fn"),
		MaxSizeMb: viper.GetInt("logger.max_size_mb"),
		MaxFiles:  viper.GetInt("logger.max_files"),
	}
}
