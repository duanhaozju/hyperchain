package info

import "sync"

type Info struct {
	rwmutex *sync.RWMutex
	id int
	isPrimary bool
}

func NewInfo(id int)*Info {
	return &Info{
		rwmutex:new(sync.RWMutex),
		id:id,
		isPrimary:false,
	}
}
//get nodeID
func(i *Info)SetID(id int){
	i.rwmutex.Lock()
	defer i.rwmutex.Unlock()
	i.id = id
}

//get nodeID
func(i *Info)GetID() int{
	i.rwmutex.RLock()
	defer i.rwmutex.RUnlock()
	return i.id
}

//set node primary info
func(i *Info)SetPrimary(flag bool){
	i.rwmutex.Lock()
	defer i.rwmutex.Unlock()
	i.isPrimary = flag
}
// get the node info primary info
func(i *Info)GetPrimary() bool{
	i.rwmutex.RLock()
	defer i.rwmutex.RUnlock()
	return i.isPrimary
}
