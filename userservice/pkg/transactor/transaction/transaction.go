package transaction

import (
	"context"
	"sync/atomic"

	"github.com/jmoiron/sqlx"
)

type Tx interface {
	Tx() *sqlx.Tx
	Do(ctx context.Context, fn func(ctx context.Context) error) error
	Commit() error
	Rollback() error
}

type transaction struct {
	tx *sqlx.Tx

	cnt atomic.Int32
}

func New(tx *sqlx.Tx) Tx {
	transaction := &transaction{
		tx: tx,

		cnt: atomic.Int32{},
	}

	return transaction
}

func (tx *transaction) Tx() *sqlx.Tx {
	return tx.tx
}

func (tx *transaction) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	tx.cnt.Add(1)

	return fn(ctx)
}

func (tx *transaction) Commit() error {
	tx.cnt.Add(-1)

	if tx.isActive() {
		return nil
	}

	return tx.tx.Commit()
}

func (tx *transaction) Rollback() error {
	tx.cnt.Add(-1)

	if tx.isActive() {
		return nil
	}

	return tx.tx.Rollback()
}

func (tx *transaction) isActive() bool {
	return tx.cnt.Load() > 0
}
