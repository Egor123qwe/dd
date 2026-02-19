package config

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	HTTP     HTTPConfig     `mapstructure:"http"`
	Database DatabaseConfig `mapstructure:"database"`
	Api      ApiConfig      `mapstructure:"api"`
	Tracer   TracerConfig   `mapstructure:"tracer"`
}

type HTTPConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

type ApiConfig struct {
	AuthConfig AuthConfig `mapstructure:"auth"`
}

type AuthConfig struct {
	URL string `mapstructure:"url"`
}

type TracerConfig struct {
	Endpoint    string `mapstructure:"endpoint"`
	ServiceName string `mapstructure:"service"`
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
