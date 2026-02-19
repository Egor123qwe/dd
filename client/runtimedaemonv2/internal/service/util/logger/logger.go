package logger

import (
	"context"

	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/storage"
	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/pkg/logger"
)

func NewLoggerOpts() (logger.Options, error) {
	s, err := storage.New()
	if err != nil {
		return logger.Options{}, err
	}

	result := logger.Options{
		StateGetter: func(ctx context.Context) map[string]interface{} {
			settings, err := s.Runtime().Settings().Get()
			if err != nil {
				return nil
			}

			return map[string]interface{}{
				"mode": settings.Mode.String(),
			}
		},
	}

	return result, nil
}
