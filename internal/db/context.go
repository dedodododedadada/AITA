package db

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type txKey struct{}

var activeTxKey = txKey{}

func injectTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, activeTxKey, tx)
}

func extractTx(ctx context.Context) *sqlx.Tx {
	val := ctx.Value(activeTxKey)
	if tx, ok := val.(*sqlx.Tx); ok {
		return tx
	}
	return nil
}

