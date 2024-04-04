//go:build !dev

package store

import (
	"context"
	"database/sql"
	"time"
)

func openDB() *Queries {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		panic(err)
	}
	_, err = db.ExecContext(ctx, `PRAGMA foreign_keys = true;`)
	if err != nil {
		panic(err)
	}
	_, err = db.ExecContext(ctx, schema)
	if err != nil {
		panic(err)
	}
	return New(db)
}
