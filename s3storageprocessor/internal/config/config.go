package config

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	GRPC GRPC `mapstucture:"grpc"`
	S3   S3   `mapstructure:"s3"`
}

type GRPC struct {
	Host    string `mapstucture:"host"`
	Port    string `mapstructure:"port"`
	Timeout int64  `mapstructure:"timeout"`
}

type S3 struct {
	Host         string `mapstucture:"host"`
	Port         string `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DefaultQuota int64  `mapstructure:"default_quota"`
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
