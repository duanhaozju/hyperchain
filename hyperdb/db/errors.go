package db

import "errors"

// Error codes returned by failures.
var (
	// legacy error code
	DB_NOT_FOUND = errors.New("db not found")

	// new error codes
	ErrDbNotFound = errors.New("ErrDbNotFound")
	//ErrDbClosed    = errors.New("ErrDbClosed")
	ErrKeyNotFound = errors.New("ErrKeyNotFound")
	ErrNotSupport  = errors.New("ErrNotSupport")
)
