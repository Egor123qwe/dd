package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string     `yaml:"env" env-default:"local"`
	DbConfig   DbConfig   `yaml:"db"`
	HttpConfig HttpConfig `yaml:"http"`
	User       User       `yaml:"user"`
}

type User struct {
	UserPath string        `yaml:"user_path" env:"USER_PATH"`
	Timeout  time.Duration `yaml:"timeout"`
	Method   string        `yaml:"user_method"`
}

type HttpConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
	Host    string        `yaml:"host"`
}

type DbConfig struct {
	DbUser     string `yaml:"user" env:"DB_USER"`
	DbPassword string `yaml:"password" env:"DB_PASSWORD"`
	DbHost     string `yaml:"host" env:"DB_HOST"`
	DbPort     string `yaml:"port" env:"DB_PORT"`
	DbName     string `yaml:"name" env:"DB_NAME"`
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
	// В Docker env USER_PATH переопределяет user_path из yaml (внутренний URL userservice)
	if v := os.Getenv("USER_PATH"); v != "" {
		cfg.User.UserPath = v
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
