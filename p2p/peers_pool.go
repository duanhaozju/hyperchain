package p2p

import (
	"sync"
	"github.com/pkg/errors"
)

type PeersPool struct {
	rwMutex *sync.RWMutex
	peers map[int]*peer
}

func NewPeersPool()*PeersPool {
	return &PeersPool{
		rwMutex:new(sync.RWMutex),
		peers:make(map[int]*peer),
	}
}

func (pool *PeersPool)GetIterator()[]*peer{
	pool.rwMutex.RLock()
	defer pool.rwMutex.RUnlock()
	peerList := make([]*peer,0)
	for _,value := range pool.peers{
		peerList = append(peerList,value)
	}
	return peerList
}

//add a peer into peers pool instance
func (pool *PeersPool)AddPeer(id int,p *peer)error{
	pool.rwMutex.Lock()
	defer pool.rwMutex.Unlock()
	if _,ok := pool.peers[id];ok {
		return errors.New("this peer already in peers pool")
	}
	pool.peers[id] = p
	return nil
}

//delete a peer from peers pool instance
func(pool *PeersPool)DeletePeer(id int)error{
	pool.rwMutex.Lock()
	defer pool.rwMutex.Unlock()
	if _,ok := pool.peers[id];!ok {
		return  nil
	}
	delete(pool.peers,id)
	return nil
}
