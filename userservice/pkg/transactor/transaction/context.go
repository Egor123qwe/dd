package transaction

import (
	"context"
	"github.com/jmoiron/sqlx"
)

type txContextKey struct{}

func WrapToContext(ctx context.Context, tx Tx) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

func FromContext(ctx context.Context) Tx {
	if tx, ok := ctx.Value(txContextKey{}).(Tx); ok {
		return tx
	}

	return nil
}

// SelectExecutor - try to find transaction in context or return db
func SelectExecutor(ctx context.Context, db *sqlx.DB) sqlx.ExtContext {
	if tx := FromContext(ctx); tx != nil {
		return tx.Tx()
	}

	return db
}
