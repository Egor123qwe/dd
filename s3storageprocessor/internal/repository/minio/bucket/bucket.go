package bucket

import (
	"context"

	"github.com/minio/madmin-go/v3"
	"github.com/minio/minio-go/v7"
	"github.com/rs/zerolog"
)

const (
	kb = 1024
)

type Storage struct {
	minioAdminClient *madmin.AdminClient
	minioClient      *minio.Client
}

func New(minioAdminClient *madmin.AdminClient, minioClient *minio.Client) *Storage {
	s := &Storage{
		minioAdminClient: minioAdminClient,
		minioClient:      minioClient,
	}

	return s
}

func (s *Storage) BucketExists(ctx context.Context, log zerolog.Logger, userID string) (bool, error) {
	return s.minioClient.BucketExists(ctx, userID)
}

func (s *Storage) MakeBucket(ctx context.Context, log zerolog.Logger, userID string) error {
	return s.minioClient.MakeBucket(ctx, userID, minio.MakeBucketOptions{})
}

func (s *Storage) SetBucketQuota(ctx context.Context, log zerolog.Logger, userID string, quota int64) error {
	return s.minioAdminClient.SetBucketQuota(ctx, userID, &madmin.BucketQuota{
		Quota: uint64(quota * kb * kb * kb),
		Type:  madmin.HardQuota,
	})
}
