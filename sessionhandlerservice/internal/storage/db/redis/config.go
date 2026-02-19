package redis

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	URL      string
	Password string
}

func NewConfig() Config {
	host := viper.GetString("db.redis.host")
	port := viper.GetString("db.redis.port")

	return Config{
		URL:      fmt.Sprintf("%s:%s", host, port),
		Password: viper.GetString("db.redis.password"),
	}
}
