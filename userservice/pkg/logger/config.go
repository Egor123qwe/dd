package logger

import (
	"time"

	"github.com/rs/zerolog"
)

const (
	DefaultDir               = "/app/data/logs"
	DefaultFilename          = "app.log"
	DefaultServiceName       = "main"
	DefaultTimeFormat        = time.RFC3339
	DefaultMaxSizeMB         = 50
	DefaultMaxBackups        = 10
	DefaultMaxAgeDays        = 30
	DefaultCompress          = true
	DefaultDuplicateToStdout = true
)

const (
	LevelDebug = "DEBUG"
	LevelInfo  = "INFO"
	LevelWarn  = "WARNING"
	LevelError = "ERROR"
	LevelFatal = "FATAL"
	LevelPanic = "PANIC"
)

type Config struct {
	Dir               string
	Filename          string
	Level             string
	MaxSizeMB         int
	MaxBackups        int
	MaxAgeDays        int
	Compress          bool
	DuplicateToStdout bool
	TimeFormat        string
	ServiceName       string
}

func validateConfig(cfg Config) Config {
	if cfg.Dir == "" {
		cfg.Dir = DefaultDir
	}

	if cfg.Filename == "" {
		cfg.Filename = DefaultFilename
	}

	if cfg.TimeFormat == "" {
		cfg.TimeFormat = DefaultTimeFormat
	}

	if cfg.MaxSizeMB == 0 {
		cfg.MaxSizeMB = DefaultMaxSizeMB
	}

	if cfg.MaxBackups == 0 {
		cfg.MaxBackups = DefaultMaxBackups
	}

	if cfg.MaxAgeDays == 0 {
		cfg.MaxAgeDays = DefaultMaxAgeDays
	}

	if cfg.ServiceName == "" {
		cfg.ServiceName = DefaultServiceName
	}

	return cfg
}

func parseLogLevel(s string) zerolog.Level {
	switch s {
	case LevelDebug:
		return zerolog.DebugLevel
	case LevelInfo:
		return zerolog.InfoLevel
	case LevelWarn:
		return zerolog.WarnLevel
	case LevelError:
		return zerolog.ErrorLevel
	case LevelFatal:
		return zerolog.FatalLevel
	case LevelPanic:
		return zerolog.PanicLevel
	default:
		return zerolog.DebugLevel
	}
}
