package p2p

import (
	"sync"
	"github.com/pkg/errors"
	"github.com/orcaman/concurrent-map"
	"hyperchain/p2p/utils"
	"fmt"
)

type PeersPool struct {
	namespace string
	vpPool cmap.ConcurrentMap
	nvpPool cmap.ConcurrentMap
	//put the exist peers into this exist
	existMap cmap.ConcurrentMap
}

func NewPeersPool(namespace string)*PeersPool {
	return &PeersPool{
		namespace:namespace,
		vpPool:cmap.New(),
		nvpPool:cmap.New(),
		existMap:cmap.New(),
	}
}

func (pool *PeersPool)GetIterator()[]*peer{
	peerList := make([]*peer,0)

	for _,value := range pool.peers{
		peerList = append(peerList,value)
	}
	return peerList
}

//add a peer into peers pool instance
func (pool *PeersPool)AddVPPeer(id int,p *peer)error{
	hash := utils.GetPeerHash(pool.namespace,id)
	if tipe,ok := pool.existMap.Get(hash);ok{
		return errors.New(fmt.Sprintf("this peer already in peers pool type: [%s]",tipe.(string)))
	}
	pool.vpPool.Set(hash,p)
	pool.existMap.Set(hash,"VP")
	return nil
}

//add a peer into peers pool instance
func (pool *PeersPool)AddNVPPeer(id int,p *peer)error{
	hash := utils.GetPeerHash(pool.namespace,id)
	if tipe,ok := pool.existMap.Get(hash);ok{
		return errors.New(fmt.Sprintf("this peer already in peers pool type: [%s]",tipe.(string)))
	}
	pool.nvpPool.Set(hash,p)
	pool.existMap.Set(hash,"NVP")
	return nil
}

//delete a peer from peers pool instance
func(pool *PeersPool)DeletePeer(id int)error{
	if _,ok := pool.peers[id];!ok {
		return  nil
	}
	delete(pool.peers,id)
	return nil
}
