package server

import "sync/atomic"

type ID struct {
	id uint64
}

func NewId(id uint64) *ID {
	return &ID{
		id:id,
	}
}

func (id *ID) Inc()  {
	atomic.AddUint64(&id.id, 1)
}

func (id *ID) Value() uint64 {
	return atomic.LoadUint64(&id.id)
}

func (id *ID) IncAndGet() uint64 {
	id.Inc()
	return id.Value()
}