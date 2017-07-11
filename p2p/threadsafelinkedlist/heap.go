package threadsafelinkedlist

import (
	"sync"
	"fmt"
)

// Heap
// for node element i
// left child is 2 * i
// right child 2 * i + 1
type Heap struct{
	rwlock *sync.RWMutex
	heap []*heapItem
}

type heapItem struct {
	value interface{}
	priority int// The priority of the item in the queue.
		    // The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

func NewHeap(list []*heapItem)*Heap{
	heap := new(Heap)
	heap.init(list)
	return heap
}



// i0 is the current adjust idx
// j is the heap's last element idx
func (h *Heap)down(i0,n int){
	i := i0
	j := i0 * 2 // left child
	// if left child is exist
	for j <=  n{
		// if  right child is exist and right child's priority large than left child priority
		if(j+1 <= n && h.heap[j].priority > h.heap[j-1].priority){
			// let left index = right child index
			j = j + 1
		}
		if(h.heap[j-1].priority > h.heap[i -1].priority){
			h.swap(j,i)
			i = j
			j = i * 2
		}else{
			break
		}
	}
}

func(h *Heap)init(items []*heapItem){
	h.heap = items
	n := len(items)
	for i:=n/2;i>=1;i--{
		h.down(i,n)
	}
}


func (h *Heap)swap(i,j int){
	temp := h.heap[i-1]
	h.heap[i-1] = h.heap[j-1]
	h.heap[j-1] =  temp
}

//time complex is O(logn)
func (h *Heap)Pop()interface{}{
	ret := h.heap[0]
	h.heap[0] = h.heap[len(h.heap) - 1]
	h.down(1,len(h.heap))
	return ret.value
}

func(h *Heap)Push(val interface{},pri int){
	hi := &heapItem{
		value:val,
		priority:pri,
	}
	h.heap = append(h.heap,hi)
	h.up(1,len(h.heap))
}

//time complex is O(logn)
func(h *Heap)up(i0,n int){
	i := n
	j := i/2 //parent (right node (2 * i + 1) /2  == 2 * i /2)
	for j > i0{
		if h.heap[j - 1].priority < h.heap[i - 1].priority{
			h.swap(i,j)
			i = j
			j = i/2
		}else{
			break
		}
	}
}

func(h *Heap)Sort(){
	for i := len(h.heap);i>1;i--{
		 h.swap(1,i)
		h.down(1,i-1)
	}
}

func(h *Heap)Walk(){
	for i :=0; i < len(h.heap);i++{
		fmt.Print(h.heap[i].priority,",")
	}
	fmt.Println()
}
