package server

import (
	"context"
	"fmt"
	"net"

	"gitlab.roy9.ru/roy9/backend/core/s3storageprocessor/internal/config"
	"gitlab.roy9.ru/roy9/backend/core/s3storageprocessor/internal/handler/bucket"
	bucketv1 "gitlab.roy9.ru/roy9/backend/core/s3storageprocessor/internal/proto/gen/bucket.v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	grpcServer *grpc.Server
	cfg        config.GRPC
	log        zerolog.Logger
}

func New(cfg config.GRPC, log zerolog.Logger, bucketHandler bucket.Handler) *GRPCServer {
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(
			logging.PayloadReceived, logging.PayloadSent,
		),
	}

	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandlerContext(func(ctx context.Context, p any) (err error) {
			return status.Errorf(codes.Unknown, "panic triggered: %v", p)
		}),
	}

	grpcSrv := grpc.NewServer(grpc.ChainUnaryInterceptor(
		logging.UnaryServerInterceptor(interceptorLogger(log), loggingOpts...),
		recovery.UnaryServerInterceptor(recoveryOpts...),
	))

	reflection.Register(grpcSrv)

	bucketv1.RegisterBucketServer(grpcSrv, bucketHandler)

	grpcServer := &GRPCServer{
		grpcServer: grpcSrv,
		cfg:        cfg,
		log:        log,
	}

	return grpcServer
}

func (s *GRPCServer) Run() {
	addr := fmt.Sprintf("%s:%s", s.cfg.Host, s.cfg.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		s.log.Fatal().Err(err).Msg("can not start grpc server")
		return
	}

	s.log.Info().Msgf("server started on %s", addr)

	if err := s.grpcServer.Serve(listener); err != nil {
		s.log.Fatal().Err(err).Msg("can not start gRPC server")
		return
	}
}

func (s *GRPCServer) Stop() {
	s.grpcServer.GracefulStop()
	s.log.Info().Msg("stopping gRPC server")
}

func interceptorLogger(l zerolog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l := l.With().Fields(fields).Logger()

		switch lvl {
		case logging.LevelDebug:
			l.Debug().Msg(msg)

		case logging.LevelInfo:
			l.Info().Msg(msg)

		case logging.LevelWarn:
			l.Warn().Msg(msg)

		case logging.LevelError:
			l.Error().Msg(msg)

		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}
