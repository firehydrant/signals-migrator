package store

import (
	"context"
	"database/sql"
	_ "embed"
	"time"
)

//go:embed schema.sql
var schema string

type storeCtxKey string

const (
	queryContextKey = storeCtxKey("fh-signals-migrator-store")
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
	s := NewMemoryStore()

	pragmaCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := s.conn.ExecContext(pragmaCtx, schema); err != nil {
		panic(err)
	}

	return context.WithValue(ctx, queryContextKey, s)
}

func WithContextAndDSN(ctx context.Context, dsn string) context.Context {
	s := NewStore(dsn)

	pragmaCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := s.conn.ExecContext(pragmaCtx, schema); err != nil {
		panic(err)
	}

	return context.WithValue(ctx, queryContextKey, s)
}

func FromContext(ctx context.Context) *Store {
	return ctx.Value(queryContextKey).(*Store)
}

func UseQueries(ctx context.Context) *Queries {
	return New(FromContext(ctx))
}
