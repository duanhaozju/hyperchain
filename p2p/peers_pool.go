package p2p

import (
	"github.com/pkg/errors"
	"github.com/orcaman/concurrent-map"
	"hyperchain/p2p/utils"
	"fmt"
	"hyperchain/p2p/threadsafelinkedlist"
	"encoding/json"
	"github.com/op/go-logging"
	"hyperchain/common"
)

var (	_VP_FLAG = "VP"
	_NVP_FLAG= "NVP")

type PeersPool struct {
	namespace string
	vpPool *threadsafelinkedlist.ThreadSafeLinkedList
	//nvp hasn't id so use map to storage it
	nvpPool cmap.ConcurrentMap
	//put the exist peers into this exist
	existMap cmap.ConcurrentMap
	logger *logging.Logger
}

//NewPeersPool new a peers pool
func NewPeersPool(namespace string)*PeersPool {
	return &PeersPool{
		namespace:namespace,
		vpPool:nil,
		nvpPool:cmap.New(),
		existMap:cmap.New(),
		logger:common.GetLogger(namespace,"p2p"),
	}
}

//AddVPPeer add a peer into peers pool instance
func (pool *PeersPool)AddVPPeer(id int,p *Peer)error{
	_id := id - 1
	if pool.vpPool == nil{
		if _id != 0{
			return errors.New(fmt.Sprintf("the vp peers pool is empty, could not add index: %d peer",id))
		}
		pool.vpPool = threadsafelinkedlist.NewTSLinkedList(p)
		return nil
	}
	err := pool.vpPool.Insert(int32(_id),p)
	if err != nil{
		return err
	}
	hash := utils.GetPeerHash(pool.namespace,id)
	pool.existMap.Set(hash,_VP_FLAG)
	return nil
}

//AddNVPPeer add a peer into peers pool instance
func (pool *PeersPool)AddNVPPeer(id int,p *Peer)error{
	hash := utils.GetPeerHash(pool.namespace,id)
	if tipe,ok := pool.existMap.Get(hash);ok{
		return errors.New(fmt.Sprintf("this peer already in peers pool type: [%s]",tipe.(string)))
	}
	pool.nvpPool.Set(hash,p)
	pool.existMap.Set(hash,_NVP_FLAG)
	return nil
}

//GetPeers get all peer list
func (pool *PeersPool)GetPeers()[]*Peer {
	list := make([]*Peer,0)
	if pool.vpPool == nil{
		return list
	}
	l := pool.vpPool.Iter()
	for _,item := range l{
		list = append(list,item.(*Peer))
	}
	return list
}

func (pool *PeersPool)GetPeersByHash(hash string)*Peer{
	l := pool.vpPool.Iter()
	for _,item := range l{
		p := item.(*Peer)
		if p.info.Hash == hash{
			return p
		}
	}
	return nil
}

func(pool *PeersPool)TryDelete(selfHash,delHash string)(routerhash string, selfnewid uint64,deleteid uint64,err error){
	pool.logger.Critical("selfhash",selfHash,"delhash",delHash)
	templist,err := pool.vpPool.Duplicate()
	if err !=nil{
		return
	}
	var delid int
	for _,item := range templist.Iter(){
		tempPeer := item.(*Peer)
		if tempPeer.info.Hash == delHash{
			pool.logger.Critical("<===> %+v",tempPeer.info)
			delid = tempPeer.info.Id
		}
	}
	_,err = templist.Remove(int32(delid-1))
	if err != nil{
		return
	}

	// update all peers id
	for idx,item := range templist.Iter(){
		tempPeer := item.(*Peer)
		tempPeer.info.SetID(idx + 1)
		if tempPeer.info.Hash == selfHash{
			selfnewid = uint64(tempPeer.info.Id)
		}

	}

	data := make([]string,0)
	for _,item := range templist.Iter(){
		peer := item.(*Peer)
		data = append(data,string(peer.Serialize()))
	}

	pools := struct {
		Routers []string `json:"routers"`
	}{}
	b,err := json.Marshal(pools)
	if err != nil{
		return
	}
	routerhash = common.ToHex(utils.Sha3(b))
	deleteid = uint64(delid)
	pool.logger.Criticalf("r %s,selfid: %d,deleteid: %d,e: %s",routerhash, selfnewid ,deleteid ,err)
	return

}


//DeleteVPPeer delete a peer from peers pool instance
func(pool *PeersPool)DeleteVPPeer(id int)error{
	_,err := pool.vpPool.Remove(int32(id-1))
	//update all peers id
	for idx,item := range pool.vpPool.Iter(){
		tempPeer := item.(*Peer)
		tempPeer.info.SetID(idx + 1)
	}
	return err
}

func(pool *PeersPool)DeleteVPPeerByHash(hash string)error{
	p := pool.GetPeersByHash(hash)
	pool.logger.Critical("delete node",p.info.Id)
	return pool.DeleteVPPeer(p.info.Id)
}


//DeleteNVPPeer delete the nvp peer
func(pool *PeersPool)DeleteNVPPeer(hash string){
	if _,ok := pool.nvpPool.Get(hash);ok{
		pool.nvpPool.Remove(hash)
	}
}

func (pool *PeersPool)Serlize()([]byte,error){
	peers := pool.GetPeers()
	data := make([]string,0)
	for _,peer := range peers{
		data = append(data,string(peer.Serialize()))
	}

	pools := struct {
		Routers []string `json:"routers"`
	}{}
	b,e := json.Marshal(pools)
	return b,e
}



func PeerPoolUnmarshal(raw []byte)([]string,error){
	pools := &struct {
		Routers []string `json:"routers"`
	}{}
	e := json.Unmarshal(raw,pools)
	if e!=nil{
		return nil,e
	}
	hostnames := make([]string,0)
	for _,peers :=range pools.Routers{
		h,_,_,e := PeerUnSerialize([]byte(peers))
		if e !=nil{
			fmt.Errorf("cannot unmarsal peer,%s",e.Error())
			continue
		}
		hostnames = append(hostnames,h)
	}
	return hostnames,e
}
