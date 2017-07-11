package database

type Database interface {
	Get(key []byte) ([]byte, error)
	Close() error
	NewIterator(prefix []byte) Iterator
}

type Iterator interface {
	Key() []byte
	Value() []byte
	Seek(key []byte) bool
	Next() bool
	Release()
	Error() error
}
