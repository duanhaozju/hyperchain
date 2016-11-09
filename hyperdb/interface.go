//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hyperdb

type Database interface {
	Put(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
	Delete(key []byte) error
	Close()
	NewBatch() Batch
}

type Batch interface {
	Put(key, value []byte) error
	Write() error
}
