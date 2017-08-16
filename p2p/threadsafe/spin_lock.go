package threadsafe

import "sync/atomic"

type SpinLock struct {
	state int32
}

func (s *SpinLock)TryLock() bool {
	return atomic.CompareAndSwapInt32(&s.state, 0, 1)
}

func (s *SpinLock) IsLocked() bool {
	return s.state == 1
}

func (s *SpinLock) UnLock() {
	s.state = 0
}
