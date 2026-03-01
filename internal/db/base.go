package db

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type Execer interface {
	sqlx.ExtContext
	sqlx.ExecerContext
	sqlx.PreparerContext
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	GetContext(ctx context.Context, dest interface{}, query string, args ...any) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...any) error
	
}
// func (e Execer) SelectContext(ctx context.Context, follow *[]*models.Follow, query string, followerID int64) any {
// 	panic("unimplemented")
// }

type BaseStore struct {
	database *sqlx.DB
}

func (b *BaseStore) conn(ctx context.Context) Execer {
	if tx := extractTx(ctx); tx != nil {
		return tx
	}
	return b.database
}

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

type sqlTransactor struct {
	db *sqlx.DB
}

func NewTransactor(db *sqlx.DB) *sqlTransactor {
	return &sqlTransactor{db: db}
}

func (t *sqlTransactor) Exec(ctx context.Context, fn func(ctx context.Context) error) error {
    tx, err := t.db.BeginTxx(ctx, nil)
    if err != nil {
        return err
    }

    txCtx := injectTx(ctx, tx)

    if err := fn(txCtx); err != nil {
        tx.Rollback()
        return err
    }

    return tx.Commit()
}