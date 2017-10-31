package db

import "errors"

// Error codes returned by failures.
var (
	// legacy error code
	DB_NOT_FOUND = errors.New("db not found")

	// new error codes
	ErrNotSupport = errors.New("ErrNotSupport")
)
