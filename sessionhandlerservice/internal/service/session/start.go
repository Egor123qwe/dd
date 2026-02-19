package session

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model/rent"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model/session"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/db/redis/cache"
)

func (s service) Init(ctx context.Context, req session.InitReq) (session.InitResp, error) {
	rentConfiguration, err := s.rent.GenerateRentSettings(ctx, req.Settings)
	if err != nil {
		return session.InitResp{}, fmt.Errorf("failed to generate rent settings: %w", err)
	}

	merchantIDs, err := s.storage.Session().GetMerchantIDs(ctx, req.SessionID)
	if err != nil {
		return session.InitResp{}, fmt.Errorf("failed to get merchant: %w", err)
	}

	// create repo transaction service
	rentRepo, tx, err := s.storage.Rent().WithTransaction(ctx)
	if err != nil {
		return session.InitResp{}, fmt.Errorf("failed to start rent transaction: %w", err)
	}

	defer tx.Rollback()

	// create rent in postgres
	rent := rent.Rent{
		SessionID: req.SessionID,
		ClientId:  req.Client.UserID,
		Status:    rent.PendingRentStatus,
	}

	rent, err = rentRepo.Create(ctx, rent, rentConfiguration)
	if err != nil {
		return session.InitResp{}, fmt.Errorf("failed to create rent: %w", err)
	}

	if err = rentRepo.ChangeMerchantStatus(ctx, req.SessionID, string(session.ReservedMerchantStatus)); err != nil {
		return session.InitResp{}, fmt.Errorf("failed to change merchant status: %w", err)
	}

	lastValue, err := s.storage.Cache().GetSetIfEquals(
		ctx, string(SessionIDType)+req.SessionID,
		string(session.ReadyMerchantStatus), string(session.ReservedMerchantStatus), redis.KeepTTL,
	)

	if err != nil {
		switch {
		case errors.Is(err, cache.ErrNotEqual):
			{
				switch session.MerchantStatus(lastValue) {
				case session.PendingMerchantStatus:
					return session.InitResp{}, model.ErrMerchantNotReady

				case session.ReservedMerchantStatus:
					return session.InitResp{}, model.ErrMerchantAlreadyReserved
				}
			}

		case errors.Is(err, cache.ErrTransactionFailed):
			// in such error case we should return ErrMerchantAlreadyReserved,
			// because it means that the value changed in transaction process by another client
			return session.InitResp{}, model.ErrMerchantAlreadyReserved
		}

		return session.InitResp{}, fmt.Errorf("failed to get merchant status: %w", err)
	}

	err = s.storage.Cache().Set(ctx, string(ClientIDType)+req.Client.UserID, "", s.config.clientTTL)
	if err != nil {
		return session.InitResp{}, fmt.Errorf("failed to set client id in redis: %w", err)
	}

	err = s.storage.Cache().Set(ctx, string(RequestIDType)+rent.ID, "", s.config.rentTTL)
	if err != nil {
		return session.InitResp{}, fmt.Errorf("failed to set client id in redis: %w", err)
	}

	tx.Commit()

	merchant := session.Merchant{
		UserID: merchantIDs.UserID,
		ConnID: merchantIDs.ConnID,
	}

	result := session.InitResp{
		RequestID: rent.ID,
		Merchant:  merchant,

		CreatedAt: time.Now(),

		Settings: rentConfiguration,
	}

	return result, nil
}

func (s service) Start(ctx context.Context, req session.StartReq) (session.StartResp, error) {
	currentRent, err := s.storage.Rent().Get(ctx, req.RequestID)
	if err != nil {
		return session.StartResp{}, fmt.Errorf("failed to get rent: %w", err)
	}

	if err := s.storage.Rent().ChangeStatus(ctx, req.RequestID, rent.StartedRentStatus); err != nil {
		return session.StartResp{}, fmt.Errorf("failed to change rent status: %w", err)
	}
	
	client := session.Client{
		UserID: currentRent.ClientId,
	}

	result := session.StartResp{
		Client:    client,
		SessionID: currentRent.SessionID,
	}

	return result, nil
}
