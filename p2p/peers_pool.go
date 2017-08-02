package p2p

import (
	"github.com/pkg/errors"
	"github.com/orcaman/concurrent-map"
	"hyperchain/p2p/utils"
	"fmt"
	"hyperchain/p2p/threadsafe"
	"encoding/json"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/manager/event"
)

var (	_VP_FLAG = "VP"
	_NVP_FLAG= "NVP")

type PeersPool struct {
	namespace string
	vpPool *threadsafe.Heap
	//nvp hasn't id so use map to storage it
	nvpPool cmap.ConcurrentMap
	//put the exist peers into this exist
	existMap cmap.ConcurrentMap
	evMux *event.TypeMux
	//for configuration persist
	pts *PeerTriples
	peercnf *peerCnf

	logger *logging.Logger
}

//NewPeersPool new a peers pool
func NewPeersPool(namespace string,ev *event.TypeMux,pts *PeerTriples,peercnf *peerCnf)*PeersPool {
	return &PeersPool{
		namespace:namespace,
		vpPool:nil,
		nvpPool:cmap.New(),
		existMap:cmap.New(),
		evMux:ev,
		pts: pts,
		peercnf:peercnf,
		logger:common.GetLogger(namespace,"p2p"),
	}
}

//AddVPPeer add a peer into peers pool instance
func (pool *PeersPool)AddVPPeer(id int,p *Peer)error{
	if pool.vpPool == nil{
		pool.vpPool = threadsafe.NewHeap(p)
		return nil
	}
	pool.vpPool.Push(p,p.Weight())
	hash := utils.GetPeerHash(pool.namespace,id)
	pool.existMap.Set(hash,_VP_FLAG)
	return nil
}

func(pool *PeersPool)PersistList() error {
	pool.peercnf.Lock()
	defer pool.peercnf.Unlock()
	tmppts := NewPeerTriples()
	for _,p := range pool.vpPool.Sort(){
		peer :=  p.(*Peer)
		pt := NewPeerTriple(peer.info.Namespace,peer.info.Id,peer.info.Hostname)
		tmppts.Push(pt)
	}
	return PersistPeerTriples(pool.peercnf.vip,tmppts)
}

//AddNVPPeer add a peer into peers pool instance
func (pool *PeersPool)AddNVPPeer(hash string,p *Peer)error{
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
	l := pool.vpPool.Sort()
	for _,item := range l{
		list = append(list,item.(*Peer))
	}
	return list
}

func (pool *PeersPool)GetPeerByHash(hash string)*Peer{
	l := pool.vpPool.Sort()
	for _,item := range l{
		p := item.(*Peer)
		if p.info.Hash == hash{
			return p
		}
	}
	return nil
}

func (pool *PeersPool)GetPeersByHostname(hostname string)*Peer{
	if pool.vpPool == nil{
		return nil
	}
	l := pool.vpPool.Sort()
	for _,item := range l{
		p := item.(*Peer)
		if p.info.Hostname == hostname{
			return p
		}
	}
	return nil
}

func (pool *PeersPool)GetNVPByHostname(hostname string)*Peer{
	if pool.nvpPool == nil{
		return nil
	}
	l := pool.nvpPool.IterBuffered()
	for item := range l{
		p := item.Val.(*Peer)
		if p.info.Hostname == hostname{
			return p
		}
	}
	return nil
}

func (pool *PeersPool)GetNVPByHash(hash string)*Peer{
	if pool.nvpPool == nil{
		return nil
	}
	l := pool.nvpPool.IterBuffered()
	for item := range l{
		p := item.Val.(*Peer)
		if p.info.Hash == hash{
			return p
		}
	}
	return nil
}
//TryDelete the specific hash node
func(pool *PeersPool)TryDelete(selfHash,delHash string)(routerhash string, selfnewid uint64,deleteid uint64,err error){
	pool.logger.Critical("selfhash",selfHash,"delhash",delHash)
	temppool := pool.vpPool.Duplicate()
	var delid int
	for _,item := range temppool.Sort(){
		tempPeer := item.(*Peer)
		if tempPeer.info.Hash == delHash{
			pool.logger.Critical("delete peer => %+v",tempPeer.info)
			delid = tempPeer.info.Id
		}
	}
	delitem := temppool.Remove(delid-1)
	if delitem == nil{
		err = errors.New("delete failed, the item not exist.")
	}
	// update all peers id
	for idx,item := range temppool.Sort(){
		tempPeer := item.(*Peer)
		tempPeer.info.SetID(idx + 1)
		if tempPeer.info.Hash == selfHash{
			selfnewid = uint64(tempPeer.info.Id)
		}

	}

	data := make([]string,0)
	for _,item := range temppool.Sort(){
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
	if pool.vpPool == nil{
		return nil
	}
	v := pool.vpPool.Remove(id-1)
	if v == nil{
		return errors.New("cannot remove the peer, the peer is not exist.")
	}
	//update all peers id
	for idx,item := range pool.vpPool.Sort(){
		tempPeer := item.(*Peer)
		tempPeer.info.SetID(idx + 1)
	}
	pool.pts.Remove(id)
	return nil
}

func(pool *PeersPool)DeleteVPPeerByHash(hash string)error{
	p := pool.GetPeerByHash(hash)
	if p != nil{
		pool.logger.Critical("delete validate peer",p.info.Id)
	}else{
		pool.logger.Notice("delete validate peer failed.")
		return errors.New(fmt.Sprintf("this validate peer (%s) not exist",hash))
	}
	return pool.DeleteVPPeer(p.info.Id)
}


//DeleteNVPPeer delete the nvp peer
func(pool *PeersPool)DeleteNVPPeer(hash string)error{
	if _,ok := pool.nvpPool.Get(hash);ok{
		pool.nvpPool.Remove(hash)
	}
	return nil
}

func(pool *PeersPool)GetVPNum() int{
	return len(pool.vpPool.Sort())
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
