package store

import (
	_ "embed"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

var Query = openDB()
