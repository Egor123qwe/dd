package http

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Host         string
	Port         int
	ShutdownTime time.Duration
	ReadTime     time.Duration
}

func NewConfig() Config {
	return Config{
		Host:         viper.GetString("http.host"),
		Port:         viper.GetInt("http.port"),
		ShutdownTime: time.Duration(viper.GetInt("http.shutdown_time")) * time.Second,
		ReadTime:     time.Duration(viper.GetInt("http.read_time")) * time.Second,
	}
}
