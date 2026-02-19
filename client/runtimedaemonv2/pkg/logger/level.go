package logger

import "github.com/op/go-logging"

const (
	LevelDebug    = "DEBUG"
	LevelInfo     = "INFO"
	LevelNotice   = "NOTICE"
	LevelWarn     = "WARNING"
	LevelError    = "ERROR"
	LevelCritical = "CRITICAL"
)

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

func availableLevel(min string, target string) bool {
	return parseLogLevel(min) >= parseLogLevel(target)
}
