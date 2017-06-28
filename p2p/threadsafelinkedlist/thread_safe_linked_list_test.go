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
	list.Insert(0,"second")
	assert.Equal(t,list.GetCapcity(),int32(2))
	list.Walk()
}

func TestThreadSafeLinkedList_Insert2(t *testing.T) {
	list := NewTSLinkedList("head")
	err := list.Insert(0,"second")
	assert.Nil(t,err)
	err = list.Insert(1,"three")
	assert.Nil(t,err)
	assert.Equal(t,list.GetCapcity(),int32(3))
	list.Walk()
}

func TestThreadSafeLinkedList_Insert3(t *testing.T) {
	list := NewTSLinkedList("head")
	list.Insert(0,"second")
	list.Insert(1,"three")
	list.Insert(1,"append three")
	assert.Equal(t,list.GetCapcity(),int32(4))
	list.Walk()
}

func TestThreadSafeLinkedList_Insert4(t *testing.T) {
	list := NewTSLinkedList("head")
	list.Insert(0,"second")
	list.Insert(1,"three")
	list.Insert(1,"replace three")
	list.Insert(2,"replace four")
	assert.Equal(t,list.GetCapcity(),int32(5))
	list.Walk()
}

func TestThreadSafeLinkedList_Insert5(t *testing.T) {
	list := NewTSLinkedList("head")
	err := list.Insert(1,"second")
	assert.NotNil(t,err)
}

func TestThreadSafeLinkedList_Insert6(t *testing.T) {
	list := NewTSLinkedList("head")
	err := list.Insert(-1,"second")
	assert.NotNil(t,err)
}

func TestThreadSafeLinkedList_Remove(t *testing.T) {
	list := NewTSLinkedList("head")
	list.Insert(0,"second")
	list.Insert(1,"three")
	list.Insert(1,"replace three")
	list.Insert(2,"replace four")
	assert.Equal(t,list.GetCapcity(),int32(5))
	list.Walk()
	v,e := list.Remove(2)
	assert.Nil(t,e)
	assert.Equal(t,"replace three",v.(string))
	list.Walk()
}
func TestThreadSafeLinkedList_Remove2(t *testing.T) {
	list := NewTSLinkedList("head")
	list.Insert(0,"second")
	list.Insert(1,"three")
	list.Insert(1,"replace three")
	list.Insert(2,"replace four")
	assert.Equal(t,list.GetCapcity(),int32(5))
	list.Walk()
	v,e := list.Remove(4)
	assert.Nil(t,e)
	assert.Equal(t,"three",v.(string))
	list.Walk()
}

func TestThreadSafeLinkedList_Remove3(t *testing.T) {
	list := NewTSLinkedList("head")
	list.Insert(0,"second")
	list.Insert(1,"three")
	assert.Equal(t,list.GetCapcity(),int32(3))
	list.Walk()
	_,e := list.Remove(4)
	assert.NotNil(t,e)
}

func TestThreadSafeLinkedList_Remove4(t *testing.T) {
	list := NewTSLinkedList("head")
	list.Insert(0,"second")
	list.Insert(1,"three")
	assert.Equal(t,list.GetCapcity(),int32(3))
	list.Walk()
	_,e := list.Remove(-1)
	assert.NotNil(t,e)
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

func TestThreadSafeLinkedList_Find2(t *testing.T) {
	list := NewTSLinkedList("head")
	list.Insert(0,"second")
	list.Insert(1,"three")
	list.Insert(1,"replace three")
	list.Insert(2,"replace four")
	assert.Equal(t,list.capacity,int32(5))
	list.Walk()
	_,e := list.Find(8)
	assert.NotNil(t,e)
}

func TestThreadSafeLinkedList_Find3(t *testing.T) {
	list := NewTSLinkedList("head")
	list.Insert(0,"second")
	list.Insert(1,"three")
	list.Insert(1,"replace three")
	list.Insert(2,"replace four")
	assert.Equal(t,list.capacity,int32(5))
	list.Walk()
	_,e := list.Find(-1)
	assert.NotNil(t,e)
}

func TestThreadSafeLinkedList_Contains(t *testing.T) {
	list := NewTSLinkedList("head")
	list.Insert(0,"second")
	list.Insert(1,"three")
	list.Insert(1,"replace three")
	list.Insert(2,"replace four")
	assert.Equal(t,list.GetCapcity(),int32(5))
	list.Walk()
	idx,e := list.Contains("three")
	assert.Nil(t,e)
	assert.Equal(t,int32(4),idx)
	list.Walk()
}
func TestThreadSafeLinkedList_Contains2(t *testing.T) {
	list := NewTSLinkedList("head")
	list.Insert(0,"second")
	list.Insert(1,"three")
	list.Insert(1,"replace three")
	list.Insert(2,"replace four")
	assert.Equal(t,list.GetCapcity(),int32(5))
	list.Walk()
	_,e := list.Contains("not exist")
	assert.NotNil(t,e)
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

func TestThreadSafeLinkedList_Iter(t *testing.T) {
	//setup code
	list := NewTSLinkedList("head")
	list.Insert(0,"second")
	list.Insert(1,"three")
	list.Insert(1,"replace three")
	list.Insert(2,"replace four")
	l := list.Iter()
	assert.Equal(t,5,len(l))
}

func TestThreadSafeLinkedList_Iter2(t *testing.T) {
	//setup code
	list := NewTSLinkedList("head")
	list.Insert(0,"second")
	list.Insert(1,"three")
	list.Insert(1,"replace three")
	list.Insert(2,"replace four")
	t.Run("Parallel run 1", func(t *testing.T) {
		t.Parallel()
		l1 := list.Iter()
		assert.Equal(t,5,len(l1))
	})
	t.Run("Parallel run 2", func(t *testing.T) {
		t.Parallel()
		l2 := list.Iter()
		assert.Equal(t,5,len(l2))
	})
	t.Run("Parallel run 3", func(t *testing.T) {
		t.Parallel()
		l3 := list.Iter()
		assert.Equal(t,5,len(l3))
	})
}
func TestThreadSafeLinkedList_IterSlow(t *testing.T) {
	//setup code
	list := NewTSLinkedList("head")
	list.Insert(0,"second")
	list.Insert(1,"three")
	list.Insert(1,"replace three")
	list.Insert(2,"replace four")
	l := list.IterSlow()
	assert.Equal(t,5,len(l))
}


func TestThreadSafeLinkedList_Duplicate(t *testing.T) {
	list := NewTSLinkedList("head")
	list.Insert(0,"second")
	list.Insert(1,"three")
	list.Insert(1,"replace three")
	list.Insert(2,"replace four")
	assert.Equal(t,list.GetCapcity(),int32(5))
	list.Walk()
	newlist,err := list.Duplicate()
	assert.Nil(t,err)
	newlist.Walk()
	_,e := newlist.Contains("replace four")
	assert.Nil(t,e)
	assert.Equal(t,newlist.GetCapcity(),int32(5))
}


func BenchmarkThreadSafeLinkedList_Iter(b *testing.B) {
	//setup code
	list := NewTSLinkedList("head")
	for i:=0;i<100000;i++{
		list.Insert(list.GetCapcity() - 1,"test")
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next(){
			list.Iter()
		}
	})
}

func BenchmarkThreadSafeLinkedList_IterSlow(b *testing.B) {
	//setup code
	list := NewTSLinkedList("head")
	for i:=0;i<100000;i++{
		list.Insert(list.GetCapcity() - 1,"test")
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next(){
			list.IterSlow()
		}
	})
}
