package transactor

import (
	"context"
	"github.com/Interpuls/ifc2-service-farm/pkg/transactor/transaction"
	"github.com/jmoiron/sqlx"
)

type Transactor interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

func New(db *sqlx.DB) Transactor {
	return transactor{db}
}

type transactor struct {
	db *sqlx.DB
}

func (t transactor) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := t.getTx(ctx)
	if err != nil {
		return err
	}

	err = tx.Do(transaction.WrapToContext(ctx, tx), fn)
	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			return txErr
		}

		return err
	}

	return tx.Commit()
}

func (t transactor) getTx(ctx context.Context) (transaction.Tx, error) {
	tx := transaction.FromContext(ctx)
	if tx != nil {
		return tx, nil
	}

	return t.newTx(ctx)
}

func (t transactor) newTx(ctx context.Context) (transaction.Tx, error) {
	tx, err := t.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return transaction.New(tx), nil
}
