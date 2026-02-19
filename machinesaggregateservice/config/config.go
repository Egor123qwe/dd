package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `yaml:"env" env-default:"dev"`
	DbConfig   `yaml:"db_config"`
	HTTPServer `yaml:"http_server"`
	GrpcServer GrpcServer   `yaml:"grpc_server"`
	Tracer     TracerConfig `mapstructure:"tracer" yaml:"tracer"`
}

type GrpcServer struct {
	Address string `yaml:"address" env:"GRPC_ADDRESS" env-default:"0.0.0.0:8001"`
}

type DbConfig struct {
	DbUser     string `yaml:"db_user" env:"DATABASE_USER"`
	DbPassword string `yaml:"db_password" env:"DATABASE_PASSWORD"`
	DbHost     string `yaml:"db_host" env:"DATABASE_HOST"`
	DbPort     string `yaml:"db_port" env:"DATABASE_PORT"`
	DbName     string `yaml:"db_name" env:"DATABASE_NAME"`
}

type HTTPServer struct {
	Address     string        `yaml:"address"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
	User        string        `yaml:"user" env-required:"true"`
}

type TracerConfig struct {
	Endpoint    string `mapstructure:"endpoint" yaml:"endpoint"`
	ServiceName string `mapstructure:"service" yaml:"service"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}
	return &cfg
}
