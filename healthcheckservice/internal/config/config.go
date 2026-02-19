package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	ModeLocal = "local"
	ModeDev   = "dev"
)

type Config struct {
	Env         string      `yaml:"env" env-default:"local"`
	DbConfig    DbConfig    `yaml:"db"`
	KafkaConfig KafkaConfig `yaml:"kafka"`
	HttpConfig  HttpConfig  `yaml:"http"`
	RedisConfig RedisConfig `yaml:"redis"`
	UserConfig  UserConfig  `yaml:"user"`
}

type HttpConfig struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

type DbConfig struct {
	DbUser     string `yaml:"db_user" env:"DATABASE_USER"`
	DbPassword string `yaml:"db_password" env:"DATABASE_PASSWORD"`
	DbHost     string `yaml:"db_host" env:"DATABASE_HOST"`
	DbPort     string `yaml:"db_port" env:"DATABASE_PORT"`
	DbName     string `yaml:"db_name" env:"DATABASE_NAME"`
}

type Consumer struct {
	Topic         string `yaml:"topic"`
	ConsumerGroup string `yaml:"consumer_group"`
}

type RedisConfig struct {
	Password string        `yaml:"password" env:"REDIS_PASSWORD"`
	Host     string        `yaml:"host" env:"REDIS_HOST"`
	Port     string        `yaml:"port" env:"REDIS_PORT"`
	TTL      time.Duration `yaml:"ttl"`
}

type KafkaConfig struct {
	Username string   `yaml:"username" env:"KAFKA_USERNAME"`
	Password string   `yaml:"password" env:"KAFKA_PASSWORD"`
	Brokers  []string `yaml:"brokers"`
	Consumer Consumer `yaml:"consumer"`
}

type UserConfig struct {
	UserPath    string        `yaml:"user_path"`
	UserTimeout time.Duration `yaml:"timeout"`
	UserMethod  string        `yaml:"user_method"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	return MustLoadPath(configPath)
}

func MustLoadPath(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}
	// В Docker env USER_PATH переопределяет user_path из yaml (внутренний URL userservice для validate)
	if v := os.Getenv("USER_PATH"); v != "" {
		cfg.UserConfig.UserPath = v
	}
	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
