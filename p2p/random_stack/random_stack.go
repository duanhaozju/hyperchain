package random_stack

import (
	"sync"
	"math/rand"
	"fmt"
)

type Stack struct {
	rwLock *sync.Mutex
	list   []interface{}
	size   int
}

func NewStack() *Stack {
	return &Stack{
		size: 0,
		list:make([]interface{},0),
		rwLock:new(sync.Mutex),
	}
}

// Push adds on an item on the top of the Stack
func (s *Stack) Push(item interface{}) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()
	s.list = append(s.list,item)
	s.size +=1
}

func (s *Stack)Empty() bool{
	s.rwLock.Lock()
	defer s.rwLock.Unlock()
	return s.size == 0
}

// Pop removes and returns the item on the top of the Stack
func (s *Stack) Pop() interface{} {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()
	if s.size == 1{
		ret := s.list[0]
		s.list = make([]interface{},0)
		return ret
	}
	ret := s.list[s.size -1]
	s.list = append(make([]interface{},0),s.list[:s.size-1]...)
	s.size -=1
	return ret
}

func (s *Stack) RandomPop() interface{} {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()
	idx := rand.Intn(s.size)
	// [0,1,2]
	// s.cap = 3
	// idx == 2
	//ret == 2
	ret := s.list[idx]
	//temp list == [0,1]
	templist := make([]interface{}, 0)
	if idx >= 1{
		templist = append(make([]interface{}, 0), s.list[:idx]...)
	}
	if idx < s.size - 1{
		// [0,1,2]
		// s.cap = 3
		// idx == 1
		//ret == 1
		//temp list1 == [0]
		//temp list2 = [0] append s.list[2:2]
		templist = append(templist,s.list[idx+1:s.size]...)
	}
	s.list = templist
	s.size = len(s.list)
	return ret
}

func(s *Stack)Walk(){
	s.rwLock.Lock()
	defer s.rwLock.Unlock()
	for idx,item := range s.list{
		fmt.Println(idx,item)
	}
}
