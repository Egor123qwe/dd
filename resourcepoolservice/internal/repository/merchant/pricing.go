package merchant

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/domain/category"

	sq "github.com/Masterminds/squirrel"
	"github.com/rs/zerolog/log"
	"github.com/texttheater/golang-levenshtein/levenshtein"
)

const (
	gpuSimilarityThreshold = 15 // max Levenshtein distance to consider a match
)

// GetPricingConfig returns the single row from pricing_config.
func (r *Repository) GetPricingConfig(ctx context.Context) (category.PricingConfig, error) {
	c, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	var cfg category.PricingConfig
	err := sq.
		Select("base_per_minute", "ram_per_gb_per_minute", "storage_hdd_per_gb_per_minute",
			"storage_ssd_per_gb_per_minute", "internet_per_mbit_per_minute").
		From("pricing_config").
		Limit(1).
		PlaceholderFormat(sq.Dollar).
		RunWith(r.db).
		QueryRowContext(c).
		Scan(&cfg.BasePerMinute, &cfg.RAMPerGBPerMinute, &cfg.StorageHDDPerGBPerMinute,
			&cfg.StorageSSDPerGBPerMinute, &cfg.InternetPerMbitPerMinute)
	if err != nil {
		return category.PricingConfig{}, err
	}
	return cfg, nil
}

// GetCPUByName returns CPU from cpu_dict by normalized name. Returns nil, nil if not found.
func (r *Repository) GetCPUByName(ctx context.Context, name string) (*category.CPUDict, error) {
	c, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	norm := strings.ToLower(strings.TrimSpace(name))
	var cpu category.CPUDict
	err := sq.
		Select("id", "name", "price_per_minute").
		From("cpu_dict").
		Where(sq.Expr("LOWER(TRIM(name)) = ?", norm)).
		PlaceholderFormat(sq.Dollar).
		RunWith(r.db).
		QueryRowContext(c).
		Scan(&cpu.ID, &cpu.Name, &cpu.PricePerMinute)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &cpu, nil
}

// InsertCPU inserts a CPU with price_per_minute = 0. Used when hardware is unrecognized.
func (r *Repository) InsertCPU(ctx context.Context, name string) error {
	c, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	_, err := sq.
		Insert("cpu_dict").
		Columns("name", "price_per_minute").
		Values(strings.TrimSpace(name), 0).
		PlaceholderFormat(sq.Dollar).
		RunWith(r.db).
		ExecContext(c)
	if err != nil {
		log.Warn().Err(err).Str("name", name).Msg("insert cpu_dict")
		return err
	}
	return nil
}

// GetGPUDictList returns all gpu_dict rows (id, name, price) for matching.
func (r *Repository) GetGPUDictList(ctx context.Context) ([]category.GPUDict, error) {
	c, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	rows, err := sq.
		Select("id", "name", "price").
		From("gpu_dict").
		PlaceholderFormat(sq.Dollar).
		RunWith(r.db).
		QueryContext(c)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []category.GPUDict
	for rows.Next() {
		var g category.GPUDict
		var id int64
		var price sql.NullFloat64
		if err := rows.Scan(&id, &g.Name, &price); err != nil {
			return nil, err
		}
		g.ID = fmt.Sprintf("%d", id)
		if price.Valid {
			g.Price = price.Float64
		}
		list = append(list, g)
	}
	return list, rows.Err()
}

// GetGPUBestMatch finds the best matching GPU by name (Levenshtein). Returns (nil, false) if no match within threshold.
func (r *Repository) GetGPUBestMatch(ctx context.Context, name string) (*category.GPUDict, bool) {
	list, err := r.GetGPUDictList(ctx)
	if err != nil || len(list) == 0 {
		return nil, false
	}

	target := []rune(strings.TrimSpace(name))
	bestIdx := -1
	bestDist := gpuSimilarityThreshold + 1

	for i := range list {
		d := levenshtein.DistanceForStrings(target, []rune(list[i].Name), levenshtein.DefaultOptions)
		if d < bestDist {
			bestDist = d
			bestIdx = i
		}
	}
	if bestIdx < 0 || bestDist > gpuSimilarityThreshold {
		return nil, false
	}
	g := list[bestIdx]
	return &g, true
}

// InsertGPU inserts a GPU with price = 0. Used when hardware is unrecognized.
func (r *Repository) InsertGPU(ctx context.Context, name string, totalVram int64, avgDlperf float64) error {
	c, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	_, err := sq.
		Insert("gpu_dict").
		Columns("name", "total_vram", "price", "avg_dlperf").
		Values(strings.TrimSpace(name), totalVram, 0, avgDlperf).
		PlaceholderFormat(sq.Dollar).
		RunWith(r.db).
		ExecContext(c)
	if err != nil {
		log.Warn().Err(err).Str("name", name).Msg("insert gpu_dict")
		return err
	}
	return nil
}
