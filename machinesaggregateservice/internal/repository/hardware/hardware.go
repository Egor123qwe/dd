package hardware

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

func (r Repository) GPUList(ctx context.Context, log slog.Logger, filter model.FilterRepo) ([]model.GPU, error) {
	var gpus []model.GPU
	var gpu model.GPU

	c, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	rows, err := r.queryWithFilter(c, filter, GPUListQuery)
	if err != nil {
		log.Error(err.Error())
		return []model.GPU{}, fmt.Errorf("GetGPUs: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(
			&gpu.ID,
			&gpu.Name,
			&gpu.AvailableVRAM,
			&gpu.UsedVRAM,
			&gpu.TotalVRAM,
			&gpu.Dlperf,
			&gpu.Price,
		); err != nil {
			return []model.GPU{}, fmt.Errorf("GetGPUs : %v", err)
		}
		gpus = append(gpus, gpu)
	}

	if rows.Err() != nil {
		log.Error(fmt.Sprintf("%v", rows.Err()))
		return []model.GPU{}, fmt.Errorf("GetGPUs : %v", rows.Err())
	}

	if len(gpus) == 0 {
		return []model.GPU{}, model.ErrGPUsNotFound
	}

	return gpus, nil
}

func (r Repository) queryWithFilter(ctx context.Context, filter model.FilterRepo, query string) (*sqlx.Rows, error) {

	switch filter.Type {
	case model.TypeGPU:
		query += "where g.id = $1"

		return r.db.QueryxContext(ctx, query, filter.ID)

	case model.TypeGPUDict:
		query += "and gd.id = $1"

		return r.db.QueryxContext(ctx, query, filter.ID)

	case model.TypeSession:
		query += "and s.id = $1"

		return r.db.QueryxContext(ctx, query, filter.ID)

	case model.TypeTariff:
		query += `
				AND c.id = $1
			`
		return r.db.QueryxContext(ctx, query, filter.ID)

	default:
		return r.db.QueryxContext(ctx, query)
	}
}
