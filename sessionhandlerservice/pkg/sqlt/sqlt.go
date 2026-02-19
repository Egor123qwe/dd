package sqlt

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/pkg/sqlt/transaction"
	"gitlab.roy9.ru/roy9/backend/statemachine/sessionhandlerservice/pkg/sqlt/transactor"
)

var (
	ErrNilDBAndTx = errors.New("db and tx are nil")

	ErrNotTx            = errors.New("not a transaction")
	ErrTxAlreadyStarted = errors.New("transaction already started")
)

type DB struct {
	db *sqlx.DB
	tx *sqlx.Tx
}

func NewDB(db *sqlx.DB) *DB {
	return &DB{
		db: db,
	}
}

func NewTX(tx *sqlx.Tx) (*DB, transaction.Service) {
	return &DB{tx: tx}, transactor.New(tx)
}

// BeginTxx returns a transaction and error ErrTxAlreadyStarted if a transaction is already started
func (d *DB) BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {
	if d.db != nil {
		return d.db.BeginTxx(ctx, opts)
	}

	return d.tx, ErrTxAlreadyStarted
}

func (d *DB) Rollback() error {
	if d.tx != nil {
		return d.tx.Rollback()
	}

	return ErrNotTx
}

func (d *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	switch {
	case d.tx != nil:
		return d.tx.QueryRowContext(ctx, query, args...)

	case d.db != nil:
		return d.db.QueryRowContext(ctx, query, args...)
	}

	panic(ErrNilDBAndTx)
}

func (d *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	switch {
	case d.tx != nil:
		return d.tx.QueryContext(ctx, query, args...)

	case d.db != nil:
		return d.db.QueryContext(ctx, query, args...)
	}

	return nil, ErrNilDBAndTx
}

func (d *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	switch {
	case d.tx != nil:
		return d.tx.Exec(query, args...)

	case d.db != nil:
		return d.db.Exec(query, args...)
	}

	return nil, ErrNilDBAndTx
}

func (d *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	switch {
	case d.tx != nil:
		return d.tx.Query(query, args...)

	case d.db != nil:
		return d.db.Query(query, args...)
	}

	return nil, ErrNilDBAndTx
}

func (d *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	switch {
	case d.tx != nil:
		return d.tx.ExecContext(ctx, query, args...)

	case d.db != nil:
		return d.db.ExecContext(ctx, query, args...)
	}

	return nil, ErrNilDBAndTx
}
