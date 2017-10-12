package p2p

import (
	"encoding/json"
	"fmt"
	"github.com/op/go-logging"
	"github.com/orcaman/concurrent-map"
	"github.com/pkg/errors"
	"hyperchain/common"
	"hyperchain/manager/event"
	"hyperchain/p2p/threadsafe"
	"hyperchain/p2p/utils"
)

var (
	_VP_FLAG  = "VP"
	_NVP_FLAG = "NVP"
)

type PeersPool struct {
	namespace string
	vpPool    *threadsafe.Heap          //todo 当前namespace所有vp节点？
	//nvp hasn't id so use map to storage it
	nvpPool cmap.ConcurrentMap			//todo 当前namespace所有跟node／peer	连的nvp节点？
	//put the exist peers into this exist
	existMap cmap.ConcurrentMap			//todo 当前namespace中所有节点, 包括vp节点和nvp节点？
	// pending map
	pendingMap cmap.ConcurrentMap		//todo 当前namespace中准备连接的节点，可能是vp节点也可能是nvp节点？
	evMux      *event.TypeMux
	//for configuration persist
	pts     *PeerTriples
	peercnf *peerCnf

	logger *logging.Logger
}

// NewPeersPool creates and returns a new PeersPool instance.
func NewPeersPool(namespace string, ev *event.TypeMux, pts *PeerTriples, peercnf *peerCnf) *PeersPool {
	return &PeersPool{
		namespace:  namespace,
		vpPool:     nil,
		nvpPool:    cmap.New(),
		pendingMap: cmap.New(),
		existMap:   cmap.New(),
		evMux:      ev,
		pts:        pts,
		peercnf:    peercnf,
		logger:     common.GetLogger(namespace, "p2p"),
	}
}

func (pool *PeersPool) Ready() bool {
	return pool.vpPool != nil
}

//AddVPPeer add a peer into peers pool instance
func (pool *PeersPool) AddVPPeer(id int, p *Peer) error {
	if pool.vpPool == nil {
		pool.vpPool = threadsafe.NewHeap(p)
		return nil
	}
	pool.vpPool.Push(p, p.Weight())
	hash := utils.GetPeerHash(pool.namespace, id)
	pool.existMap.Set(hash, _VP_FLAG)
	return nil
}

func (pool *PeersPool) PersistList() error {
	pool.peercnf.Lock()
	defer pool.peercnf.Unlock()
	tmppts := NewPeerTriples()
	for _, p := range pool.vpPool.Sort() {
		peer := p.(*Peer)
		pt := NewPeerTriple(peer.info.Namespace, peer.info.Id, peer.info.Hostname)
		tmppts.Push(pt)
	}
	return PersistPeerTriples(pool.peercnf.vip, tmppts)
}

//AddNVPPeer add a peer into peers pool instance
func (pool *PeersPool) AddNVPPeer(hash string, p *Peer) error {
	if tipe, ok := pool.existMap.Get(hash); ok {
		return errors.New(fmt.Sprintf("this peer already in peers pool type: [%s]", tipe.(string)))
	}
	pool.nvpPool.Set(hash, p)
	pool.existMap.Set(hash, _NVP_FLAG)
	return nil
}

//GetPeers get all peer list
func (pool *PeersPool) GetPeers() []*Peer {
	list := make([]*Peer, 0)
	if pool.vpPool == nil {
		return list
	}
	l := pool.vpPool.Sort()
	for _, item := range l {
		list = append(list, item.(*Peer))
	}
	return list
}

func (pool *PeersPool) MaxID() int {
	max := 1
	l := pool.vpPool.Sort()
	for _, item := range l {
		if max < item.(*Peer).info.Id {
			max = item.(*Peer).info.Id
		}
	}
	return max
}

func (pool *PeersPool) GetPeerByHash(hash string) *Peer {
	if pool.vpPool == nil {
		return nil
	}
	l := pool.vpPool.Sort()
	for _, item := range l {
		p := item.(*Peer)
		if p.info.Hash == hash {
			return p
		}
	}
	return nil
}

func (pool *PeersPool) GetPeersByHostname(hostname string) (*Peer, bool) {
	if pool.vpPool == nil {
		return nil, false
	}
	l := pool.vpPool.Sort()
	for _, item := range l {
		p := item.(*Peer)
		if p.info.Hostname == hostname {
			return p, true
		}
	}
	return nil, false
}

func (pool *PeersPool) GetNVPByHostname(hostname string) (*Peer, bool) {
	if pool.nvpPool == nil {
		return nil, false
	}
	l := pool.nvpPool.IterBuffered()
	for item := range l {
		p := item.Val.(*Peer)
		if p.info.Hostname == hostname {
			return p, true
		}
	}
	return nil, false
}

func (pool *PeersPool) GetNVPByHash(hash string) *Peer {
	if pool.nvpPool == nil {
		return nil
	}
	l := pool.nvpPool.IterBuffered()
	for item := range l {
		p := item.Val.(*Peer)
		if p.info.Hash == hash {
			return p
		}
	}
	return nil
}

//TryDelete the specific hash node
func (pool *PeersPool) TryDelete(selfHash, delHash string) (routerhash string, selfnewid uint64, deleteid uint64, err error) {
	pool.logger.Critical("selfhash", selfHash, "delhash", delHash)
	temppool := pool.vpPool.Duplicate()
	var delid int
	for _, item := range temppool.Sort() {
		tempPeer := item.(*Peer)
		if tempPeer.info.Hash == delHash {
			delid = tempPeer.info.Id
		}
	}
	delitem := temppool.Remove(delid)
	if delitem == nil {
		err = errors.New("delete failed, the item not exist.")
	}
	// update all peers id
	for idx, item := range temppool.Sort() {
		tempPeer := item.(*Peer)
		tempPeer.info.SetID(idx + 1)
		if tempPeer.info.Hash == selfHash {
			selfnewid = uint64(tempPeer.info.Id)
		}

	}

	data := make([]string, 0)
	for _, item := range temppool.Sort() {
		peer := item.(*Peer)
		data = append(data, string(peer.Serialize()))
	}

	pools := struct {
		Routers []string `json:"routers"`
	}{}
	b, err := json.Marshal(pools)
	if err != nil {
		return
	}
	routerhash = common.ToHex(utils.Sha3(b))
	deleteid = uint64(delid)
	pool.logger.Criticalf("router hash %s,self id: %d,delete id: %d,e: %v", routerhash, selfnewid, deleteid, err)
	return

}

//DeleteVPPeer delete a peer from peers pool instance
func (pool *PeersPool) DeleteVPPeer(id int) error {
	if pool.vpPool == nil {
		return nil
	}
	v := pool.vpPool.Remove(id)
	if v == nil {
		return errors.New("cannot remove the peer, the peer is not exist.")
	}
	//update all peers id
	for idx, item := range pool.vpPool.Sort() {
		tempPeer := item.(*Peer)
		tempPeer.info.SetID(idx + 1)
	}
	pool.pts.Remove(id)
	return nil
}

func (pool *PeersPool) DeleteVPPeerByHash(hash string) error {
	p := pool.GetPeerByHash(hash)
	if p != nil {
		pool.logger.Critical("delete validate peer", p.info.Id)
	} else {
		pool.logger.Notice("delete validate peer failed.")
		return errors.New(fmt.Sprintf("this validate peer (%s) not exist", hash))
	}
	return pool.DeleteVPPeer(p.info.Id)
}

//DeleteNVPPeer delete the nvp peer
func (pool *PeersPool) DeleteNVPPeer(hash string) error {
	if _, ok := pool.existMap.Get(hash); ok {
		pool.existMap.Remove(hash)
	}
	if _, ok := pool.nvpPool.Get(hash); ok {
		pool.nvpPool.Remove(hash)
	}
	return nil
}

func (pool *PeersPool) GetVPNum() int {
	return len(pool.vpPool.Sort())
}

func (pool *PeersPool) GetNVPNum() int {
	return pool.nvpPool.Count()
}

func (pool *PeersPool) Serlize() ([]byte, error) {
	peers := pool.GetPeers()
	data := make([]string, 0)
	for _, peer := range peers {
		data = append(data, string(peer.Serialize()))
	}

	pools := struct {
		Routers []string `json:"routers"`
	}{}
	b, e := json.Marshal(pools)
	return b, e
}

func PeerPoolUnmarshal(raw []byte) ([]string, error) {
	pools := &struct {
		Routers []string `json:"routers"`
	}{}
	e := json.Unmarshal(raw, pools)
	if e != nil {
		return nil, e
	}
	hostnames := make([]string, 0)
	for _, peers := range pools.Routers {
		h, _, _, e := PeerDeSerialize([]byte(peers))
		if e != nil {
			fmt.Errorf("cannot unmarsal peer,%s", e.Error())
			continue
		}
		hostnames = append(hostnames, h)
	}
	return hostnames, e
}
