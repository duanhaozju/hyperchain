package threadsafe

import "sync/atomic"

type SpinLock struct {
	state int32
}

func (s *SpinLock) TryLock() bool {
	return atomic.CompareAndSwapInt32(&s.state, 0, 1)
}

func (s *SpinLock) IsLocked() bool {
	return s.state == 1
}

func (s *SpinLock) UnLock() {
	s.state = 0

	// todo 这里应该线程安全， 用下面的代码
	//atomic.CompareAndSwapInt32(&s.state, 1, 0)
}
