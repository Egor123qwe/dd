package config

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server          ServerConfig
	Kafka           KafkaConfig
	Database        DatabaseConfig
	Redis           RedisConfig
	PriceService    PriceServiceConfig
	SessionHandler  SessionHandlerConfig
}

type SessionHandlerConfig struct {
	URL string `mapstructure:"url"` // base URL of sessionhandlerservice, e.g. http://sessionhandlerservice:8096
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type KafkaConfig struct {
	Brokers  []string `mapstructure:"brokers"`
	Username string   `mapstructure:"username"`
	Password string   `mapstructure:"password"`
	Consumer Consumer `mapstructure:"consumer"`
	Producer Producer `mapstructure:"producer"`
}

type Consumer struct {
	Topics        Topics `mapstructure:"topics"`
	ConsumerGroup string `mapstructure:"consumergroup"`
}

type Topics struct {
	Input           string `mapstructure:"input"`
	TTLNotification string `mapstructure:"ttlnotification"`
}

type Producer struct {
	Topic string `mapstructure:"topic"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       string `mapstructure:"db"`
	TTL      int    `mapstructure:"ttl"`
}

type PriceServiceConfig struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
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
