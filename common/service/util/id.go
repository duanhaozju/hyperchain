package util

import "sync/atomic"

//ID atomic increment id.
type ID struct {
	id uint64 // base id should be unique for different goroutine
}

func NewId(id uint64) *ID {
	return &ID{
		id: id,
	}
}

func (id *ID) Inc() {
	atomic.AddUint64(&id.id, 1)
}

func (id *ID) Value() uint64 {
	return atomic.LoadUint64(&id.id)
}

//IncAndGet increment the id and fetch the value
func (id *ID) IncAndGet() uint64 {
	id.Inc()
	return id.Value()
}
