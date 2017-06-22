package info

import (
	"sync"
	"hyperchain/crypto/sha3"
	"hyperchain/common"
)

type Info struct {
	rwmutex *sync.RWMutex
	id int
	isPrimary bool
	hostname string
	namespace string
	hash string
}

func NewInfo(id int,hostname string,namespcace string)*Info {
	h := sha3.NewKeccak256()
	h.Write([]byte(namespcace))
	hash := h.Sum([]byte(hostname))
	return &Info{
		rwmutex:new(sync.RWMutex),
		id:id,
		isPrimary:false,
		hostname:hostname,
		hash:common.Bytes2Hex(hash),
	}
}

func (i *Info)GetHash()string{
	i.rwmutex.RLock()
	defer i.rwmutex.RUnlock()
	return i.hash
}


func(i *Info)GetHostName(hostname string){
	i.rwmutex.RLock()
	defer i.rwmutex.RUnlock()
	i.hostname = hostname
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
