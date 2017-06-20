package threadsafelinkedlist

import (
	"testing"
	"github.com/stretchr/testify/assert"
)


func TestNewTSLinkedList(t *testing.T) {
	list := NewTSLinkedList("head")
	t.Log(list.capacity)
}

func TestThreadSafeLinkedList_Insert(t *testing.T) {
	list := NewTSLinkedList("head")
	list.Insert(1,"second")
	assert.Equal(t,list.capacity,int32(2))
	list.Walk()
}

func TestThreadSafeLinkedList_Insert2(t *testing.T) {
	list := NewTSLinkedList("head")
	list.Insert(1,"second")
	list.Insert(2,"three")
	assert.Equal(t,list.capacity,int32(3))
	list.Walk()
}

func TestThreadSafeLinkedList_Insert3(t *testing.T) {
	list := NewTSLinkedList("head")
	list.Insert(0,"second")
	list.Insert(1,"three")
	list.Insert(1,"append three")
	assert.Equal(t,list.capacity,int32(4))
	list.Walk()
}

func TestThreadSafeLinkedList_Insert4(t *testing.T) {
	list := NewTSLinkedList("head")
	list.Insert(0,"second")
	list.Insert(1,"three")
	list.Insert(1,"replace three")
	list.Insert(2,"replace four")
	assert.Equal(t,list.capacity,int32(5))
	list.Walk()
}

func TestThreadSafeLinkedList_Remove(t *testing.T) {
	list := NewTSLinkedList("head")
	list.Insert(0,"second")
	list.Insert(1,"three")
	list.Insert(1,"replace three")
	list.Insert(2,"replace four")
	assert.Equal(t,list.capacity,int32(5))
	list.Walk()
	v,e := list.Remove(2)
	assert.Nil(t,e)
	assert.Equal(t,"replace three",v.(string))
	list.Walk()
}


func BenchmarkThreadSafeLinkedList_Insert(b *testing.B) {
	list := NewTSLinkedList("head")
	for i:=0; i<b.N; i++{
		idx := list.capacity
		list.Insert(idx,"second")
	}
}

func TestThreadSafeLinkedList_Find(t *testing.T) {
	list := NewTSLinkedList("head")
	list.Insert(0,"second")
	list.Insert(1,"three")
	list.Insert(1,"replace three")
	list.Insert(2,"replace four")
	assert.Equal(t,list.capacity,int32(5))
	list.Walk()
	v,e := list.Find(2)
	assert.Nil(t,e)
	assert.Equal(t,"replace three",v.(string))
	list.Walk()
}

func TestThreadSafeLinkedList_Contains(t *testing.T) {
	list := NewTSLinkedList("head")
	list.Insert(0,"second")
	list.Insert(1,"three")
	list.Insert(1,"replace three")
	list.Insert(2,"replace four")
	assert.Equal(t,list.capacity,int32(5))
	list.Walk()
	idx,e := list.Contains("three")
	assert.Nil(t,e)
	assert.Equal(t,int32(4),idx)
	list.Walk()
}

func BenchmarkThreadSafeLinkedList_Find(b *testing.B) {
	list := NewTSLinkedList("head")
	for i:=0;i<100000;i++{
		list.Insert(list.GetCapcity() - 1,"test")
	}
	for i:=0;i<b.N;i++{
		list.Find(99999)
	}
}

func BenchmarkThreadSafeLinkedList_Contains(b *testing.B) {
	list := NewTSLinkedList("head")
	list.Insert(0,"second")
	list.Insert(1,"three")
	list.Insert(1,"replace three")
	list.Insert(2,"replace four")
	for i:= 0;i<b.N;i++{
		list.Contains("three")
	}
}

