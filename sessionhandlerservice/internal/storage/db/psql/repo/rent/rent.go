package rent

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	rentModel "gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/model/rent"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/storage/repo/rent"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/util"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/internal/util/parser"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/pkg/sqlt"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/pkg/sqlt/transaction"
)

const (
	requestTimeout = 15 * time.Second
)

var (
	ErrRentNotFound     = errors.New("rent not found")
	ErrMerchantNotFound = errors.New("merchant not found")
)

type repo struct {
	db *sqlt.DB
}

func New(db *sqlt.DB) rent.Rent {
	return &repo{
		db: db,
	}
}

func (r repo) WithTransaction(ctx context.Context) (rent.Rent, transaction.Service, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	db, transactor := sqlt.NewTX(tx)

	return New(db), transactor, nil
}

func (r repo) Get(ctx context.Context, id string) (rentModel.Rent, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	var modelRent rentModel.Rent

	query := `SELECT 
    				id, session_id, client_id, status, created_at
			  FROM rent WHERE id = $1 AND deleted_at IS NULL`

	err := r.db.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(
		&modelRent.ID,
		&modelRent.SessionID,
		&modelRent.ClientId,
		&modelRent.Status,
		&modelRent.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return rentModel.Rent{}, ErrRentNotFound
		}

		return modelRent, err
	}

	return modelRent, nil
}

func (r repo) Create(ctx context.Context, rent rentModel.Rent, configuration rentModel.Settings) (rentModel.Rent, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil && !errors.Is(err, sqlt.ErrTxAlreadyStarted) {
		return rentModel.Rent{}, err
	}

	// if error is not ErrTxAlreadyStarted
	if err == nil {
		defer tx.Commit()
	}

	var templateSettingsID int64

	templateSettingsQuery := `INSERT INTO template_settings 
        							(template_id, title, type, description, short_description,
        							 version, image_name, image_tag, ports, envs, volumes, use_gpu,
        							 login, password)
      		  				  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) RETURNING id`

	err = r.db.QueryRowContext(ctx, templateSettingsQuery,
		configuration.Template.Template.ID,
		configuration.Template.Template.Title,
		configuration.Template.Template.Type,
		configuration.Template.Template.Description,
		configuration.Template.Template.ShortDescription,
		configuration.Template.Template.Version,
		configuration.Template.Template.ImageName,
		configuration.Template.Template.ImageTag,
		pq.StringArray(util.ArrayInitializer(parser.ArrayToJSON(configuration.Template.Template.Ports))),
		pq.StringArray(util.ArrayInitializer(parser.ArrayToJSON(configuration.Template.Template.Envs))),
		pq.StringArray(util.ArrayInitializer(configuration.Template.Template.Volumes)),
		configuration.Template.Template.UseGPU,
		configuration.Template.Authentication.Login,
		configuration.Template.Authentication.Password,
	).Scan(
		&templateSettingsID,
	)

	if err != nil {
		tx.Rollback()

		return rentModel.Rent{}, err
	}

	var networkWaySettingsID int64

	if configuration.Network.Piko != nil {
		pikoSettingsQuery := `INSERT INTO piko_settings 
        							(auth_key, endpoints)
      		  				  VALUES ($1, $2) RETURNING id`

		err = r.db.QueryRowContext(ctx, pikoSettingsQuery,
			configuration.Network.Piko.AuthKey,
			pq.StringArray(util.ArrayInitializer(parser.ArrayToJSON(configuration.Network.Piko.Endpoints))),
		).Scan(
			&networkWaySettingsID,
		)

		if err != nil {
			tx.Rollback()

			return rentModel.Rent{}, err
		}
	}

	if configuration.Network.Tailscale != nil {
		tailscaleSettingsQuery := `INSERT INTO tailscale_settings 
        								(merchant_key, client_key)
      		  				  	   VALUES ($1, $2) RETURNING id`

		err = r.db.QueryRowContext(ctx, tailscaleSettingsQuery,
			configuration.Network.Tailscale.MerchantAuthKey,
			configuration.Network.Tailscale.ClientAuthKey,
		).Scan(
			&networkWaySettingsID,
		)

		if err != nil {
			tx.Rollback()

			return rentModel.Rent{}, err
		}
	}

	var networkSettingsID int64

	networkSettingsQuery := `INSERT INTO network_settings 
        								(network_way_id)
      		  				 VALUES ($1) RETURNING id`

	err = r.db.QueryRowContext(ctx, networkSettingsQuery,
		networkWaySettingsID,
	).Scan(
		&networkSettingsID,
	)

	if err != nil {
		tx.Rollback()

		return rentModel.Rent{}, err
	}

	var rentSettingsID int64

	rentSettingsQuery := `INSERT INTO rent_settings 
        								(mode, template_id, network_id)
      		  				 VALUES ($1, $2, $3) RETURNING id`

	err = r.db.QueryRowContext(ctx, rentSettingsQuery,
		configuration.Mode,
		templateSettingsID,
		networkSettingsID,
	).Scan(
		&rentSettingsID,
	)

	if err != nil {
		tx.Rollback()

		return rentModel.Rent{}, err
	}

	rent.ID = uuid.New().String()

	query := `INSERT INTO rent 
        			(id, session_id, client_id, status, settings_id)
      		  VALUES ($1, $2, $3, $4, $5) RETURNING created_at`

	err = r.db.QueryRowContext(ctx, query,
		rent.ID,
		rent.SessionID,
		rent.ClientId,
		rent.Status,
		rentSettingsID,
	).Scan(
		&rent.CreatedAt,
	)

	if err != nil {
		tx.Rollback()

		return rentModel.Rent{}, err
	}

	return rent, nil
}

func (r repo) ChangeStatus(ctx context.Context, id string, status rentModel.Status) error {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	query := `UPDATE rent SET status = $1 WHERE id = $2`

	res, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRentNotFound
	}

	return nil
}

func (r repo) Stop(ctx context.Context, id string, reason string, cost float64) error {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	query := `UPDATE rent SET status = $1, deleted_at = $2, stop_reason = $3, cost = $4 WHERE id = $5`

	deletedAt := time.Now().UTC()
	costRounded := math.Round(cost*100) / 100

	res, err := r.db.ExecContext(ctx, query,
		rentModel.FinishedRentStatus,
		deletedAt,
		reason,
		costRounded,
		id,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRentNotFound
	}

	return nil
}
