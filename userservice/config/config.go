package config

import (
	_ "embed"

	"os"
	"time"

	"github.com/Interpuls/ifc2-service-farm/pkg/logger"
	"github.com/Interpuls/ifc2-service-farm/pkg/server/grpc"
	"github.com/Interpuls/ifc2-service-farm/pkg/server/http"
	"github.com/spf13/viper"
)

var (
	AppVersion string
)

//go:embed config_local_dev.yaml
var LocalDevData []byte

//go:embed config_docker.yaml
var DockerData []byte

func GetConfigData() []byte {
	if os.Getenv("IS_COMPOSE_CONTEXT") == "1" {
		return DockerData
	}
	return LocalDevData
}

type Config struct {
	Http http.Config
	Grpc grpc.Config

	Auth AuthConfig

	Logger logger.Config

	User UserConfig
}

type UserConfig struct {
	Postgres     PostgresConfig
	Token        TokenConfig
	Registration RegistrationConfig
}

type RegistrationConfig struct {
	DefaultRoleID int
}

type PostgresConfig struct {
	URL            string
	MigrationsDir  string
	TestSeedersDir string
	RunTestSeeders bool
}

type AuthConfig struct {
	Secret string
}

type TokenConfig struct {
	AtExp time.Duration
	RtExp time.Duration
}

func New() Config {
	AppVersion = viper.GetString("version")

	return Config{
		Http: NewHttp(),
		Grpc: NewGrpc(),

		Auth: NewAuth(),

		Logger: NewLogger(),

		User: NewUser(),
	}
}

func NewHttp() http.Config {
	return http.Config{
		Port: viper.GetInt("http.port"),

		ShutdownTime: viper.GetDuration("http.shutdown_time"),
		ReadTime:     viper.GetDuration("http.read_time"),
	}
}

func NewGrpc() grpc.Config {
	return grpc.Config{
		Port: viper.GetInt("grpc.port"),
	}
}

func NewUser() UserConfig {
	return UserConfig{
		Postgres: PostgresConfig{
			URL:            viper.GetString("user.postgres.url"),
			MigrationsDir:  viper.GetString("user.postgres.migrations_dir"),
			TestSeedersDir: viper.GetString("user.postgres.test_seeders_dir"),
			RunTestSeeders: viper.GetBool("user.postgres.run_test_seeders"),
		},

		Token: TokenConfig{
			AtExp: viper.GetDuration("user.token.at_exp"),
			RtExp: viper.GetDuration("user.token.rt_exp"),
		},

		Registration: RegistrationConfig{
			DefaultRoleID: viper.GetInt("user.registration.default_role_id"),
		},
	}
}

func NewAuth() AuthConfig {
	return AuthConfig{
		Secret: viper.GetString("auth.secret"),
	}
}

func NewLogger() logger.Config {
	return logger.Config{
		Dir:               viper.GetString("logger.dir"),
		Filename:          viper.GetString("logger.filename"),
		Level:             viper.GetString("logger.level"),
		MaxSizeMB:         viper.GetInt("logger.max_size_mb"),
		MaxBackups:        viper.GetInt("logger.max_backups"),
		MaxAgeDays:        viper.GetInt("logger.max_age_days"),
		Compress:          viper.GetBool("logger.compress"),
		DuplicateToStdout: viper.GetBool("logger.duplicate_to_stdout"),
		TimeFormat:        viper.GetString("logger.time_format"),
		ServiceName:       viper.GetString("instance.name"),
	}
}
