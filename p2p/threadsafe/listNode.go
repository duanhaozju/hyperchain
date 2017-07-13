// Author: chenquan
// Date: 2017-06-20

package threadsafe

import "sync"

type listNode struct {
	prev *listNode
	next *listNode
	value interface{}
	index int32
	//mark this node is logic deleted
	marked bool
	rwLock sync.Mutex
}

func newListNode(prev,next *listNode,value interface{},index int32)*listNode{
	return &listNode{
		prev:prev,
		next:next,
		value:value,
		index:index,
	}
}

type listElement interface {
	Weight() int
}

func (node *listNode)setPrev(p *listNode){
}


