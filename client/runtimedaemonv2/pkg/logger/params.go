package logger

import (
	"github.com/op/go-logging"
)

var logFormat = logging.MustStringFormatter(
	"%{time:2006-01-02 15:04:05.000}: (%{level}) package \"%{module}\": %{message}",
)

type LogParams struct {
	ToSentry      bool
	ToSentryLevel string
}

func DefaultWithSentry() LogParams {
	return LogParams{ToSentry: true, ToSentryLevel: LevelError}
}
