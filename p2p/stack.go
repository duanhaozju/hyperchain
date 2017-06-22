package p2p

type Stack struct {
	data []interface{}
}

func (s *Stack) Push(x interface{}) {
	s.data = append(s.data, x)
}

func (s *Stack) Pop() interface{} {
	x := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]
	return x
}

func (s *Stack) UnShift(x interface{}) {
	s.data = append([]interface{}{x}, s.data...)
}

func (s *Stack) Length() int {
	return len(s.data)
}
