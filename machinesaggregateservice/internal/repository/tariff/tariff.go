package tariff

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/internal/model"

	"github.com/jmoiron/sqlx"
)

const (
	queryTimeout = 10 * time.Second
)

type Repository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) Repository {
	r := Repository{
		db: db,
	}

	return r
}

func (r Repository) List(ctx context.Context, log slog.Logger) ([]model.Tariff, error) {
	var (
		t         model.Tariff
		tariffs   []model.Tariff
		sessionID string
		GPUDict   model.GPUDict
	)

	tx, err := r.db.Beginx()
	if err != nil {
		return []model.Tariff{}, fmt.Errorf("TariffList: %v", err)
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Error("rollback error: %v", err.Error(), nil)
			}
		}
	}()

	c, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	rows, err := tx.QueryxContext(c, TariffListQuery, model.StatusReady, model.Effective, model.Optimal)
	if err != nil {
		log.Error("List of tariffs", "error", err.Error())
		return []model.Tariff{}, fmt.Errorf("TariffList: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.StructScan(&t); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return []model.Tariff{}, model.ErrCatsNotFound
			}

			return nil, fmt.Errorf("TariffList: %v", err)
		}

		tariffs = append(tariffs, t)
	}

	sesRows, err := tx.QueryxContext(c, SessionsForCategory, model.StatusReady)
	if err != nil {
		return []model.Tariff{}, fmt.Errorf("TariffList: %v", err)
	}
	defer sesRows.Close()

	for sesRows.Next() {
		if err = sesRows.Scan(&t.ID, &sessionID); err != nil {
			return []model.Tariff{}, fmt.Errorf("TariffList: %v", err)
		}

		for i := range tariffs {
			if tariffs[i].ID == t.ID {
				tariffs[i].ListSessionID = append(tariffs[i].ListSessionID, sessionID)
				break
			}
		}
	}

	for i := range tariffs {
		if len(tariffs[i].ListSessionID) == 0 {
			tariffs[i].ListSessionID = make([]string, 0)
		}
	}

	dictRows, err := tx.QueryxContext(c, GPUDictForTariff)
	if err != nil {
		return []model.Tariff{}, fmt.Errorf("TariffList: %v", err)
	}
	defer dictRows.Close()

	for dictRows.Next() {
		if err = dictRows.Scan(&GPUDict.Name, &GPUDict.TotalVRAM, &t.ID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return []model.Tariff{}, model.ErrGPUsNotFound
			}

			return []model.Tariff{}, fmt.Errorf("TariffList: %v", err)
		}

		for i := range tariffs {
			if tariffs[i].ID == t.ID {
				tariffs[i].GPUMeta.Name = GPUDict.Name
				tariffs[i].GPUMeta.TotalVRAM = GPUDict.TotalVRAM
				break
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return []model.Tariff{}, fmt.Errorf("TariffList: %v", err)
	}

	return tariffs, nil
}
