package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.roy9.ru/roy9/backend/core/s3storageprocessor/internal/config"
	hbucket "gitlab.roy9.ru/roy9/backend/core/s3storageprocessor/internal/handler/bucket"
	"gitlab.roy9.ru/roy9/backend/core/s3storageprocessor/internal/repository/minio/bucket"
	server "gitlab.roy9.ru/roy9/backend/core/s3storageprocessor/internal/server/grpc"
	svcbucket "gitlab.roy9.ru/roy9/backend/core/s3storageprocessor/internal/service/bucket"

	"github.com/minio/madmin-go/v3"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stderr).
		With().
		Timestamp().
		Logger().
		Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeLocation: time.UTC, TimeFormat: time.RFC3339})

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal().Err(err).Msg("can not load config")
	}

	s3Addr := fmt.Sprintf("%s:%s", cfg.S3.Host, cfg.S3.Port)
	creds := credentials.NewStaticV4(cfg.S3.User, cfg.S3.Password, "")

	minioAdminClient, err := madmin.NewWithOptions(s3Addr, &madmin.Options{
		Creds:  creds,
		Secure: false,
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("can not connect to S3")
	}

	minioClient, err := minio.New(s3Addr, &minio.Options{
		Creds:  creds,
		Secure: false,
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("can not connect to S3")
	}

	storage := bucket.New(minioAdminClient, minioClient)
	bucketService := svcbucket.New(cfg.S3, storage)
	bucketHandler := hbucket.New(logger, bucketService)

	grpcServer := server.New(cfg.GRPC, logger, bucketHandler)

	go grpcServer.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	grpcServer.Stop()
	logger.Info().Msg("Server gracefully stopped")
}
