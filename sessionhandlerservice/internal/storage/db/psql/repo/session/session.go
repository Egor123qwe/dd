package session

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	sessionModel "gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model/session"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/repo/session"
)

const (
	requestTimeout = 15 * time.Second
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

type repo struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) session.Session {
	return &repo{
		db: db,
	}
}

func (r repo) GetMerchantIDs(ctx context.Context, sessionID string) (sessionModel.Merchant, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	var merchant sessionModel.Merchant

	query := `SELECT 
    			  user_id, connection_id
			  FROM session WHERE id = $1`

	err := r.db.QueryRowContext(
		ctx,
		query,
		sessionID,
	).Scan(
		&merchant.UserID,
		&merchant.ConnID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sessionModel.Merchant{}, ErrSessionNotFound
		}

		return sessionModel.Merchant{}, err
	}

	return merchant, nil
}

func (r repo) GetSessionIDByMerchant(ctx context.Context, merchantID string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	var sessionID string

	query := `SELECT 
    			  id
			  FROM session WHERE user_id = $1`

	err := r.db.QueryRowContext(
		ctx,
		query,
		merchantID,
	).Scan(
		&sessionID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrSessionNotFound
		}

		return "", err
	}

	return sessionID, nil
}

func (r repo) GetTotalPrice(ctx context.Context, sessionID string) (float64, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	var totalPrice float64
	query := `SELECT COALESCE(total_price, 0) FROM session WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(&totalPrice)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return totalPrice, nil
}
