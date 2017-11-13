package threadsafe

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type tesItem struct {
	value    int
	priority int
}

func (i *tesItem) Weight() int {
	return i.priority
}

func (i *tesItem) Value() interface{} {
	return i.value
}

func TestHeap_Walk(t *testing.T) {
	item1 := &tesItem{
		priority: 1,
		value:    1,
	}
	heap := NewHeap(item1)
	h1 := heap.Sort()
	heap.Walk()
	fmt.Println()
	for _, i := range h1 {
		fmt.Println(i)
	}
	fmt.Println("------------")
	heap.Push(3, 3)
	heap.Walk()
	fmt.Println()
	h2 := heap.Sort()
	for _, i := range h2 {
		fmt.Print(i)
	}
	fmt.Println()

	//heap.Push(5,5)
	//assert.Equal(t, 5, heap.Pop().(int))

	fmt.Println("------------")
	heap.Push(4, 4)
	heap.Walk()
	fmt.Println()
	h3 := heap.Sort()
	for _, i := range h3 {
		fmt.Print(i)
	}
	fmt.Println()

	fmt.Println("------------")
	heap.Push(5, 5)
	heap.Walk()
	fmt.Println()
	h4 := heap.Sort()
	for _, i := range h4 {
		fmt.Print(i)
	}
	fmt.Println()

	fmt.Println("------------")
	heap.Remove(4)
	heap.Walk()
	fmt.Println()
	h5 := heap.Sort()
	for _, i := range h5 {
		fmt.Print(i)
	}

	assert.Equal(t, 5, heap.Pop())
}
