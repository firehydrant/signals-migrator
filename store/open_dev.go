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

	"github.com/fatih/color"
)

type loggedQueries struct {
	q *sql.DB
}

func (q *loggedQueries) log(t time.Duration, queryStr string) {
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

func (q *loggedQueries) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	t := time.Now()
	defer func() { q.log(time.Since(t), query) }()
	return q.q.ExecContext(ctx, query, args...)
}

func (q *loggedQueries) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	t := time.Now()
	defer func() { q.log(time.Since(t), query) }()
	return q.q.PrepareContext(ctx, query)
}

func (q *loggedQueries) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	t := time.Now()
	defer func() { q.log(time.Since(t), query) }()
	return q.q.QueryContext(ctx, query, args...)
}

func (q *loggedQueries) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	t := time.Now()
	defer func() { q.log(time.Since(t), query) }()
	return q.q.QueryRowContext(ctx, query, args...)
}

func openDB() *Queries {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	f := filepath.Join(os.TempDir(), "signals-migrator.db")
	log.Printf("using db file %s", f)

	db, err := sql.Open("sqlite", f)
	if err != nil {
		panic(err)
	}
	dbtx := &loggedQueries{q: db}
	_, err = dbtx.ExecContext(ctx, `PRAGMA foreign_keys = true;`)
	if err != nil {
		panic(err)
	}
	_, err = dbtx.ExecContext(ctx, schema)
	if err != nil {
		panic(err)
	}
	return New(dbtx)
}
