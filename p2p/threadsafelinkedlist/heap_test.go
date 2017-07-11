package threadsafelinkedlist

import "testing"

func TestHeap_Walk(t *testing.T) {
	items := []*heapItem{
		{
			priority:1,
			value:1,
		},
		{
			priority:3,
			value:3,
		},
		{
			priority:5,
			value:5,
		},
		{
			priority:2,
			value:2,
		},
	}
	heap := NewHeap(items)
	//heap.Sort()
	heap.Walk()
	heap.Push(4,4)
	//heap.Sort()
	heap.Walk()
}