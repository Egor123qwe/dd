package config

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Kafka KafkaConfig
	Redis RedisConfig
}

type KafkaConfig struct {
	Brokers  []string `mapstructure:"brokers"`
	Username string   `mapstructure:"username"`
	Password string   `mapstructure:"password"`
	Producer Producer `mapstructure:"producer"`
	Consumer Consumer `mapstructure:"consumer"`
}

type Producer struct {
	Topic string `mapstructure:"topic"`
}

type Consumer struct {
	Topic   string `mapstructure:"topic"`
	GroupID string `mapstructure:"group"`
}

type RedisConfig struct {
	Host      string `mapstructure:"host"`
	Port      string `mapstructure:"port"`
	Password  string `mapstructure:"password"`
	DBNumbers []int  `mapstructure:"dbnumbers"`
	TTL       int    `mapstructure:"ttl"`
}

func LoadConfig() (Config, error) {
	var config *Config

	_, b, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(b), "../..")
	viper.AddConfigPath(root)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		return Config{}, err
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return Config{}, err
	}

	return *config, nil
}
