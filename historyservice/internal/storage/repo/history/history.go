package history

import (
	"context"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	"gitlab.roy9.ru/roy9/backend/core/historyservice/internal/domain/model"
	"gitlab.roy9.ru/roy9/backend/core/historyservice/internal/storage/repo"
)

const (
	queryTimeout   = 7 * time.Second
	finishedStatus = "finished"
)

type historyRepo struct {
	db  *sqlx.DB
	log *slog.Logger
}

func New(db *sqlx.DB, log *slog.Logger) repo.History {
	return historyRepo{db: db, log: log}
}

func (hr historyRepo) RentHistory(ctx context.Context, userID string) ([]model.Rent, error) {
	var rents []model.Rent

	c, cancel := context.WithTimeout(ctx, queryTimeout)

	defer cancel()

	query := `SELECT r.id, r.client_id as user_id,
				r.created_at as started_at, r.deleted_at as ended_at,
				ROUND(EXTRACT(EPOCH FROM (r.deleted_at - r.created_at)) / 60)::int AS duration,
				COALESCE(r.cost, (s.total_price * ROUND(EXTRACT(EPOCH FROM (r.deleted_at - r.created_at)) / 60)::int)::double precision) AS cost,
				s.total_price as price, f.score as rating,
				ts.template_id, ts.title as template_title
				FROM rent as r
				LEFT JOIN session as s ON s.id = r.session_id
				LEFT JOIN feedback as f ON f.rent_id = r.id
				LEFT JOIN rent_settings as rs ON r.settings_id = rs.id
				LEFT JOIN template_settings as ts ON rs.template_id = ts.id
				WHERE r.client_id::text = $1 AND r.status = $2
				`

	rows, err := hr.db.QueryxContext(c, query, userID, finishedStatus)
	if err != nil {
		hr.log.Error("failed to fetch rent history", err.Error(), nil)
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var rent model.Rent
		err := rows.StructScan(&rent)
		if err != nil {
			hr.log.Error("failed to scan row", err.Error(), nil)
			return nil, err
		}

		dur := rent.Duration
		if dur < 0 {
			dur += 180
		}
		rent.Duration = dur

		rents = append(rents, rent)
	}

	if len(rents) == 0 {
		return nil, repo.ErrNoHistory
	}

	return rents, nil
}

func (hr historyRepo) SessionHistory(ctx context.Context, userID string) ([]model.Rent, error) {
	var rents []model.Rent

	c, cancel := context.WithTimeout(context.Background(), queryTimeout)

	defer cancel()

	query := `SELECT r.id as id, s.user_id as user_id,
				r.created_at as started_at, r.deleted_at as ended_at,
				ROUND(EXTRACT(EPOCH FROM (r.deleted_at - r.created_at)) / 60)::int AS duration,
				COALESCE(r.cost, (s.total_price * ROUND(EXTRACT(EPOCH FROM (r.deleted_at - r.created_at)) / 60)::int)::double precision) AS cost,
				s.total_price AS price, f.score as rating,
				ts.template_id, ts.title as template_title
				FROM session as s
				LEFT JOIN rent as r ON s.id = r.session_id
				LEFT JOIN feedback as f ON f.rent_id = r.id
				LEFT JOIN rent_settings as rs ON r.settings_id = rs.id
				LEFT JOIN template_settings as ts ON rs.template_id = ts.id
				WHERE s.user_id = $1 AND r.status = $2
				`
	rows, err := hr.db.QueryxContext(c, query, userID, finishedStatus)
	if err != nil {
		hr.log.Error("failed to fetch lease history", err.Error(), nil)
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var rent model.Rent
		err := rows.StructScan(&rent)
		if err != nil {
			hr.log.Error("failed to scan row", err.Error(), nil)
			return nil, err
		}

		dur := rent.Duration
		if dur < 0 {
			dur += 180
		}
		rent.Duration = dur

		rents = append(rents, rent)
	}

	if len(rents) == 0 {
		return nil, repo.ErrNoHistory
	}

	return rents, nil
}

func (hr historyRepo) AllRents(ctx context.Context) ([]model.AdminRent, error) {
	var out []model.AdminRent

	c, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	query := `SELECT r.id,
				r.client_id::text as client_id,
				s.user_id::text as merchant_id,
				r.created_at as started_at, r.deleted_at as ended_at,
				ROUND(EXTRACT(EPOCH FROM (r.deleted_at - r.created_at)) / 60)::int AS duration,
				COALESCE(r.cost, (s.total_price * ROUND(EXTRACT(EPOCH FROM (r.deleted_at - r.created_at)) / 60)::int)::double precision)::real AS cost,
				f.score as rating,
				ts.template_id, ts.title as template_title
				FROM rent as r
				LEFT JOIN session as s ON s.id = r.session_id
				LEFT JOIN feedback as f ON f.rent_id = r.id
				LEFT JOIN rent_settings as rs ON r.settings_id = rs.id
				LEFT JOIN template_settings as ts ON rs.template_id = ts.id
				WHERE r.status = $1
				ORDER BY r.created_at DESC`

	rows, err := hr.db.QueryxContext(c, query, finishedStatus)
	if err != nil {
		hr.log.Error("failed to fetch all rents", "error", err.Error())
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var row struct {
			ID            string     `db:"id"`
			ClientID      string     `db:"client_id"`
			MerchantID    string     `db:"merchant_id"`
			StartedAt     *time.Time `db:"started_at"`
			EndedAt       *time.Time `db:"ended_at"`
			Duration      int        `db:"duration"`
			Cost          float32    `db:"cost"`
			Rating        *int       `db:"rating"`
			TemplateID    string     `db:"template_id"`
			TemplateTitle string     `db:"template_title"`
		}
		if err := rows.StructScan(&row); err != nil {
			hr.log.Error("failed to scan row", "error", err.Error())
			return nil, err
		}
		dur := row.Duration
		if dur < 0 {
			dur += 180
		}
		out = append(out, model.AdminRent{
			ID:            row.ID,
			ClientID:      row.ClientID,
			MerchantID:    row.MerchantID,
			StartedAt:     row.StartedAt,
			EndedAt:       row.EndedAt,
			Duration:      dur,
			Cost:          row.Cost,
			Rating:        row.Rating,
			TemplateID:    row.TemplateID,
			TemplateTitle: row.TemplateTitle,
		})
	}

	return out, nil
}
