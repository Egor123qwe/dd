package bucket

import (
	"context"

	bucketv1 "gitlab.roy9.ru/roy9/backend/core/s3storageprocessor/internal/proto/gen/bucket.v1"

	"github.com/rs/zerolog"
)

type Handler struct {
	log     zerolog.Logger
	service BucketService
	bucketv1.UnimplementedBucketServer
}

type BucketService interface {
	Create(ctx context.Context, log zerolog.Logger, userID string) error
	ChangeQuota(ctx context.Context, log zerolog.Logger, userID string, qouta int64) error
}

func New(log zerolog.Logger, service BucketService) Handler {
	h := Handler{
		log:     log,
		service: service,
	}

	return h
}

func (h Handler) Create(ctx context.Context, req *bucketv1.BucketRequest) (*bucketv1.BucketResponse, error) {
	err := h.service.Create(ctx, h.log, req.UserId)
	if err != nil {
		return &bucketv1.BucketResponse{}, err
	}

	res := &bucketv1.BucketResponse{
		UserId:   req.UserId,
		BucketId: req.UserId,
	}

	return res, nil
}
