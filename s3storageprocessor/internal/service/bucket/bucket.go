package bucket

import (
	"context"
	"errors"

	"github.com/rs/zerolog"
	"gitlab.roy9.ru/roy9/backend/core/s3storageprocessor/internal/config"
)

var (
	ErrBucketAlreadyExists = errors.New("bucket already exists")
)

type Service struct {
	cfg     config.S3
	storage Storage
}

type Storage interface {
	BucketExists(ctx context.Context, log zerolog.Logger, userID string) (bool, error)
	MakeBucket(ctx context.Context, log zerolog.Logger, userID string) error
	SetBucketQuota(ctx context.Context, log zerolog.Logger, userID string, quota int64) error
}

func New(cfg config.S3, storage Storage) Service {
	s := Service{
		cfg:     cfg,
		storage: storage,
	}

	return s
}

func (s Service) Create(ctx context.Context, log zerolog.Logger, userID string) error {
	exist, err := s.storage.BucketExists(ctx, log, userID)
	if err != nil {
		return err
	}

	if exist {
		return ErrBucketAlreadyExists
	}

	err = s.storage.MakeBucket(ctx, log, userID)
	if err != nil {
		return err
	}

	err = s.storage.SetBucketQuota(ctx, log, userID, s.cfg.DefaultQuota)
	if err != nil {
		return err
	}

	return nil
}

func (s Service) ChangeQuota(ctx context.Context, log zerolog.Logger, userID string, qouta int64) error {
	return s.storage.SetBucketQuota(ctx, log, userID, qouta)
}
