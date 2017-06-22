package threadsafelinkedlist

import (
	"sync/atomic"
	"reflect"
	"fmt"
	"github.com/pkg/errors"
	"sync"
)

type ThreadSafeLinkedList struct {
	head *listNode
	tail *listNode
	capacity int32
	rwlock *sync.RWMutex
	// this is this current safe snap shot
	// it should be update when mark is false
	snapshot []interface{}
	// when insert or remove, this should be mark as false
	mark bool
	markLock *sync.RWMutex
}

func NewTSLinkedList(head interface{})*ThreadSafeLinkedList{
	headnode := newListNode(nil,nil,head,int32(0))
	return &ThreadSafeLinkedList{
		head:headnode,
		tail:headnode,
		rwlock:new(sync.RWMutex),
		capacity:int32(1),
		mark:false,
		markLock:new(sync.RWMutex),
	}
}


func (list *ThreadSafeLinkedList)GetCapcity()int32{
	list.rwlock.RLock()
	defer list.rwlock.RUnlock()
	return list.capacity
}


func (list *ThreadSafeLinkedList)Walk(){
	list.rwlock.RLock()
	defer list.rwlock.RUnlock()
	curr := list.head
	for curr != nil{
		fmt.Print(reflect.ValueOf(curr.value),"(",curr.index,")","->")
		curr = curr.next
	}

	fmt.Println()
}

func (list *ThreadSafeLinkedList)IterSlow()[]interface{}{
	// list.mark is false
	list.rwlock.RLock()
	defer list.rwlock.RUnlock()
	curr := list.head
	var l []interface{}
	for curr != nil{
		l = append(l,curr.value)
		curr = curr.next
	}
	return l
}


// Iter get a iterator for this linked list,
// this use a snap_shoot for thread safe,
// improve iter performance
func (list *ThreadSafeLinkedList)Iter() []interface{}{
	list.markLock.RLock()
	if list.mark {
		s := list.snapshot
		list.markLock.RUnlock()
		return s
	}
	list.markLock.RUnlock()
	// list.mark is false
	list.markLock.Lock()
	defer list.markLock.Unlock()
	list.rwlock.RLock()
	defer list.rwlock.RUnlock()
	curr := list.head
	var l []interface{}
	for curr != nil{
		l = append(l,curr.value)
		curr = curr.next
	}
	list.snapshot = l
	list.mark = true
	return list.snapshot
}

//Insert a element after the index, all elements' index after the index will add 1
func (list *ThreadSafeLinkedList)Insert(index int32, item interface{})error{
	list.rwlock.Lock()
	defer list.rwlock.Unlock()
	list.markLock.Lock()
	defer list.markLock.Unlock()
	list.mark = false
	// 前驱，当前
	var curr *listNode
	if index >= list.tail.index{
		newNode := newListNode(list.tail,nil,item,list.tail.index + 1)
		list.tail.next = newNode
		list.tail = newNode
		atomic.AddInt32(&list.capacity,1)
		return nil
	}

	curr = list.head

	//找到需要插入index
	for curr.index < index && curr.next != nil{
		curr = curr.next
	}
	// 0   1   2   3
	// p   c   c.n
	//
	// p   c  (N)  c.n
	if (curr.index == index){
		newNode := newListNode(curr,curr.next,item,index+1)
		curr.next.prev = newNode
		curr.next = newNode
		curr.prev = newNode
		curr = newNode.next
		// 将在该节点之后的节点index全部加1
		for(curr != nil){
			atomic.AddInt32(&curr.index,1)
			curr = curr.next }
		atomic.AddInt32(&list.capacity,1)
		return nil
	}else{
		return errors.New("cannot find the index elements")
	}
}

//Remove the index element and return the removed element
func (list *ThreadSafeLinkedList)Remove(index int32)(interface{},error){
	list.rwlock.Lock()
	defer list.rwlock.Unlock()

	list.markLock.Lock()
	defer list.markLock.Unlock()
	list.mark = false

	if index > list.tail.index {
		return nil,errors.New("the index larger than list size")
	}

	if index == list.tail.index{
		curr := list.tail
		list.tail = curr.prev
		list.tail.next = nil
		return curr.value,nil

	}

	curr := list.head
	prev := list.head
	for curr.index < index && curr.next != nil{
		prev = curr
		curr = curr.next
	}

	if curr.index == index{
		prev.next = curr.next
		curr.next.prev = prev
		value := curr.value
		c := curr.next
		curr = nil
		for (c != nil){
			atomic.AddInt32(&c.index,-1)
			c =c.next
		}
		return value,nil
	}
	return nil, errors.New("cannot find the index elements")
}

//Find the index element and return element
func (list *ThreadSafeLinkedList)Find(index int32)(interface{},error){
	list.rwlock.RLock()
	defer list.rwlock.RUnlock()
	curr := list.head
	if index > list.tail.index{
		return nil,errors.New("the index large than list size")
	}
	for curr.index < index && curr.next != nil{
		curr = curr.next
	}

	if curr.index == index{
		return curr.value,nil
	}
	return nil,errors.New("not found the element")

}

//Contains the element
func (list *ThreadSafeLinkedList)Contains(item interface{})(int32,error){
	list.rwlock.RLock()
	defer list.rwlock.RUnlock()
	curr := list.head
	for curr != nil{
		if reflect.DeepEqual(curr.value,item){
			return curr.index,nil
		}
		curr = curr.next
	}
	return -1,errors.New("Not found")
}