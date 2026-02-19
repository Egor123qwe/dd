package logger

import (
	"io"
	"os"

	"github.com/op/go-logging"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	LevelDebug    = "DEBUG"
	LevelInfo     = "INFO"
	LevelNotice   = "NOTICE"
	LevelWarn     = "WARNING"
	LevelError    = "ERROR"
	LevelCritical = "CRITICAL"
)

var log = logging.MustGetLogger("logger")

var logFormat = logging.MustStringFormatter(
	"%{time:2006-01-02 15:04:05.000}: (%{level}) package \"%{module}\": %{message}",
)

func Init() {
	config := newConfig()

	logLevel := parseLogLevel(config.Level)

	backends := make([]logging.Backend, 0)

	if config.ToFile {
		logFile := &lumberjack.Logger{
			Filename:   config.Fn,
			MaxSize:    config.MaxSizeMb,
			MaxBackups: config.MaxFiles,
		}

		backends = append(backends, setupLogBackend(logFile, logLevel))
	}

	if config.ToStderr {
		backends = append(backends, setupLogBackend(os.Stderr, logLevel))
	}

	logging.SetBackend(backends...)
}

func parseLogLevel(s string) logging.Level {
	switch s {
	case LevelDebug:
		return logging.DEBUG

	case LevelInfo:
		return logging.INFO

	case LevelNotice:
		return logging.NOTICE

	case LevelWarn:
		return logging.WARNING

	case LevelError:
		return logging.ERROR

	case LevelCritical:
		return logging.CRITICAL

	default:
		return logging.DEBUG
	}
}

func setupLogBackend(out io.Writer, logLevel logging.Level) logging.Backend {
	backend := logging.NewLogBackend(out, "", 0)

	formatter := logging.NewBackendFormatter(backend, logFormat)

	leveled := logging.AddModuleLevel(formatter)
	leveled.SetLevel(logLevel, "")

	return leveled
}
