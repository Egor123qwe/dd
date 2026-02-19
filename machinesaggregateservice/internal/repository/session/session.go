package session

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sort"
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

func (r Repository) FeedbackList(ctx context.Context, log slog.Logger, sessions []model.Session) ([]model.Session, error) {
	var (
		s         []string
		score     float32
		sessionID string
	)

	c, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	for _, sess := range sessions {
		s = append(s, sess.ID)
	}

	query, args, err := sqlx.In(FeedbackList, s)
	if err != nil {
		return []model.Session{}, fmt.Errorf("FeedbackList : %v", err)
	}

	query = r.db.Rebind(query)

	rows, err := r.db.QueryxContext(c, query, args...)
	if err != nil {
		log.Error(err.Error())
		return []model.Session{}, fmt.Errorf("FeedbackList query: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&sessionID, &score)
		if err != nil {
			return []model.Session{}, fmt.Errorf("FeedbackList : %v", err)
		}

		for i := range sessions {
			if sessions[i].ID == sessionID {
				sessions[i].Score = score
				break
			}
		}
	}

	return sessions, nil
}

func (r Repository) SessionList(ctx context.Context, log slog.Logger, filter model.FilterRepo) ([]model.Session, error) {
	c, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	rows, err := r.queryWithFilter(c, filter, SessionQuery)
	if err != nil {
		log.Error(err.Error())
		return []model.Session{}, fmt.Errorf("SessionList query: %v", err)
	}
	defer rows.Close()

	var sessionsMap = make(map[string]*model.Session)

	for rows.Next() {
		var (
			gpu               model.GPU
			cpu               model.CPU
			storage           model.Storage
			session           model.Session
			prepool           model.PrePull
			prepoolID         sql.NullString
			prepoolTemplateID sql.NullString
			gpu_dict          model.GPUDict
		)

		err := rows.Scan(
			&session.ID, &session.TotalRam, &session.AvailableRam, &session.UsedRam, &session.PriceRam, &session.LoadSpeed,
			&session.UploadSpeed, &session.Ping, &session.PriceInternet, &session.TotalPrice, &session.CreatedAt, &session.Uptime, &session.Reliability,
			&gpu.ID, &gpu.Name, &gpu.AvailableVRAM, &gpu.UsedVRAM, &gpu.TotalVRAM, &gpu.Dlperf, &gpu.Price,
			&gpu_dict.ID, &gpu_dict.Name, &gpu_dict.TotalVRAM,
			&cpu.ID, &cpu.Name, &cpu.Total, &cpu.Available, &cpu.Price,
			&storage.ID, &storage.Name, &storage.Type, &storage.Total, &storage.Available, &storage.Used, &storage.Bandwidth, &storage.Price,
			&prepoolID, &prepoolTemplateID,
			&session.UserId, &session.Avatar,
		)

		if err != nil {
			return []model.Session{}, fmt.Errorf("SessionList: %v", err)
		}

		if _, exists := sessionsMap[session.ID]; !exists {
			session.Reliability = max(0.60, session.Reliability)
			sessionsMap[session.ID] = &session
		}

		s := sessionsMap[session.ID]

		if !containsGPU(s.GPUs, gpu.ID) {
			gpu.GPUDict = gpu_dict
			s.GPUs = append(s.GPUs, gpu)
		}

		if !containsCPU(s.CPUs, cpu.ID) {
			s.CPUs = append(s.CPUs, cpu)
		}

		if !containsPrePull(s.PrePull, prepoolID.String) {
			if prepoolID.Valid && prepoolTemplateID.Valid {
				prepool.ID = prepoolID.String
				prepool.TemplateId = prepoolTemplateID.String

				s.PrePull = append(s.PrePull, prepool)
			}
		}

		if !containsStorage(s.Storage, storage.ID) {
			s.Storage = append(s.Storage, storage)
		}
	}

	var sessions []model.Session
	for _, session := range sessionsMap {
		sessions = append(sessions, *session)
	}

	if rows.Err() != nil {
		log.Error(fmt.Sprintf("%v", rows.Err()))
		return []model.Session{}, fmt.Errorf("SessionList : %v", rows.Err())
	}

	if len(sessions) == 0 {
		return []model.Session{}, model.ErrSessionsNotFound
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].TotalPrice > sessions[j].TotalPrice
	})

	return sessions, nil
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

func containsGPU(gpus []model.GPU, id string) bool {
	for _, gpu := range gpus {
		if gpu.ID == id {
			return true
		}
	}

	return false
}

func containsCPU(cpus []model.CPU, id string) bool {
	for _, cpu := range cpus {
		if cpu.ID == id {
			return true
		}
	}

	return false
}

func containsPrePull(prePull []model.PrePull, id string) bool {
	for _, pp := range prePull {
		if pp.ID == id {
			return true
		}
	}

	return false
}

func containsStorage(storages []model.Storage, id string) bool {
	for _, st := range storages {
		if st.ID == id {
			return true
		}
	}

	return false
}
