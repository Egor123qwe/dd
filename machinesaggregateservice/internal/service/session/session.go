package session

import (
	"context"
	"log/slog"

	"gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/internal/model"
)

type Service struct {
	sessionRepo SessionRepository
}

type SessionRepository interface {
	FeedbackList(ctx context.Context, log slog.Logger, sessions []model.Session) ([]model.Session, error)
	SessionList(ctx context.Context, log slog.Logger, filter model.FilterRepo) ([]model.Session, error)
}

func New(sessionRepo SessionRepository) Service {
	s := Service{
		sessionRepo: sessionRepo,
	}

	return s
}

func (s Service) ReceiveByID(ctx context.Context, log slog.Logger, sessionID string) (model.SessionResponse, error) {
	filter := model.FilterRepo{
		ID:   sessionID,
		Type: model.TypeSession,
	}

	sess, err := s.sessionRepo.SessionList(ctx, log, filter)
	if err != nil {
		return model.SessionResponse{}, err
	}

	sess, err = s.sessionRepo.FeedbackList(ctx, log, sess)
	if err != nil {
		return model.SessionResponse{}, err
	}

	res := s.newSessionResponse(sess[0])

	return res, nil
}

func (s Service) ListByTariffID(ctx context.Context, log slog.Logger, tariffID string) (model.SessionList, error) {
	var response model.SessionList

	filter := model.FilterRepo{
		ID:   tariffID,
		Type: model.TypeTariff,
	}

	sessions, err := s.sessionRepo.SessionList(ctx, log, filter)
	if err != nil {
		return model.SessionList{}, err
	}

	sessions, err = s.sessionRepo.FeedbackList(ctx, log, sessions)
	if err != nil {
		return model.SessionList{}, err
	}

	response.Sessions = sessions
	newResponse := s.storangeSum(&response)

	return *newResponse, nil
}

func (s Service) ListByGPUID(ctx context.Context, log slog.Logger, gpuDictID string) (model.SessionList, error) {
	var response model.SessionList

	filter := model.FilterRepo{
		ID:   gpuDictID,
		Type: model.TypeGPUDict,
	}

	sessions, err := s.sessionRepo.SessionList(ctx, log, filter)
	if err != nil {
		return model.SessionList{}, err
	}

	sessions, err = s.sessionRepo.FeedbackList(ctx, log, sessions)
	if err != nil {
		return model.SessionList{}, err
	}

	response.Sessions = sessions
	newResponse := s.storangeSum(&response)

	return *newResponse, nil
}

func (s Service) newSessionResponse(session model.Session) model.SessionResponse {
	return model.SessionResponse{
		Category: model.Category{
			GPU: session.GPUs[0].GPUDict,
		},
		Session: session,
	}
}

func (s Service) storangeSum(response *model.SessionList) *model.SessionList {
	var maxRam, minRam, maxStorage, minStorage int64
	for _, session := range response.Sessions {
		for _, storage := range session.Storage {
			if storage.Total > maxStorage {
				maxStorage = storage.Total
			}

			if storage.Total < minStorage || minStorage == 0 {
				minStorage = storage.Total
			}
		}
		if session.TotalRam > maxRam {
			maxRam = session.TotalRam
		}

		if session.TotalRam < minRam || minRam == 0 {
			minRam = session.TotalRam
		}
	}
	response.MaxRAM = maxRam
	response.MinRAM = minRam
	response.MinStorage = minStorage
	response.MaxStorage = maxStorage

	return response
}
