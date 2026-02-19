package rent

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/lib/pq"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model/rent"
)

func (r repo) GetSettings(ctx context.Context, requestID string) (rent.Settings, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	var result rent.Settings

	query := `SELECT 
    			  s.mode, 
    			  ts.template_id, 
    			  ts.type,
    			  ts.title, ts.description, ts.short_description,
    			  ts.version, ts.image_name, ts.image_tag, 
    			  ts.login, ts.password, 
    			  ts.ports, ts.envs, ts.volumes, ts.use_gpu,
    			  ns.network_way_id
			  FROM rent 
			  INNER JOIN rent_settings 	   AS s  ON rent.settings_id = s.id
			  INNER JOIN template_settings AS ts ON s.template_id    = ts.id
			  INNER JOIN network_settings  AS ns ON s.network_id     = ns.id
			  WHERE rent.id = $1 AND rent.deleted_at IS NULL
			  `

	var (
		networkWayID         int64
		ports, envs, volumes pq.ByteaArray
	)

	err := r.db.QueryRowContext(
		ctx,
		query,
		requestID,
	).Scan(
		&result.Mode,

		&result.Template.Template.ID,
		&result.Template.Template.Type,
		&result.Template.Template.Title,
		&result.Template.Template.Description,
		&result.Template.Template.ShortDescription,

		&result.Template.Template.Version,
		&result.Template.Template.ImageName,
		&result.Template.Template.ImageTag,

		&result.Template.Authentication.Login, &result.Template.Authentication.Password,

		&ports, &envs, &volumes, &result.Template.Template.UseGPU,

		&networkWayID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return rent.Settings{}, ErrRentNotFound
		}

		return rent.Settings{}, err
	}

	result.Template.Template.Ports = pqArrayToPorts(ports)
	result.Template.Template.Envs = pqArrayToEnvs(envs)
	result.Template.Template.Volumes = pqArrayToSlice(volumes)

	switch result.Mode {
	case rent.ProxyMode:
		{
			piko := rent.Piko{}
			var pikEndpoints pq.ByteaArray

			query := `SELECT 
    			  			auth_key, endpoints
			  		  FROM piko_settings 
			  	      WHERE id = $1
			  `

			err = r.db.QueryRowContext(
				ctx,
				query,
				networkWayID,
			).Scan(
				&piko.AuthKey,
				&pikEndpoints,
			)

			if err != nil {
				return rent.Settings{}, err
			}

			piko.Endpoints = pqArrayToEndpoints(pikEndpoints)

			result.Network.Piko = &piko
		}

	case rent.P2PMode:
		{
			tailscale := rent.Tailscale{}

			query := `SELECT 
    			  			merchant_key, client_key
			  		  FROM tailscale_settings 
			  	      WHERE id = $1
			  `

			err = r.db.QueryRowContext(
				ctx,
				query,
				networkWayID,
			).Scan(
				&tailscale.MerchantAuthKey,
				&tailscale.ClientAuthKey,
			)

			if err != nil {
				return rent.Settings{}, err
			}

			result.Network.Tailscale = &tailscale
		}
	}

	return result, nil
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

func pqArrayToEndpoints(pqArray pq.ByteaArray) []rent.PikoEndpoint {
	var result []rent.PikoEndpoint

	for _, data := range pqArrayToSlice(pqArray) {
		var endpoint rent.PikoEndpoint

		if err := json.Unmarshal([]byte(data), &endpoint); err != nil {
			continue
		}

		result = append(result, endpoint)
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
