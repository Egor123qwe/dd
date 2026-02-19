package logger

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/op/go-logging"
	"gopkg.in/natefinch/lumberjack.v2"
)

var cfg config
var opts Options

type Options struct {
	StateGetter func(ctx context.Context) map[string]interface{}
}

func Init(options Options) error {
	cfg = newConfig()
	opts = options

	logLevel := parseLogLevel(cfg.Level)
	backends := make([]logging.Backend, 0)

	if cfg.ToFile {
		logFile := &lumberjack.Logger{
			Filename:   cfg.Fn,
			MaxSize:    cfg.MaxSizeMb,
			MaxBackups: cfg.MaxFiles,
		}

		backends = append(backends, setupLogBackend(logFile, logLevel))
	}

	if cfg.ToStderr {
		backends = append(backends, setupLogBackend(os.Stderr, logLevel))
	}

	logging.SetBackend(backends...)

	if cfg.ToSentry {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              cfg.SentryDSN,
			TracesSampleRate: 1.0,
		})

		if err != nil {
			return fmt.Errorf("failed to init sentry: %w", err)
		}
	}

	return nil
}

func setupLogBackend(out io.Writer, logLevel logging.Level) logging.Backend {
	backend := logging.NewLogBackend(out, "", 0)

	formatter := logging.NewBackendFormatter(backend, logFormat)

	leveled := logging.AddModuleLevel(formatter)
	leveled.SetLevel(logLevel, "")

	return leveled
}
