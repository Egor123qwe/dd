package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/repo/cache"
)

var (
	ErrNotFound = errors.New("not found")
	ErrNotEqual = errors.New("not equal")

	ErrTransactionFailed = errors.New("transaction failed")
)

const (
	requestTimeout = 15 * time.Second
)

type repo struct {
	db *redis.Client
}

func New(db *redis.Client) cache.Repo {
	return repo{
		db: db,
	}
}

func (r repo) Get(ctx context.Context, key string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	val, err := r.db.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrNotFound

		}

		return "", err
	}

	return val, nil
}

func (r repo) Set(ctx context.Context, key string, value any, exp time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	return r.db.Set(ctx, key, value, exp).Err()
}

func (r repo) SetIfExists(ctx context.Context, key string, value any, exp time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	return r.db.SetXX(ctx, key, value, exp).Err()
}

// SetIfEquals sets new value if old value equals to compare value
// this code use redis "Watch" method, so if value changed in transaction, it will return ErrTransactionFailed
func (r repo) SetIfEquals(ctx context.Context, key string, compare any, value any, exp time.Duration) error {
	_, err := r.GetSetIfEquals(ctx, key, compare, value, exp)
	return err
}

// GetSetIfEquals returns old value and sets new value if old value equals to compare value
// this code use redis "Watch" method, so if value changed in transaction, it will return ErrTransactionFailed
func (r repo) GetSetIfEquals(ctx context.Context, key string, compare any, value any, exp time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	var val string
	var err error

	err = r.db.Watch(ctx, func(tx *redis.Tx) error {
		val, err = tx.Get(ctx, key).Result()
		if err != nil {
			return err
		}

		if val != compare {
			return ErrNotEqual
		}

		// runs only if the watched keys remain unchanged
		_, err = tx.Pipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, key, value, exp)

			return nil
		})

		return err
	}, key)

	if errors.Is(err, redis.TxFailedErr) {
		return "", ErrTransactionFailed
	}

	return val, err
}

func (r repo) Delete(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	if err := r.db.Del(ctx, key).Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrNotFound
		}

		return err
	}

	return nil
}
