package grpc

import (
	"github.com/spf13/viper"
)

type Config struct {
	Port int
}

func NewConfig() Config {
	return Config{
		Port: viper.GetInt("grpc_port"),
	}
}
