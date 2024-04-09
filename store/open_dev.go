//go:build dev

package store

import (
	"context"
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/fatih/color"
)

func NewStore() *Store {
	f := filepath.Join(os.TempDir(), "signals-migrator.db")
	log.Printf("using db file: %s", f)

	db, err := sql.Open("sqlite", f)
	if err != nil {
		panic(err)
	}
	return &Store{conn: db}
}

func (s *Store) log(t time.Duration, queryStr string) {
	qInfo := strings.SplitN(queryStr, "\n", 2)
	name := strings.TrimSpace(qInfo[0])
	query := ""
	if len(qInfo) > 1 {
		query = "  " + strings.ReplaceAll(qInfo[1], "\n", "\n  ")
	}
	log.Printf(
		"%s %s\n%s",
		color.HiCyanString("%s", t.String()),
		color.HiMagentaString("%s", name),
		color.WhiteString("%s", query),
	)
}

func (s *Store) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	t := time.Now()
	defer func() { s.log(time.Since(t), query) }()
	return s.conn.ExecContext(ctx, query, args...)
}

func (s *Store) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	t := time.Now()
	defer func() { s.log(time.Since(t), query) }()
	return s.conn.PrepareContext(ctx, query)
}

func (s *Store) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	t := time.Now()
	defer func() { s.log(time.Since(t), query) }()
	return s.conn.QueryContext(ctx, query, args...)
}

func (s *Store) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	t := time.Now()
	defer func() { s.log(time.Since(t), query) }()
	return s.conn.QueryRowContext(ctx, query, args...)
}
