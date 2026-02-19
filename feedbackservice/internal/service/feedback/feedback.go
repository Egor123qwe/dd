package feedback

import (
	"context"

	"gitlab.roy9.ru/roy9/backend/core/feedbackservice/internal/domain/feedback"
	
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

const (
	localType = "local"
)

type Service struct {
	feedbackRepository FeedbackRepository
}

type FeedbackRepository interface {
	Create(ctx context.Context, fb feedback.Feedback) (int64, error)
	HasFeedbackForRent(ctx context.Context, rentID string) (bool, error)
	CreateFeedbackLocal(ctx context.Context, fl feedback.FeedbackLocal) (int64, error)
	CreateFeedbackPartnership(ctx context.Context, fp feedback.FeedbackPartnership) (int64, error)
}

func New(feedbackRepo FeedbackRepository) *Service {
	svc := &Service{
		feedbackRepository: feedbackRepo,
	}

	return svc
}

func (s *Service) Create(ctx context.Context, fb feedback.Feedback) (int64, error) {
	tracer := otel.Tracer("feedbackservice")
	ctx, span := tracer.Start(ctx, "service.CreateFeedback")
	defer span.End()

	span.SetAttributes(
		attribute.Int("feedback.score", fb.Score),
		attribute.String("feedback.rental_id", fb.RentID),
	)

	_, repoSpan := tracer.Start(ctx, "repository.HasFeedbackForRent")
	exists, err := s.feedbackRepository.HasFeedbackForRent(ctx, fb.RentID)
	if err != nil {
		repoSpan.SetStatus(codes.Error, "repository error")
		repoSpan.RecordError(err)

		return 0, err
	}

	if exists {
		repoSpan.SetStatus(codes.Error, "feedback already exists")
		repoSpan.RecordError(feedback.ErrFeedbackExists)

		return 0, feedback.ErrFeedbackExists
	}
	repoSpan.End()

	_, repoSpan = tracer.Start(ctx, "repository.Create")
	id, err := s.feedbackRepository.Create(ctx, fb)
	if err != nil {
		repoSpan.SetStatus(codes.Error, "repository error")
		repoSpan.RecordError(err)

		return 0, err
	}
	repoSpan.End()

	span.SetAttributes(attribute.Int64("created_feedback_id", id))
	return id, nil
}

func (s *Service) CreateFeedbackLocal(ctx context.Context, fl feedback.FeedbackLocal) (int64, error) {
	tracer := otel.Tracer("feedbackservice")
	ctx, span := tracer.Start(ctx, "service.CreateFeedbackLocal")
	defer span.End()

	fl.Type = localType

	_, repoSpan := tracer.Start(ctx, "repository.CreateFeedbackLocal")
	id, err := s.feedbackRepository.CreateFeedbackLocal(ctx, fl)
	if err != nil {
		repoSpan.SetStatus(codes.Error, "repository error")
		repoSpan.RecordError(err)
		repoSpan.End()

		return 0, err
	}
	repoSpan.End()

	return id, nil
}

func (s *Service) CreateFeedbackPartnership(ctx context.Context, fp feedback.FeedbackPartnership) (int64, error) {
	tracer := otel.Tracer("feedbackservice")
	ctx, span := tracer.Start(ctx, "service.CreateFeedbackPartnership")
	defer span.End()

	_, repoSpan := tracer.Start(ctx, "repository.CreateFeedbackPartnership")
	id, err := s.feedbackRepository.CreateFeedbackPartnership(ctx, fp)
	if err != nil {
		repoSpan.SetStatus(codes.Error, "repository error")
		repoSpan.RecordError(err)
		repoSpan.End()

		return 0, err
	}
	repoSpan.End()

	return id, nil
}
