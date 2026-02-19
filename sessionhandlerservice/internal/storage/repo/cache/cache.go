package cache

import (
	"context"
	"time"
)

type Repo interface {
	Get(ctx context.Context, key string) (string, error)

	Set(ctx context.Context, key string, value any, exp time.Duration) error
	SetIfExists(ctx context.Context, key string, value any, exp time.Duration) error
	SetIfEquals(ctx context.Context, key string, compare any, value any, exp time.Duration) error

	GetSetIfEquals(ctx context.Context, key string, compare any, value any, exp time.Duration) (string, error)

	Delete(ctx context.Context, key string) error
}
