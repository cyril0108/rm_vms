package repository

import (
	"errors"

	sqlite "modernc.org/sqlite"
)

// Define the modernc SQLite extended error codes as named constants
// to eliminate "magic numbers" in the codebase.
const (
	errCodeConstraintPrimaryKey = 1555
	errCodeConstraintUnique     = 2067
)

// isUniqueConstraintViolation checks if the provided error is a 
// modernc.org/sqlite unique constraint violation.
func isUniqueConstraintViolation(err error) bool {
	if err == nil {
		return false
	}

	var sqliteErr *sqlite.Error
	if errors.As(err, &sqliteErr) {
		code := sqliteErr.Code()
		return code == errCodeConstraintUnique || code == errCodeConstraintPrimaryKey
	}

	return false
}

func IsUniqueConstraintViolation(err error) bool {
	return isUniqueConstraintViolation(err)
}