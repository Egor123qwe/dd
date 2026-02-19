package category

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gitlab.roy9.ru/roy9/backend/statemachine/machinesaggregateservice/internal/model"

	"github.com/jmoiron/sqlx"
)

const (
	queryTimeout = 5 * time.Second
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

func (r Repository) GPUDictList(ctx context.Context, log slog.Logger, vramFrom, vramTo int) ([]model.GPUDict, error) {
	c, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	rows, err := r.db.QueryxContext(c, GPUDictQuery, model.StatusReady, vramFrom, vramTo)
	if err != nil {
		log.Error(err.Error())
		return []model.GPUDict{}, fmt.Errorf("GetGPUNames: %v", err)
	}
	defer rows.Close()

	var categories []model.GPUDict

	for rows.Next() {
		var category model.GPUDict

		if err := rows.StructScan(&category); err != nil {
			return []model.GPUDict{}, fmt.Errorf("GetGPUNames : %v", err)
		}

		categories = append(categories, category)
	}

	if rows.Err() != nil {
		log.Error(rows.Err().Error())
		return []model.GPUDict{}, fmt.Errorf("GetGPUCNames : %v", rows.Err())
	}

	if len(categories) == 0 {
		return []model.GPUDict{}, model.ErrCatsNotFound
	}

	return categories, nil
}
