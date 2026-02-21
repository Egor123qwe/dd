package template

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model/rent"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/repo/template"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/util"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/util/parser"
)

const (
	requestTimeout = 15 * time.Second
)

var (
	ErrTemplateNotFound = errors.New("template not found")
)

type repo struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) template.Template {
	return &repo{
		db: db,
	}
}

func (r repo) Get(ctx context.Context, id string) (rent.Template, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	result := rent.Template{
		ID: id,
	}

	var pqPorts, pqEnvs, pqVolumes pq.ByteaArray

	query := `SELECT 
    			    title, type, description, short_description,
    			    version, container_image_name, container_image_tag,
    			    ports, envs, volumes, use_gpu,
    			    min_cpu, min_ram_bytes, min_storage_bytes, min_volume_storage_bytes
			  FROM templates_template_info WHERE template_id = $1`

	var minVolumeStorage []int64
	err := r.db.QueryRowContext(
		ctx,
		query,
		result.ID,
	).Scan(
		&result.Title, &result.Type, &result.Description, &result.ShortDescription,
		&result.Version, &result.ImageName, &result.ImageTag,
		&pqPorts, &pqEnvs, &pqVolumes, &result.UseGPU,
		&result.MinCPU, &result.MinRAMBytes, &result.MinStorageBytes, pq.Array(&minVolumeStorage),
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return rent.Template{}, ErrTemplateNotFound
		}

		return rent.Template{}, err
	}

	result.MinVolumeStorageBytes = int64SliceToUint64(minVolumeStorage)
	result.Ports = pqArrayToPorts(pqPorts)
	result.Envs = pqArrayToEnvs(pqEnvs)
	result.Volumes = pqArrayToSlice(pqVolumes)

	return result, nil
}

func (r repo) ListAll(ctx context.Context) ([]rent.Template, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	query := `SELECT template_id, title, type, description, short_description,
	    version, container_image_name, container_image_tag, ports, envs, volumes, use_gpu,
	    min_cpu, min_ram_bytes, min_storage_bytes, min_volume_storage_bytes
	  FROM templates_template_info ORDER BY template_id`
	rows, err := r.db.QueryxContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []rent.Template
	for rows.Next() {
		var t rent.Template
		var pqPorts, pqEnvs, pqVolumes pq.ByteaArray
		var minVolumeStorage []int64
		err := rows.Scan(
			&t.ID, &t.Title, &t.Type, &t.Description, &t.ShortDescription,
			&t.Version, &t.ImageName, &t.ImageTag, &pqPorts, &pqEnvs, &pqVolumes, &t.UseGPU,
			&t.MinCPU, &t.MinRAMBytes, &t.MinStorageBytes, pq.Array(&minVolumeStorage),
		)
		if err != nil {
			return nil, err
		}
		t.Ports = pqArrayToPorts(pqPorts)
		t.Envs = pqArrayToEnvs(pqEnvs)
		t.Volumes = pqArrayToSlice(pqVolumes)
		t.MinVolumeStorageBytes = int64SliceToUint64(minVolumeStorage)
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r repo) Create(ctx context.Context, t rent.Template) error {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	query := `INSERT INTO templates_template_info (
		template_id, title, type, description, short_description,
		version, container_image_name, container_image_tag, ports, envs, volumes, use_gpu,
		min_cpu, min_ram_bytes, min_storage_bytes, min_volume_storage_bytes
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`
	_, err := r.db.ExecContext(ctx, query,
		t.ID, t.Title, t.Type, t.Description, t.ShortDescription,
		t.Version, t.ImageName, t.ImageTag,
		pq.StringArray(util.ArrayInitializer(parser.ArrayToJSON(t.Ports))),
		pq.StringArray(util.ArrayInitializer(parser.ArrayToJSON(t.Envs))),
		pq.StringArray(util.ArrayInitializer(t.Volumes)),
		t.UseGPU,
		t.MinCPU, t.MinRAMBytes, t.MinStorageBytes, pq.Array(uint64SliceToInt64(t.MinVolumeStorageBytes)),
	)
	return err
}

func (r repo) Update(ctx context.Context, id string, title, type_, description, shortDescription, version, imageName, imageTag string, useGPU bool, ports []rent.Port, envs []rent.Env, volumes []string, minCPU int32, minRAMBytes, minStorageBytes uint64, minVolumeStorageBytes []uint64) error {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	query := `UPDATE templates_template_info SET
		title = $2, type = $3, description = $4, short_description = $5,
		version = $6, container_image_name = $7, container_image_tag = $8, use_gpu = $9,
		ports = $10, envs = $11, volumes = $12,
		min_cpu = $13, min_ram_bytes = $14, min_storage_bytes = $15, min_volume_storage_bytes = $16
		WHERE template_id = $1`
	_, err := r.db.ExecContext(ctx, query, id, title, type_, description, shortDescription, version, imageName, imageTag, useGPU,
		pq.StringArray(util.ArrayInitializer(parser.ArrayToJSON(ports))),
		pq.StringArray(util.ArrayInitializer(parser.ArrayToJSON(envs))),
		pq.StringArray(util.ArrayInitializer(volumes)),
		minCPU, minRAMBytes, minStorageBytes, pq.Array(uint64SliceToInt64(minVolumeStorageBytes)),
	)
	return err
}

func pqArrayToSlice(pqArray pq.ByteaArray) []string {
	var result []string

	for _, str := range pqArray {
		result = append(result, string(str))
	}

	return result
}

func pqArrayToPorts(pqArray pq.ByteaArray) []rent.Port {
	var result []rent.Port

	for _, data := range pqArrayToSlice(pqArray) {
		var port rent.Port

		if err := json.Unmarshal([]byte(data), &port); err != nil {
			continue
		}

		result = append(result, port)
	}

	return result
}

func pqArrayToEnvs(pqArray pq.ByteaArray) []rent.Env {
	var result []rent.Env

	for _, data := range pqArrayToSlice(pqArray) {
		var e rent.Env

		if err := json.Unmarshal([]byte(data), &e); err != nil {
			continue
		}

		result = append(result, e)
	}

	return result
}

func int64SliceToUint64(s []int64) []uint64 {
	out := make([]uint64, len(s))
	for i, v := range s {
		if v < 0 {
			out[i] = 0
		} else {
			out[i] = uint64(v)
		}
	}
	return out
}

func uint64SliceToInt64(s []uint64) []int64 {
	out := make([]int64, len(s))
	for i, v := range s {
		if v > 1<<63-1 {
			out[i] = 1<<63 - 1
		} else {
			out[i] = int64(v)
		}
	}
	return out
}
