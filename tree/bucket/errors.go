package bucket

import "errors"

var (
	ErrNotFound    = errors.New("leveldb: not found")
)