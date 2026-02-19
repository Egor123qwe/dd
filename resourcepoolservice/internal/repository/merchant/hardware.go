package merchant

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/config"
	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/domain/category"
	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/domain/hardware"
	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/domain/sharep2p"
	"gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/domain/status"
	pricingv1 "gitlab.roy9.ru/roy9/backend/statemachine/resourcepoolservice/internal/proto/gen/pricing.v1"

	"github.com/lib/pq"
	sq "github.com/Masterminds/squirrel"
	"github.com/rs/zerolog/log"
	"github.com/texttheater/golang-levenshtein/levenshtein"
)

const (
	queryTimeout    = 5 * time.Second
	sessionIDPrefix = "session_id"
)

type Repository struct {
	db    *sql.DB
	cache Cache
	cfg   config.RedisConfig
}

type Cache interface {
	Set(ctx context.Context, key string, value any, exp time.Duration) error
	Del(ctx context.Context, key string) error
	SetXX(ctx context.Context, key string, value any, exp time.Duration) (bool, error)
}

func New(db *sql.DB, cache Cache, cfg config.RedisConfig) *Repository {
	repo := &Repository{
		db:    db,
		cache: cache,
		cfg:   cfg,
	}

	return repo
}

func (r *Repository) Create(ctx context.Context, hw *pricingv1.HardwareResponse, prepull []hardware.Prepull, userID, connectionID, nodeName string) error {
	c, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	tx, err := r.db.BeginTx(c, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Error().Err(err).Msg("error rolling back transaction")

			return
		}
	}()

	_, err = sq.
		Insert("session").
		Columns(
			"id",
			"node_name",
			"total_ram",
			"available_ram",
			"used_ram",
			"price_ram",
			"load_speed",
			"upload_speed",
			"ping",
			"price_internet",
			"user_id",
			"connection_id",
			"total_price",
			"status",
		).
		Values(
			hw.Id,
			nodeName,
			hw.TotalRam,
			hw.AvailableRam,
			hw.UsedRam,
			hw.PriceRam,
			hw.LoadSpeed,
			hw.UploadSpeed,
			hw.Ping,
			hw.PriceInternet,
			userID,
			connectionID,
			hw.TotalPrice,
			status.CREATED,
		).
		PlaceholderFormat(sq.Dollar).
		RunWith(tx).
		ExecContext(c)

	if err != nil {
		log.Error().Err(err).Msg("error creating session")
		return err
	}

	for _, gpu := range hw.Gpus {
		var (
			gpuDict     category.GPUDict
			gpuDictList []category.GPUDict
		)

		rows, err := sq.
			Select("id", "name").
			From("gpu_dict").
			RunWith(tx).
			QueryContext(c)

		if err != nil {
			log.Error().Err(err).Msg("error getting gpu_dict")
			return err
		}

		for rows.Next() {
			var id int64
			if err = rows.Scan(&id, &gpuDict.Name); err != nil {
				log.Error().Err(err).Msg("error scanning gpu_dict")
				return err
			}
			gpuDict.ID = fmt.Sprintf("%d", id)
			gpuDictList = append(gpuDictList, gpuDict)
		}

		gpuDictID := findMaxSimilarity(gpu.Name, gpuDictList)

		_, err = sq.
			Insert("gpu").
			Columns("session_id", "gpu_dict_id", "name", "total_vram", "available_vram", "used_vram", "dlperf", "price").
			Values(hw.Id, gpuDictID, gpu.Name, gpu.Total, gpu.Available, gpu.Used, gpu.Dlperf, gpu.Price).
			PlaceholderFormat(sq.Dollar).
			RunWith(tx).
			ExecContext(c)

		if err != nil {
			log.Error().Err(err).Msg("error creating gpu")
			return err
		}
	}

	for _, cpu := range hw.Cpus {
		_, err = sq.
			Insert("cpu").
			Columns("session_id", "name", "total", "available", "price").
			Values(hw.Id, cpu.Name, cpu.Total, cpu.Available, cpu.Price).
			PlaceholderFormat(sq.Dollar).
			RunWith(tx).
			ExecContext(c)

		if err != nil {
			log.Error().Err(err).Msg("error creating cpu")
			return err
		}
	}

	for _, storage := range hw.Storages {
		_, err = sq.
			Insert("storage").
			Columns("session_id", "type", "name", "total", "available", "used", "bandwidth", "price").
			Values(hw.Id, storage.Type, storage.Name, storage.Total, storage.Available, storage.Used, storage.Bandwidth, storage.Price).
			PlaceholderFormat(sq.Dollar).
			RunWith(tx).
			ExecContext(c)

		if err != nil {
			log.Error().Err(err).Msg("error creating storage")
			return err
		}
	}

	for _, prepull := range prepull {
		_, err = sq.
			Insert("prepull").
			Columns("session_id", "template_id").
			Values(hw.Id, prepull.TemplateID).
			PlaceholderFormat(sq.Dollar).
			RunWith(tx).
			ExecContext(c)

		if err != nil {
			log.Error().Err(err).Msg("error creating prepull")
			return err
		}
	}

	key := fmt.Sprintf("%s %s", sessionIDPrefix, hw.Id)

	err = r.cache.Set(ctx, key, status.PENDING, time.Minute*time.Duration(r.cfg.TTL))
	if err != nil {
		log.Error().Err(err).Msg("error setting cache")
		return err
	}

	if err = tx.Commit(); err != nil {
		log.Error().Err(err).Msg("error committing tx")
		return err
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, sessionID, deletionReason string) (string, string, error) {
	c, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	var (
		userID       string
		connectionID string
	)

	err := sq.
		Update("session").
		Set("deleted_at", time.Now().UTC()).
		Set("deletion_reason", deletionReason).
		Set("status", status.FINISHED).
		Where(sq.Eq{"id": sessionID}).
		Suffix("RETURNING user_id, connection_id").
		PlaceholderFormat(sq.Dollar).
		RunWith(r.db).
		QueryRowContext(c).
		Scan(&userID, &connectionID)
	if err != nil {
		log.Error().Err(err).Msg("error soft deleting session")
		return "", "", err
	}

	return userID, connectionID, nil
}

func (r *Repository) Stop(ctx context.Context, sessionID, deletionReason string) error {
	c, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	tx, err := r.db.BeginTx(c, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Error().Err(err).Msg("error rolling back transaction")

			return
		}
	}()

	_, err = sq.
		Update("session").
		Set("deleted_at", time.Now().UTC()).
		Set("deletion_reason", deletionReason).
		Set("status", status.FINISHED).
		Where(sq.Eq{"id": sessionID}).
		PlaceholderFormat(sq.Dollar).
		RunWith(tx).
		ExecContext(c)

	if err != nil {
		log.Error().Err(err).Msg("error soft deleting session")
		return err
	}

	key := fmt.Sprintf("session_id %s", sessionID)
	err = r.cache.Del(ctx, key)
	if err != nil {
		log.Error().Err(err).Msg("error setting cache")
		return err
	}

	if err = tx.Commit(); err != nil {
		log.Error().Err(err).Msg("error committing tx")
		return err
	}

	return nil
}

func (r *Repository) Ready(ctx context.Context, sessionID string) error {
	c, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	tx, err := r.db.BeginTx(c, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Error().Err(err).Msg("error rolling back transaction")

			return
		}
	}()

	_, err = sq.Update("session").
		Set("status", status.READY).
		Where(sq.Eq{"id": sessionID}).
		PlaceholderFormat(sq.Dollar).
		RunWith(tx).
		ExecContext(c)
	if err != nil {
		log.Error().Err(err).Msg("error updating session status")
		return err
	}

	ttl := time.Minute * time.Duration(r.cfg.TTL)
	key := fmt.Sprintf("%s %s", sessionIDPrefix, sessionID)

	success, err := r.cache.SetXX(ctx, key, status.READY, ttl)
	if err != nil {
		return err
	}
	if !success {
		return sharep2p.ErrSessionNotFound
	}

	if err = tx.Commit(); err != nil {
		log.Error().Err(err).Msg("error committing tx")
		return err
	}

	return nil
}

// ReadySession — строка сессии, готовая к аренде (deleted_at IS NULL, status = 'ready').
type ReadySession struct {
	ID     string `db:"id"`
	UserID string `db:"user_id"`
}

// ReadySessionDetails — сессия с деталями для списка мерчантов.
type ReadySessionDetails struct {
	ReadySession
	NodeName       string           `db:"node_name"`
	TotalPrice     float64          `db:"total_price"`
	TotalRAM       int64            `db:"total_ram"`
	AvailableRAM   int64            `db:"available_ram"`
	PriceRAM       float64          `db:"price_ram"`
	LoadSpeed      float64          `db:"load_speed"`
	UploadSpeed    float64          `db:"upload_speed"`
	Ping           int64            `db:"ping"`
	PriceInternet  float64          `db:"price_internet"`
	GPUs           []SessionGPU     `json:"gpus"`
	CPUs           []SessionCPU     `json:"cpus"`
	Storages       []SessionStorage `json:"storages"`
	Templates      []string         `json:"templates"` // template_id из prepull
}

type SessionGPU struct {
	Name         string  `json:"name" db:"name"`
	TotalVRAM    int64   `json:"total_vram" db:"total_vram"`
	AvailableVRAM int64  `json:"available_vram" db:"available_vram"`
	Dlperf       float64 `json:"dlperf" db:"dlperf"`
	Price        float64 `json:"price" db:"price"`
}

type SessionCPU struct {
	Name      string  `json:"name" db:"name"`
	Total     int     `json:"total" db:"total"`
	Available int     `json:"available" db:"available"`
	Price     float64 `json:"price" db:"price"`
}

type SessionStorage struct {
	Type      string  `json:"type" db:"type"`
	Name      string  `json:"name" db:"name"`
	Total     int64   `json:"total" db:"total"`
	Available int64   `json:"available" db:"available"`
	Bandwidth float64 `json:"bandwidth" db:"bandwidth"`
	Price     float64 `json:"price" db:"price"`
}

// MerchantSessionRow — строка для списка «мои узлы» поставщика (сессии по user_id, не удалённые).
type MerchantSessionRow struct {
	ID         string  `db:"id"`
	NodeName   string  `db:"node_name"`
	Status     string  `db:"status"`
	TotalPrice float64 `db:"total_price"`
}

// ListSessionsByUserID возвращает сессии поставщика (user_id = userID, deleted_at IS NULL) для портала.
func (r *Repository) ListSessionsByUserID(ctx context.Context, userID string) ([]MerchantSessionRow, error) {
	if userID == "" {
		return nil, nil
	}
	c, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	query := `
		SELECT id, COALESCE(node_name, '') AS node_name, COALESCE(TRIM(status), '') AS status, COALESCE(total_price, 0) AS total_price
		FROM session
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY id
	`
	rows, err := r.db.QueryContext(c, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []MerchantSessionRow
	for rows.Next() {
		var row MerchantSessionRow
		if err := rows.Scan(&row.ID, &row.NodeName, &row.Status, &row.TotalPrice); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// ListReadySessions возвращает список сессий, готовых к аренде: deleted_at IS NULL и status = 'ready'.
func (r *Repository) ListReadySessions(ctx context.Context) ([]ReadySession, error) {
	c, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	query := `
		SELECT id, user_id FROM session
		WHERE deleted_at IS NULL AND LOWER(TRIM(status)) = $1
		ORDER BY id
	`
	rows, err := r.db.QueryContext(c, query, "ready")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ReadySession
	for rows.Next() {
		var s ReadySession
		if err := rows.Scan(&s.ID, &s.UserID); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// ListReadySessionsWithDetails возвращает сессии с полями session и деталями (железо, цены).
func (r *Repository) ListReadySessionsWithDetails(ctx context.Context) ([]ReadySessionDetails, error) {
	c, cancel := context.WithTimeout(ctx, queryTimeout*3)
	defer cancel()

	query := `
		SELECT id, user_id, node_name, total_price, total_ram, available_ram, price_ram, load_speed, upload_speed, ping, price_internet
		FROM session
		WHERE deleted_at IS NULL AND LOWER(TRIM(status)) = $1
		ORDER BY id
	`
	rows, err := r.db.QueryContext(c, query, "ready")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []ReadySessionDetails
	for rows.Next() {
		var s ReadySessionDetails
		if err := rows.Scan(&s.ID, &s.UserID, &s.NodeName, &s.TotalPrice, &s.TotalRAM, &s.AvailableRAM, &s.PriceRAM, &s.LoadSpeed, &s.UploadSpeed, &s.Ping, &s.PriceInternet); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(sessions) == 0 {
		return sessions, nil
	}

	ids := make([]string, 0, len(sessions))
	for i := range sessions {
		ids = append(ids, sessions[i].ID)
	}

	gpuMap, _ := r.listSessionGPUs(c, ids)
	cpuMap, _ := r.listSessionCPUs(c, ids)
	storageMap, _ := r.listSessionStorages(c, ids)
	templateMap, _ := r.listSessionTemplates(c, ids)

	for i := range sessions {
		if g := gpuMap[sessions[i].ID]; g != nil {
			sessions[i].GPUs = g
		} else {
			sessions[i].GPUs = []SessionGPU{}
		}
		if c := cpuMap[sessions[i].ID]; c != nil {
			sessions[i].CPUs = c
		} else {
			sessions[i].CPUs = []SessionCPU{}
		}
		if s := storageMap[sessions[i].ID]; s != nil {
			sessions[i].Storages = s
		} else {
			sessions[i].Storages = []SessionStorage{}
		}
		if t := templateMap[sessions[i].ID]; t != nil {
			sessions[i].Templates = t
		} else {
			sessions[i].Templates = []string{}
		}
	}
	return sessions, nil
}

// GetSessionDetailsByID возвращает детали сессии по id (любой статус, кроме удалённых). Для активной аренды клиента.
func (r *Repository) GetSessionDetailsByID(ctx context.Context, sessionID string) (*ReadySessionDetails, error) {
	if sessionID == "" {
		return nil, nil
	}
	c, cancel := context.WithTimeout(ctx, queryTimeout*2)
	defer cancel()

	query := `
		SELECT id, user_id, node_name, total_price, total_ram, available_ram, price_ram, load_speed, upload_speed, ping, price_internet
		FROM session
		WHERE deleted_at IS NULL AND id = $1
	`
	var s ReadySessionDetails
	err := r.db.QueryRowContext(c, query, sessionID).Scan(
		&s.ID, &s.UserID, &s.NodeName, &s.TotalPrice, &s.TotalRAM, &s.AvailableRAM, &s.PriceRAM,
		&s.LoadSpeed, &s.UploadSpeed, &s.Ping, &s.PriceInternet,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	ids := []string{sessionID}
	gpuMap, _ := r.listSessionGPUs(c, ids)
	cpuMap, _ := r.listSessionCPUs(c, ids)
	storageMap, _ := r.listSessionStorages(c, ids)
	templateMap, _ := r.listSessionTemplates(c, ids)

	if g := gpuMap[s.ID]; g != nil {
		s.GPUs = g
	} else {
		s.GPUs = []SessionGPU{}
	}
	if c := cpuMap[s.ID]; c != nil {
		s.CPUs = c
	} else {
		s.CPUs = []SessionCPU{}
	}
	if st := storageMap[s.ID]; st != nil {
		s.Storages = st
	} else {
		s.Storages = []SessionStorage{}
	}
	if t := templateMap[s.ID]; t != nil {
		s.Templates = t
	} else {
		s.Templates = []string{}
	}
	return &s, nil
}

func (r *Repository) listSessionTemplates(ctx context.Context, sessionIDs []string) (map[string][]string, error) {
	if len(sessionIDs) == 0 {
		return nil, nil
	}
	query := `SELECT session_id, template_id FROM prepull WHERE session_id = ANY($1::text[]) ORDER BY session_id, template_id`
	rows, err := r.db.QueryContext(ctx, query, pq.Array(sessionIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string][]string)
	for rows.Next() {
		var sid, tid string
		if err := rows.Scan(&sid, &tid); err != nil {
			return nil, err
		}
		out[sid] = append(out[sid], tid)
	}
	return out, rows.Err()
}

// ListAllTemplateIDs возвращает все уникальные template_id из prepull (доступные темплейты для любой сессии).
func (r *Repository) ListAllTemplateIDs(ctx context.Context) ([]string, error) {
	c, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	query := `SELECT DISTINCT template_id FROM prepull ORDER BY template_id`
	rows, err := r.db.QueryContext(c, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var tid string
		if err := rows.Scan(&tid); err != nil {
			return nil, err
		}
		out = append(out, tid)
	}
	return out, rows.Err()
}

func (r *Repository) listSessionGPUs(ctx context.Context, sessionIDs []string) (map[string][]SessionGPU, error) {
	if len(sessionIDs) == 0 {
		return nil, nil
	}
	// placeholder for IN clause
	query := `SELECT session_id, name, total_vram, available_vram, dlperf, price FROM gpu WHERE session_id = ANY($1::text[]) ORDER BY session_id`
	rows, err := r.db.QueryContext(ctx, query, pq.Array(sessionIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string][]SessionGPU)
	for rows.Next() {
		var sid string
		var g SessionGPU
		if err := rows.Scan(&sid, &g.Name, &g.TotalVRAM, &g.AvailableVRAM, &g.Dlperf, &g.Price); err != nil {
			return nil, err
		}
		out[sid] = append(out[sid], g)
	}
	return out, rows.Err()
}

func (r *Repository) listSessionCPUs(ctx context.Context, sessionIDs []string) (map[string][]SessionCPU, error) {
	if len(sessionIDs) == 0 {
		return nil, nil
	}
	query := `SELECT session_id, name, total, available, price FROM cpu WHERE session_id = ANY($1::text[]) ORDER BY session_id`
	rows, err := r.db.QueryContext(ctx, query, pq.Array(sessionIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string][]SessionCPU)
	for rows.Next() {
		var sid string
		var c SessionCPU
		if err := rows.Scan(&sid, &c.Name, &c.Total, &c.Available, &c.Price); err != nil {
			return nil, err
		}
		out[sid] = append(out[sid], c)
	}
	return out, rows.Err()
}

func (r *Repository) listSessionStorages(ctx context.Context, sessionIDs []string) (map[string][]SessionStorage, error) {
	if len(sessionIDs) == 0 {
		return nil, nil
	}
	query := `SELECT session_id, type, name, total, available, bandwidth, price FROM storage WHERE session_id = ANY($1::text[]) ORDER BY session_id`
	rows, err := r.db.QueryContext(ctx, query, pq.Array(sessionIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string][]SessionStorage)
	for rows.Next() {
		var sid string
		var s SessionStorage
		if err := rows.Scan(&sid, &s.Type, &s.Name, &s.Total, &s.Available, &s.Bandwidth, &s.Price); err != nil {
			return nil, err
		}
		out[sid] = append(out[sid], s)
	}
	return out, rows.Err()
}

func findMaxSimilarity(targetGpuName string, gpuDictList []category.GPUDict) string {

	closestID := ""
	minDistance := math.MaxInt32

	for _, dict := range gpuDictList {

		distance := levenshtein.DistanceForStrings([]rune(targetGpuName), []rune(dict.Name), levenshtein.DefaultOptions)
		if distance < minDistance {
			minDistance = distance
			closestID = dict.ID
		}
	}

	return closestID
}
