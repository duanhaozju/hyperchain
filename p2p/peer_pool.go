package p2p

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"google.golang.org/grpc"
	"hyperchain/admittance"
	"hyperchain/common"
	pb "hyperchain/p2p/message"
	"hyperchain/p2p/persist"
	"hyperchain/p2p/transport"
	"sort"
)

type PeersPool struct {
	peers      map[string]*Peer
	tempPeers  map[string]*Peer
	TM         *transport.TransportManager
	alivePeers int
	localAddr  *pb.PeerAddr
	CM         *admittance.CAManager
	idHashMap  map[int]string
	hashIDMap  map[string]int
	namespace  string
	logger     *logging.Logger
}

// the peers pool instance
//var prPoolIns PeersPool

// NewPeerPool get a new peer pool instance
func NewPeersPool(TM *transport.TransportManager, localAddr *pb.PeerAddr, cm *admittance.CAManager, namespace string) *PeersPool {
	var pPool PeersPool
	pPool.peers = make(map[string]*Peer)
	pPool.localAddr = localAddr
	pPool.tempPeers = make(map[string]*Peer)
	pPool.TM = TM
	pPool.CM = cm
	pPool.alivePeers = 0
	pPool.namespace = namespace
	pPool.logger = common.GetLogger(namespace, "peerspool")
	return &pPool
}

func (pp *PeersPool) setIDHash(id int, hash string) {
	pp.idHashMap[id] = hash
	pp.hashIDMap[hash] = id
}

func (pp *PeersPool) delIDHash(id int) {

}

func (pp *PeersPool) delHashID() {

}

func (pp *PeersPool) getIDByHash(hash string) (int, error) {
	return 0, nil
}

func (pp *PeersPool) getHashByID(id int) (string, error) {
	return "", nil
}

// PutPeer put a peer into the peer pool and get a peer point
func (pp *PeersPool) PutPeer(addr pb.PeerAddr, client *Peer) error {
	//log.Println("Add a peer:",addrString)
	if _, ok := pp.peers[addr.Hash]; ok {
		// the pool already has pp client
		pp.logger.Critical("Node ", addr.ID, ":", addr.IP, addr.Port, "The client already in")
		pp.peers[addr.Hash].Close()
		pp.peers[addr.Hash] = client
		pp.peers[addr.Hash].Alive = true

		persistAddr, err := proto.Marshal(addr.ToPeerAddress())
		if err != nil {
			pp.logger.Errorf("cannot marshal the marshal Addreass for %v", addr)
		}
		persist.PutData(addr.Hash, persistAddr, pp.namespace)
		return errors.New("The client already in, updated the peer info")

	} else {
		pp.logger.Debug("alive peers number is:", pp.alivePeers)
		pp.alivePeers += 1
		pp.peers[addr.Hash] = client
		//persist the peer hash into data base
		persistAddr, err := proto.Marshal(addr.ToPeerAddress())
		if err != nil {
			pp.logger.Errorf("cannot marshal the marshal Addreass for %v", addr)
		}
		persist.PutData(addr.Hash, persistAddr, pp.namespace)
		return nil
	}

}

// PutPeer put a peer into the peer pool and get a peer point
func (pp *PeersPool) PutPeerToTemp(addr pb.PeerAddr, client *Peer) error {
	addrString := addr.Hash
	//log.Println("Add a peer:",addrString)
	if _, ok := pp.tempPeers[addr.Hash]; ok {
		// the pool already has pp client
		pp.logger.Error(addr.IP, addr.Port, "The client already in temp")
		pp.tempPeers[addr.Hash].Close()
		return errors.New("The client already in")

	} else {
		pp.alivePeers += 1
		//pp.tempPeerKeys[addr] = addrString
		//pp.tempPeerAddr[addrString] = addr
		pp.tempPeers[addrString] = client
		return nil
	}

}

// GetPeerByHash
func (pp *PeersPool) GetPeerByHash(hash string) (*Peer, error) {
	if p, ok := pp.peers[hash]; ok {
		return p, nil
	}
	return nil, errors.New("cannot find the peer")
}

// GetPeer get a peer point by the peer address
func (pp *PeersPool) GetPeer(addr pb.PeerAddr) *Peer {
	pp.logger.Critical("inner get peer")
	if client, ok := pp.peers[addr.Hash]; ok {
		return client
	} else {
		return nil
	}
}

// GetAliveNodeNum get all alive node num
func (pp *PeersPool) GetAliveNodeNum() int {
	return pp.alivePeers
}

// GetPeers  get peers from the peer pool
func (pp *PeersPool) GetPeers() []*Peer {
	var clients []*Peer
	if pp.peers != nil {
		for _, cl := range pp.peers {
			clients = append(clients, cl)

		}
	}
	return clients
}

func (pp *PeersPool) GetPeersAddrMap() map[string]pb.PeerAddr {
	var m = make(map[string]pb.PeerAddr)
	for _, cl := range pp.peers {
		m[cl.PeerAddr.Hash] = *cl.PeerAddr
		//log.Critical("取得路由表:", cl)
	}
	return m
}

// GetPeers  get peers from the peer pool
func (pp *PeersPool) GetPeersWithTemp() []*Peer {
	var clients []*Peer
	for _, cl := range pp.peers {
		clients = append(clients, cl)
	}
	for _, tempClient := range pp.tempPeers {
		clients = append(clients, tempClient)
	}
	return clients
}

//将peerspool转换成能够传输的列表
func (pp *PeersPool) ToRoutingTable() pb.Routers {
	peers := pp.GetPeers()
	var routers pb.Routers

	for _, pers := range peers {
		routers.Routers = append(routers.Routers, pers.PeerAddr.ToPeerAddress())
	}
	routers.Routers = append(routers.Routers, pp.localAddr.ToPeerAddress())
	//需要进行排序
	sort.Sort(routers)
	return routers
}

// get routing table without specificToRoutingTableWithout hash
func (pp *PeersPool) ToRoutingTableWithout(hash string) pb.Routers {
	peers := pp.GetPeers()
	var routers pb.Routers

	for _, pers := range peers {
		if pers.PeerAddr.Hash == hash {
			continue
		}
		routers.Routers = append(routers.Routers, pers.PeerAddr.ToPeerAddress())
	}
	//加入自己
	routers.Routers = append(routers.Routers, pp.localAddr.ToPeerAddress())
	//需要进行排序
	sort.Sort(routers)
	for idx, _ := range routers.Routers {
		routers.Routers[idx].ID = int32(idx + 1)
	}
	return routers
}

// merge the route into the temp peer list
func (pp *PeersPool) MergeFromRoutersToTemp(routers pb.Routers, introducer *pb.PeerAddr) {
	payload, err := proto.Marshal(pp.localAddr.ToPeerAddress())
	if err != nil {
		pp.logger.Error("merge from routers error ", err)
		return
	}
	for _, peerAddress := range routers.Routers {
		if peerAddress.Hash == introducer.Hash {
			continue
		}
		peerAddr := pb.RecoverPeerAddr(peerAddress)

		newpeer := NewPeer(peerAddr, pp.localAddr, pp.TM, pp.CM, pp.namespace)
		_, err := newpeer.Connect(payload, pb.Message_ATTEND, false, newpeer.AttendHandler)
		if err != nil {
			pp.logger.Error("merge from routers error ", err)
			continue
		}
		pp.PutPeerToTemp(*newpeer.PeerAddr, newpeer)
	}
}

// Merge the temp peer into peers list
func (pp *PeersPool) MergeTempPeers(peer *Peer) {
	//log.Critical("old节点合并路由表!")
	//使用共识结果进行更新
	//for _, tempPeer := range pp.tempPeers {
	//	if tempPeer.RemoteAddr.Hash == address.Hash {
	if peer == nil {
		pp.logger.Error("the peer to merge is nil")
		return
	}
	pp.peers[peer.PeerAddr.Hash] = peer
	persistAddr, err := proto.Marshal(peer.PeerAddr.ToPeerAddress())
	if err != nil {
		pp.logger.Errorf("cannot marshal the marshal Addreass for %v", peer.PeerAddr)
	}
	persist.PutData(peer.PeerAddr.Hash, persistAddr, pp.namespace)
	delete(pp.tempPeers, peer.PeerAddr.Hash)
	//}
	//}
}

func (pp *PeersPool) MergeTempPeersForNewNode() {
	//使用共识结果进行更新
	for _, tempPeer := range pp.tempPeers {
		pp.peers[tempPeer.PeerAddr.Hash] = tempPeer
		delete(pp.tempPeers, tempPeer.PeerAddr.Hash)

	}
}

//reject the temp peer list
func (pp *PeersPool) RejectTempPeers() {
	for _, tempPeer := range pp.tempPeers {
		delete(pp.tempPeers, tempPeer.PeerAddr.Hash)
		pp.alivePeers -= 1
	}
}

func (pp *PeersPool) DeletePeer(peer *Peer) map[string]pb.PeerAddr {
	pp.alivePeers -= 1
	peer.Close()
	delete(pp.peers, peer.PeerAddr.Hash)
	//go persist.DelData(peer.PeerAddr.Hash)
	perlist := make(map[string]pb.PeerAddr, 1)
	perlist[peer.PeerAddr.Hash] = *peer.PeerAddr
	return perlist
}

func (pp *PeersPool) DeletePeerByHash(hash string) map[string]pb.PeerAddr {
	perlist := make(map[string]pb.PeerAddr, 1)
	for _, peer := range pp.GetPeers() {
		if peer.PeerAddr.Hash == hash {
			pp.alivePeers -= 1
			delete(pp.peers, peer.PeerAddr.Hash)
			//go persist.DelData(peer.PeerAddr.Hash)
			perlist := make(map[string]pb.PeerAddr, 1)
			perlist[peer.PeerAddr.Hash] = *peer.PeerAddr
			break
		}
	}
	return perlist
}
func (pp *PeersPool) SetConnectionByHash(hash string, conn *grpc.ClientConn) error {
	//TODO check error
	if pp.peers == nil {
		pp.peers = make(map[string]*Peer)
	}
	pp.peers[hash].Connection = conn
	return nil
}
func (pp *PeersPool) SetClientByHash(hash string, client pb.ChatClient) error {
	//TODO check error
	if pp.peers == nil {
		pp.peers = make(map[string]*Peer)
	}
	if _, ok := pp.peers[hash]; ok {
		pp.peers[hash].Client = client
		return nil
	} else {
		panic("cannot set")
	}
	//this.peers[hash].Client = client
}

func (pp *PeersPool) Clear() {
	for _, p := range pp.GetPeers() {
		p.Connection.Close()
		p.Close()
	}
	pp.peers = make(map[string]*Peer)
	pp.alivePeers = 0
}