package logger

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/op/go-logging"
)

const (
	getStateTimeout = 5 * time.Second
)

type Logger interface {
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Panic(args ...interface{})
	Panicf(format string, args ...interface{})
	Critical(args ...interface{})
	Criticalf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Warning(args ...interface{})
	Warningf(format string, args ...interface{})
	Notice(args ...interface{})
	Noticef(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
}

type logger struct {
	Logger        *logging.Logger
	Module        string
	defaultParams LogParams
}

func NewLogger(module string, defaultParams LogParams) Logger {
	logger := logger{
		Logger:        logging.MustGetLogger(module),
		Module:        module,
		defaultParams: defaultParams,
	}

	return logger
}

func (l logger) resolveParams(args ...interface{}) (LogParams, bool) {
	if len(args) > 0 {
		p, ok := args[len(args)-1].(LogParams)
		if ok {
			return p, true
		}
	}

	return l.defaultParams, false
}

func (l logger) toSentry(p LogParams, msg string, level string) {
	if p.ToSentry && availableLevel(p.ToSentryLevel, level) {
		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetTag("package", l.Module)

			if opts.StateGetter != nil {
				ctx, _ := context.WithTimeout(context.Background(), getStateTimeout)
				scopeMap := opts.StateGetter(ctx)

				for k, v := range scopeMap {
					scope.SetExtra(k, v)
				}
			}

			if availableLevel(LevelError, level) {
				sentry.CaptureException(errors.New(msg))
			} else {
				sentry.CaptureMessage(msg)
			}
		})
	}
}

func (l logger) Fatal(args ...interface{}) {
	p, ok := l.resolveParams(args...)
	if ok {
		args = args[:len(args)-1]
	}

	l.toSentry(p, fmt.Sprint(args...), LevelCritical)

	l.Logger.Fatal(args...)
}

// Fatalf is equivalent to l.Critical followed by a call to os.Exit(1).
func (l logger) Fatalf(format string, args ...interface{}) {
	p, ok := l.resolveParams(args...)
	if ok {
		args = args[:len(args)-1]
	}

	l.toSentry(p, fmt.Sprintf(format, args...), LevelCritical)

	l.Logger.Fatalf(format, args...)
}

func (l logger) Panic(args ...interface{}) {
	p, ok := l.resolveParams(args...)
	if ok {
		args = args[:len(args)-1]
	}

	l.toSentry(p, fmt.Sprint(args...), LevelCritical)

	l.Logger.Panic(args...)
}

func (l logger) Panicf(format string, args ...interface{}) {
	p, ok := l.resolveParams(args...)
	if ok {
		args = args[:len(args)-1]
	}

	l.toSentry(p, fmt.Sprintf(format, args...), LevelCritical)

	l.Logger.Panicf(format, args...)
}

func (l logger) Critical(args ...interface{}) {
	p, ok := l.resolveParams(args...)
	if ok {
		args = args[:len(args)-1]
	}

	l.toSentry(p, fmt.Sprint(args...), LevelCritical)

	l.Logger.Critical(args...)
}

func (l logger) Criticalf(format string, args ...interface{}) {
	p, ok := l.resolveParams(args...)
	if ok {
		args = args[:len(args)-1]
	}

	l.toSentry(p, fmt.Sprintf(format, args...), LevelCritical)

	l.Logger.Criticalf(format, args...)
}

func (l logger) Error(args ...interface{}) {
	p, ok := l.resolveParams(args...)
	if ok {
		args = args[:len(args)-1]
	}

	l.toSentry(p, fmt.Sprint(args...), LevelError)

	l.Logger.Error(args...)
}

func (l logger) Errorf(format string, args ...interface{}) {
	p, ok := l.resolveParams(args...)
	if ok {
		args = args[:len(args)-1]
	}

	l.toSentry(p, fmt.Sprintf(format, args...), LevelError)

	l.Logger.Errorf(format, args...)
}

func (l logger) Warning(args ...interface{}) {
	p, ok := l.resolveParams(args...)
	if ok {
		args = args[:len(args)-1]
	}

	l.toSentry(p, fmt.Sprint(args...), LevelWarn)

	l.Logger.Warning(args...)
}

func (l logger) Warningf(format string, args ...interface{}) {
	p, ok := l.resolveParams(args...)
	if ok {
		args = args[:len(args)-1]
	}

	l.toSentry(p, fmt.Sprintf(format, args...), LevelWarn)

	l.Logger.Warningf(format, args...)
}

func (l logger) Notice(args ...interface{}) {
	p, ok := l.resolveParams(args...)
	if ok {
		args = args[:len(args)-1]
	}

	l.toSentry(p, fmt.Sprint(args...), LevelNotice)

	l.Logger.Notice(args...)
}

func (l logger) Noticef(format string, args ...interface{}) {
	p, ok := l.resolveParams(args...)
	if ok {
		args = args[:len(args)-1]
	}

	l.toSentry(p, fmt.Sprintf(format, args...), LevelNotice)

	l.Logger.Noticef(format, args...)
}

func (l logger) Info(args ...interface{}) {
	p, ok := l.resolveParams(args...)
	if ok {
		args = args[:len(args)-1]
	}

	l.toSentry(p, fmt.Sprint(args...), LevelInfo)

	l.Logger.Info(args...)
}

func (l logger) Infof(format string, args ...interface{}) {
	p, ok := l.resolveParams(args...)
	if ok {
		args = args[:len(args)-1]
	}

	l.toSentry(p, fmt.Sprintf(format, args...), LevelInfo)

	l.Logger.Infof(format, args...)
}

func (l logger) Debug(args ...interface{}) {
	p, ok := l.resolveParams(args...)
	if ok {
		args = args[:len(args)-1]
	}

	l.toSentry(p, fmt.Sprint(args...), LevelDebug)

	l.Logger.Debug(args...)
}

func (l logger) Debugf(format string, args ...interface{}) {
	p, ok := l.resolveParams(args...)
	if ok {
		args = args[:len(args)-1]
	}

	l.toSentry(p, fmt.Sprintf(format, args...), LevelDebug)

	l.Logger.Debugf(format, args...)
}
