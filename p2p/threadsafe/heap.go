package threadsafe

import (
	"fmt"
	"sync"
)

// Heap
// for node element i
// left child is 2 * i
// right child 2 * i + 1
type Heap struct {
	rwlock *sync.RWMutex
	heap   []*heapItem
}

type heapItem struct {
	value    interface{}
	priority int // The priority of the item in the queue.
}

func NewHeap(items ...WeightItem) *Heap {
	list := make([]*heapItem, 0)
	for _, item := range items {
		tmp := &heapItem{
			value:    item.Value(),
			priority: item.Weight(),
		}
		list = append(list, tmp)
	}
	h := new(Heap)
	h.rwlock = new(sync.RWMutex)
	h.init(list)
	return h
}

//time complex is O(logn)
func (h *Heap) Pop() interface{} {
	h.rwlock.Lock()
	defer h.rwlock.Unlock()

	ret := h.heap[0]
	h.heap[0] = h.heap[len(h.heap)-1]
	h.down(1, len(h.heap))
	return ret.value
}

func (h *Heap) Push(val interface{}, pri int) {
	h.rwlock.Lock()
	defer h.rwlock.Unlock()
	for _, i := range h.heap {
		if i.priority == pri {
			return
		}
	}
	hi := &heapItem{
		value:    val,
		priority: pri,
	}
	h.heap = append(h.heap, hi)
	h.up(1, len(h.heap))
}

func (h *Heap) Walk() {
	h.rwlock.RLock()
	defer h.rwlock.RUnlock()
	for i := 0; i < len(h.heap); i++ {
		fmt.Print(h.heap[i].priority, ",")
	}
}

func (h *Heap) Sort() []interface{} {
	h.rwlock.RLock()
	defer h.rwlock.RUnlock()
	//first get a heap copy
	tempHeap := new(Heap)
	tempHeap.rwlock = new(sync.RWMutex)
	for _, item := range h.heap {
		tempItem := &heapItem{
			value:    item.value,
			priority: item.priority,
		}
		tempHeap.Push(tempItem.value, item.priority)
	}
	for i := len(h.heap); i > 1; i-- {
		tempHeap.swap(1, i)
		tempHeap.down(1, i-1)
	}
	ret := make([]interface{}, 0)
	for _, it := range tempHeap.heap {
		ret = append(ret, it.value)
	}
	return ret
}

func (h *Heap) Remove(priority int) interface{} {
	h.rwlock.Lock()
	defer h.rwlock.Unlock()
	var ret interface{}
	var i int
	for i = 0; i < len(h.heap); i++ {
		if h.heap[i].priority == priority {
			ret = h.heap[i].value
			break
		}
	}
	if i == len(h.heap) {
		return nil
	} else {
		for j := 0; j < len(h.heap); j++ {
			if h.heap[j].priority > priority {
				h.heap[j].priority--
			}
		}
	}
	//adjust the heap
	h.swap(i+1, len(h.heap))
	h.heap = h.heap[:len(h.heap)-1]
	h.down(i+1, len(h.heap))
	return ret
}

func (h *Heap) Duplicate() *Heap {
	h.rwlock.Lock()
	defer h.rwlock.Unlock()
	temph := new(Heap)
	temph.rwlock = new(sync.RWMutex)
	temph.heap = make([]*heapItem, 0)
	for _, item := range h.heap {
		tmpItem := &heapItem{
			value:    item.value,
			priority: item.priority,
		}
		temph.heap = append(temph.heap, tmpItem)
	}
	return temph
}

//time complex is O(logn)
func (h *Heap) up(i0, n int) {
	i := n
	j := i / 2 //parent (right node (2 * i + 1) /2  == 2 * i /2)
	for j >= i0 {
		if h.heap[j-1].priority < h.heap[i-1].priority {
			h.swap(i, j)
			i = j
			j = i / 2
		} else {
			break
		}
	}
}

// i0 is the current adjust idx
// j is the heap's last element idx
func (h *Heap) down(i0, n int) {
	i := i0
	j := i0 * 2 // left child
	// if left child is exist
	for j <= n {
		// if  right child is exist and right child's priority large than left child priority
		if j+1 <= n && h.heap[j].priority > h.heap[j-1].priority {
			// let left index = right child index
			j = j + 1
		}
		if h.heap[j-1].priority > h.heap[i-1].priority {
			h.swap(j, i)
			i = j
			j = i * 2
		} else {
			break
		}
	}
}

func (h *Heap) init(items []*heapItem) {
	h.heap = items
	n := len(items)
	for i := n / 2; i >= 1; i-- {
		h.down(i, n)
	}
}

func (h *Heap) swap(i, j int) {
	temp := h.heap[i-1]
	h.heap[i-1] = h.heap[j-1]
	h.heap[j-1] = temp
}
