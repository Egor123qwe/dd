package session

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model/session"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/db/redis/cache"
)

func (s service) Stop(ctx context.Context, req session.StopReq) (session.StopResp, error) {
	rent, err := s.storage.Rent().Get(ctx, req.RequestID)
	if err != nil {
		return session.StopResp{}, fmt.Errorf("failed to get rent: %w", err)
	}

	merchantIDs, err := s.storage.Session().GetMerchantIDs(ctx, rent.SessionID)
	if err != nil {
		return session.StopResp{}, fmt.Errorf("failed to get merchant: %w", err)
	}

	// cost = price_per_minute * floor(duration_minutes + 180)
	endTime := time.Now().UTC()
	durationMinutesFloat := endTime.Sub(rent.CreatedAt).Seconds() / 60
	const bonusMinutes = 180
	durationMinutesFloatWithBonus := durationMinutesFloat + float64(bonusMinutes)
	durationMinutes := int64(math.Floor(durationMinutesFloatWithBonus))
	if durationMinutes < 0 {
		durationMinutes = 0
	}
	pricePerMinute, _ := s.storage.Session().GetTotalPrice(ctx, rent.SessionID)
	cost := pricePerMinute * float64(durationMinutes)
	cost = math.Round(cost*100) / 100

	// Списание с клиента и начисление продавцу в той же БД (таблица user_balance).
	clientID, err := strconv.Atoi(strings.TrimSpace(rent.ClientId))
	if err != nil {
		return session.StopResp{}, fmt.Errorf("client_user_id must be numeric: %w", err)
	}
	merchantID, err := strconv.Atoi(strings.TrimSpace(merchantIDs.UserID))
	if err != nil {
		return session.StopResp{}, fmt.Errorf("merchant_user_id must be numeric: %w", err)
	}
	if err := s.storage.Balance().SettleRent(ctx, clientID, merchantID, cost, 0.9); err != nil {
		return session.StopResp{}, fmt.Errorf("balance settlement: %w", err)
	}

	// create repo transaction service
	rentRepo, tx, err := s.storage.Rent().WithTransaction(ctx)
	if err != nil {
		return session.StopResp{}, fmt.Errorf("failed to start rent transaction: %w", err)
	}

	defer tx.Rollback()

	if err := rentRepo.Stop(ctx, req.RequestID, req.Reason, cost); err != nil {
		return session.StopResp{}, fmt.Errorf("failed to delete rent: %w", err)
	}

	if err = rentRepo.ChangeMerchantStatus(ctx, rent.SessionID, string(session.ReadyMerchantStatus)); err != nil {
		return session.StopResp{}, fmt.Errorf("failed to change merchant status: %w", err)
	}

	// change session status (if session already removed, it's not an error)
	err = s.storage.Cache().SetIfExists(ctx, string(SessionIDType)+rent.SessionID, string(session.ReadyMerchantStatus), redis.KeepTTL)
	if err != nil && !errors.Is(err, cache.ErrNotFound) {
		return session.StopResp{}, fmt.Errorf("failed to delete session: %w", err)
	}

	// remove requestID from cache
	err = s.storage.Cache().Delete(ctx, string(RequestIDType)+rent.ID)
	if err != nil && !errors.Is(err, cache.ErrNotFound) {
		return session.StopResp{}, fmt.Errorf("failed to delete requestID from cache: %w", err)
	}

	tx.Commit()

	merchant := session.Merchant{
		UserID: merchantIDs.UserID,
		ConnID: merchantIDs.ConnID,
	}

	client := session.Client{
		UserID: rent.ClientId,
	}

	result := session.StopResp{
		Merchant: merchant,
		Client:   client,

		SessionID: rent.SessionID,
	}

	return result, nil
}
