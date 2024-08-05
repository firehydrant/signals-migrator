package store

import (
	"errors"

	"modernc.org/sqlite"
	sqlitelib "modernc.org/sqlite/lib"
)

type SQLError sqlite.Error

func AsSQLError(err error) (*SQLError, bool) {
	if err == nil {
		return nil, false
	}
	var e *sqlite.Error
	if errors.As(err, &e) {
		return (*SQLError)(e), true
	}
	return nil, false
}

func (e *SQLError) Error() string {
	return (*sqlite.Error)(e).Error()
}

func (e *SQLError) Code() int {
	return (*sqlite.Error)(e).Code()
}

func (e *SQLError) IsForeignKeyConstraint() bool {
	return e.Code() == sqlitelib.SQLITE_CONSTRAINT_FOREIGNKEY
}
