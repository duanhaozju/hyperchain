package random_stack

import (
	"testing"
	"fmt"
)

func TestStack_Push(t *testing.T) {
	stack := NewStack()
	stack.Push(1)
	stack.Walk()
}

func TestStack_Pop(t *testing.T) {
	stack := NewStack()
	stack.Push(1)
	stack.Push(2)
	stack.Walk()
	fmt.Println("-------------")
	stack.Pop()
	stack.Walk()
}

func TestStack_RandomPop(t *testing.T) {
	stack := NewStack()
	stack.Push(1)
	stack.Push(2)
	stack.Push(3)
	stack.Push(4)
	stack.Push(5)
	stack.Push(6)
	stack.Push(7)
	stack.Push(8)
	stack.Walk()
	fmt.Println("-------------")
	stack.RandomPop()
	stack.Walk()
	fmt.Println("-------------")
	stack.RandomPop()
	stack.Walk()
	fmt.Println("-------------")
	stack.RandomPop()
	stack.Walk()
}

func BenchmarkStack_RandomPop(b *testing.B) {
	stack := NewStack()
	for i:=0;i<b.N;i++{
		stack.Push(1)
	}

	for i:=0;i<b.N;i++{
		stack.RandomPop()
	}
}

