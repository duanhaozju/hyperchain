//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package db

type Database interface {
	Put(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
	Delete(key []byte) error
	Close()
	NewBatch() Batch
	NewIterator(prefix []byte) Iterator
	MakeSnapshot(string, []string) error
	Namespace() string
	Scan(begin, end []byte) Iterator
}

type Batch interface {
	Put(key, value []byte) error
	Delete(key []byte) error
	Write() error
	Len() int
}

type Iterator interface {
	Key() []byte
	Value() []byte
	Seek(key []byte) bool
	Next() bool
	Release()
	Error() error
}
