package store

import (
	"context"
	"database/sql"
	_ "embed"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

type storeCtxKey string

const (
	queryContextKey = storeCtxKey("fh-signals-migrator-store")
	txContextKey    = storeCtxKey("fh-signals-migrator-tx")
)

var (
	ErrNoTx = sql.ErrTxDone
)

type Connection interface {
	DBTX

	Begin() (*sql.Tx, error)
	Close() error
}

type Store struct {
	conn Connection
}

func (s *Store) Close() error {
	return s.conn.Close()
}

func WithContext(ctx context.Context) context.Context {
	s := NewStore()

	pragmaCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := s.conn.ExecContext(pragmaCtx, `PRAGMA foreign_keys = true;`); err != nil {
		panic(err)
	}
	if _, err := s.conn.ExecContext(pragmaCtx, schema); err != nil {
		panic(err)
	}

	return context.WithValue(ctx, queryContextKey, s)
}

func FromContext(ctx context.Context) *Store {
	return ctx.Value(queryContextKey).(*Store)
}

func WithTx(ctx context.Context) context.Context {
	s := FromContext(ctx)
	tx, err := s.conn.Begin()
	if err != nil {
		panic(err)
	}
	return context.WithValue(ctx, txContextKey, tx)
}

func CommitTx(ctx context.Context) error {
	txVal := ctx.Value(txContextKey)
	if txVal == nil {
		return ErrNoTx
	}
	tx := txVal.(*sql.Tx)
	return tx.Commit()
}

func RollbackTx(ctx context.Context) error {
	txVal := ctx.Value(txContextKey)
	if txVal == nil {
		return ErrNoTx
	}
	tx := txVal.(*sql.Tx)
	return tx.Rollback()
}

func UseQueries(ctx context.Context) *Queries {
	tx := ctx.Value(txContextKey)
	if tx != nil {
		return New(tx.(*sql.Tx))
	}
	return New(FromContext(ctx))
}
