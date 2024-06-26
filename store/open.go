//go:build !dev

package store

import (
	"context"
	"database/sql"

	_ "modernc.org/sqlite"
)

func NewStore(dsn string) *Store {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		panic(err)
	}
	return &Store{conn: db}
}

func NewMemoryStore() *Store {
	return NewStore("file::memory:?cache=shared")
}

func (s *Store) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return s.conn.ExecContext(ctx, query, args...)
}

func (s *Store) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return s.conn.PrepareContext(ctx, query)
}

func (s *Store) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return s.conn.QueryContext(ctx, query, args...)
}

func (s *Store) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return s.conn.QueryRowContext(ctx, query, args...)
}
