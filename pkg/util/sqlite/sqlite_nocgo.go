//go:build !cgo

package sqlite

import (
	"errors"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

//
// FIXME (@zserge)
//
// This non-CGo "implementation" is merely a stub to make Grafana compile without CGo.
// Any attempts to actually use this driver are likely to fail at runtime in the most brutal ways.
//

type Driver = sqlite.Driver

const DriverName = "sqlite"

// The errors below are used in tests to simulate specific SQLite errors. It's a temporary solution
// until we rewrite the tests not to depend on the sqlite3 package internals directly.
// Note: Since modernc.org/sqlite driver does not expose error codes like sqlite3, we cannot use the same approach.
var (
	TestErrUniqueConstraintViolation = errors.New("unique constraint violation (simulated)")
	TestErrBusy                      = errors.New("database is busy (simulated)")
	TestErrLocked                    = errors.New("database is locked (simulated)")
)

func IsBusyOrLocked(err error) bool {
	var sqliteErr sqlite.Error
	if errors.As(err, &sqliteErr) {
		// Code is 32-bit number, low 8 bits are the SQLite error code, high 24 bits are extended code.
		code := sqliteErr.Code() & 0xff
		return code == sqlite3.SQLITE_BUSY || code == sqlite3.SQLITE_LOCKED
	}
	if errors.Is(err, TestErrBusy) || errors.Is(err, TestErrLocked) {
		return true
	}
	return false
}
func IsUniqueConstraintViolation(err error) bool {
	var sqliteErr sqlite.Error
	if errors.As(err, &sqliteErr) {
		// These constants are extended codes combined with primary code, so we can check them directly.
		return sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_PRIMARY_KEY || sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE
	}
}

func ErrorMessage(err error) string {
	var sqliteErr sqlite.Error
	if errors.As(err, &sqliteErr) {
		return sqliteErr.Error()
	}
	return ""
}
