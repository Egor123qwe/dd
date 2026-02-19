package rent

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/domain/model"
	"gitlab.roy9.ru/roy9/backend/core/healthcheckservice/internal/storage/db/repo"
)

const (
	queryTimeot = 7 * time.Second
)

type rentRepo struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) repo.Rent {
	return rentRepo{
		db: db,
	}
}

func (r rentRepo) Rent(ctx context.Context, requestID string) (model.Rent, error) {
	var rent model.Rent

	c, cancel := context.WithTimeout(ctx, queryTimeot)

	defer cancel()

	query := `
	    SELECT id, session_id, status, created_at, client_id, deleted_at
        FROM rent
		WHERE id = $1`

	err := r.db.GetContext(c, &rent, query, requestID)

	if err != nil {
		return rent, err
	}

	if rent == (model.Rent{}) {
		return rent, repo.ErrRentNotFound
	}

	return rent, nil
}

// ActiveRentsByClientID возвращает активные аренды клиента (deleted_at IS NULL, статус любой: pending или started).
func (r rentRepo) ActiveRentsByClientID(ctx context.Context, clientID string) ([]model.Rent, error) {
	c, cancel := context.WithTimeout(ctx, queryTimeot)
	defer cancel()

	query := `
	    SELECT id, session_id, status, created_at, client_id
        FROM rent
        WHERE client_id = $1 AND deleted_at IS NULL
        ORDER BY created_at DESC`
	var list []model.Rent
	err := r.db.SelectContext(c, &list, query, clientID)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// ActiveRentByMerchantSession возвращает активную аренду (deleted_at IS NULL) по session_id и user_id мерчанта (владельца сессии).
func (r rentRepo) ActiveRentByMerchantSession(ctx context.Context, sessionID, userID string) (model.Rent, error) {
	c, cancel := context.WithTimeout(ctx, queryTimeot)
	defer cancel()

	query := `
		SELECT r.id, r.session_id, r.status, r.created_at, r.client_id
		FROM rent r
		INNER JOIN session s ON s.id = r.session_id AND s.user_id = $2
		WHERE r.session_id = $1 AND r.deleted_at IS NULL
		ORDER BY r.created_at DESC
		LIMIT 1`
	var rent model.Rent
	err := r.db.GetContext(c, &rent, query, sessionID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Rent{}, repo.ErrRentNotFound
		}
		return model.Rent{}, err
	}
	return rent, nil
}

func (r rentRepo) MerchantUserID(ctx context.Context, sessionID string) (string, error) {
	var userID string

	c, cancel := context.WithTimeout(ctx, queryTimeot)

	defer cancel()

	query := `
        SELECT user_id
        FROM session
        WHERE id = $1`

	err := r.db.GetContext(c, &userID, query, sessionID)

	if err != nil {
		return "", err
	}

	if userID == "" {
		return userID, repo.ErrSessionDoesNotExist
	}

	return userID, nil
}

func (r rentRepo) MerchantExist(ctx context.Context, sessionID, userID string) (bool, error) {
	var exist bool = false

	ctx, cancel := context.WithTimeout(ctx, queryTimeot)
	defer cancel()

	query := `SELECT 1 
			FROM session
			WHERE user_id = $1 and id = $2 and deleted_at IS NULL
			`

	err := r.db.GetContext(ctx, &exist, query, userID, sessionID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}

	return exist, nil
}

func (r rentRepo) MerchantSessions(ctx context.Context, userID string) ([]string, error) {
	var sessions []string

	c, cancel := context.WithTimeout(ctx, queryTimeot)

	defer cancel()

	query := `
        SELECT id
        FROM session
        WHERE user_id = $1 AND deleted_at IS NULL
		`

	err := r.db.SelectContext(c, &sessions, query, userID)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return sessions, err
	}

	if len(sessions) == 0 {
		return sessions, repo.ErrSessionDoesNotExist
	}

	return sessions, nil
}

func (r rentRepo) TemplateIDForClient(ctx context.Context, rentID string) (string, error) {
	var (
		templateID string
	)

	c, cancel := context.WithTimeout(ctx, queryTimeot)

	defer cancel()

	query := `
		SELECT template_id
		FROM template_settings
		JOIN rent ON rent.settings_id = template_settings.id
		WHERE rent.id = $1
		`

	err := r.db.GetContext(c, &templateID, query, rentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", repo.ErrPrepullDoesNotExist
		}

		return "", err
	}

	return templateID, nil
}

func (r rentRepo) PaidFlagForClient(ctx context.Context, rentID string) (bool, error) {
	var (
		paid bool
	)

	c, cancel := context.WithTimeout(ctx, queryTimeot)

	defer cancel()

	query := `SELECT EXISTS 
				(SELECT 1 
					FROM transaction
					WHERE rent_id = $1 AND status <> 'inited'
				)
	`

	err := r.db.GetContext(c, &paid, query, rentID)
	if err != nil {
		return paid, err
	}

	return paid, nil
}

func (r rentRepo) Session(ctx context.Context, sessionID, userID string) (model.Session, error) {
	query := SessionQuery

	c, cancel := context.WithTimeout(ctx, queryTimeot)
	defer cancel()

	rows, err := r.db.QueryxContext(c, query, sessionID, userID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Session{}, repo.ErrSessionDoesNotExist
		}

		return model.Session{}, err
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
			&session.UploadSpeed, &session.Ping, &session.PriceInternet, &session.TotalPrice, &session.CreatedAt,
			&gpu.ID, &gpu.Name, &gpu.AvailableVRAM, &gpu.UsedVRAM, &gpu.TotalVRAM, &gpu.Dlperf, &gpu.Price,
			&gpu_dict.ID, &gpu_dict.Name, &gpu_dict.TotalVRAM,
			&cpu.ID, &cpu.Name, &cpu.Total, &cpu.Available, &cpu.Price,
			&storage.ID, &storage.Name, &storage.Type, &storage.Total, &storage.Available, &storage.Used, &storage.Bandwidth, &storage.Price,
			&prepoolID, &prepoolTemplateID,
		)

		if err != nil {
			return model.Session{}, err
		}

		if _, exists := sessionsMap[session.ID]; !exists {
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

	return *sessionsMap[sessionID], nil
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
