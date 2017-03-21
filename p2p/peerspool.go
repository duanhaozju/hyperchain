package p2p

import (
	"errors"
	"google.golang.org/grpc"
	"hyperchain/admittance"
	pb "hyperchain/p2p/peermessage"
	"hyperchain/p2p/transport"
	"sort"
	"hyperchain/p2p/persist"
	"github.com/golang/protobuf/proto"
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
}

// the peers pool instance
//var prPoolIns PeersPool

// NewPeerPool get a new peer pool instance
func NewPeersPool(TM *transport.TransportManager, localAddr *pb.PeerAddr, cm *admittance.CAManager) *PeersPool {
	var pPool PeersPool
	pPool.peers = make(map[string]*Peer)
	//newPrPoolIns.peerAddr = make(map[string]pb.PeerAddr)
	//newPrPoolIns.peerKeys = make(map[pb.PeerAddr]string)
	pPool.localAddr = localAddr
	pPool.tempPeers = make(map[string]*Peer)
	//newPrPoolIns.tempPeerAddr = make(map[string]pb.PeerAddr)
	//newPrPoolIns.tempPeerKeys = make(map[pb.PeerAddr]string)
	pPool.TM = TM
	pPool.CM = cm
	pPool.alivePeers = 0
	return &pPool
}

func (this *PeersPool)setIDHash(id int, hash string) {
	this.idHashMap[id] = hash
	this.hashIDMap[hash] = id
}

func (this *PeersPool)delIDHash(id int) {

}

func (this *PeersPool)delHashID() {

}

func (this *PeersPool)getIDByHash(hash string) (int, error) {
	return 0, nil
}

func (this *PeersPool)getHashByID(id int) (string, error) {
	return "", nil
}

// PutPeer put a peer into the peer pool and get a peer point
func (this *PeersPool) PutPeer(addr pb.PeerAddr, client *Peer) error {
	//log.Println("Add a peer:",addrString)
	if _, ok := this.peers[addr.Hash]; ok {
		// the pool already has this client
		log.Critical("Node ", addr.ID, ":", addr.IP, addr.Port, "The client already in")
		this.peers[addr.Hash].Close()
		this.peers[addr.Hash] = client
		this.peers[addr.Hash].Alive = true

		persistAddr, err := proto.Marshal(addr.ToPeerAddress())
		if err != nil {
			log.Errorf("cannot marshal the marshal Addreass for %v", addr)
		}
		persist.PutData(addr.Hash, persistAddr)
		return errors.New("The client already in, updated the peer info")

	} else {
		log.Debug("alive peers number is:", this.alivePeers)
		this.alivePeers += 1
		this.peers[addr.Hash] = client
		//persist the peer hash into data base
		persistAddr, err := proto.Marshal(addr.ToPeerAddress())
		if err != nil {
			log.Errorf("cannot marshal the marshal Addreass for %v", addr)
		}
		persist.PutData(addr.Hash, persistAddr)
		return nil
	}

}

// PutPeer put a peer into the peer pool and get a peer point
func (this *PeersPool) PutPeerToTemp(addr pb.PeerAddr, client *Peer) error {
	addrString := addr.Hash
	//log.Println("Add a peer:",addrString)
	if _, ok := this.tempPeers[addr.Hash]; ok {
		// the pool already has this client
		log.Error(addr.IP, addr.Port, "The client already in temp")
		this.tempPeers[addr.Hash].Close()
		return errors.New("The client already in")

	} else {
		this.alivePeers += 1
		//this.tempPeerKeys[addr] = addrString
		//this.tempPeerAddr[addrString] = addr
		this.tempPeers[addrString] = client
		return nil
	}

}

// GetPeerByHash
func (this *PeersPool) GetPeerByHash(hash string) (*Peer, error) {
	if p, ok := this.peers[hash]; ok {
		return p, nil
	}
	return nil, errors.New("cannot find the peer")
}

// GetPeer get a peer point by the peer address
func (this *PeersPool) GetPeer(addr pb.PeerAddr) *Peer {
	log.Critical("inner get peer")
	if client, ok := this.peers[addr.Hash]; ok {
		return client
	} else {
		return nil
	}
}

// GetAliveNodeNum get all alive node num
func (this *PeersPool) GetAliveNodeNum() int {
	return this.alivePeers
}

// GetPeers  get peers from the peer pool
func (this *PeersPool) GetPeers() []*Peer {
	var clients []*Peer
	for _, cl := range this.peers {
		clients = append(clients, cl)
		//log.Critical("取得路由表:", cl)
	}

	return clients
}

func (this *PeersPool) GetPeersAddrMap() map[string]pb.PeerAddr {
	var m = make(map[string]pb.PeerAddr)
	for _, cl := range this.peers {
		m[cl.PeerAddr.Hash] = *cl.PeerAddr
		//log.Critical("取得路由表:", cl)
	}
	return m
}

// GetPeers  get peers from the peer pool
func (this *PeersPool) GetPeersWithTemp() []*Peer {
	var clients []*Peer
	for _, cl := range this.peers {
		clients = append(clients, cl)
	}
	for _, tempClient := range this.tempPeers {
		clients = append(clients, tempClient)
	}
	return clients
}

//将peerspool转换成能够传输的列表
func (this *PeersPool) ToRoutingTable() pb.Routers {
	peers := this.GetPeers()
	var routers pb.Routers

	for _, pers := range peers {
		routers.Routers = append(routers.Routers, pers.PeerAddr.ToPeerAddress())
	}
	routers.Routers = append(routers.Routers, this.localAddr.ToPeerAddress())
	//需要进行排序
	sort.Sort(routers)
	return routers
}

// get routing table without specificToRoutingTableWithout hash
func (this *PeersPool) ToRoutingTableWithout(hash string) pb.Routers {
	peers := this.GetPeers()
	var routers pb.Routers

	for _, pers := range peers {
		if pers.PeerAddr.Hash == hash {
			continue
		}
		routers.Routers = append(routers.Routers, pers.PeerAddr.ToPeerAddress())
	}
	//加入自己
	routers.Routers = append(routers.Routers, this.localAddr.ToPeerAddress())
	//需要进行排序
	sort.Sort(routers)
	for idx, _ := range routers.Routers {
		routers.Routers[idx].ID = int32(idx + 1)
	}
	return routers
}

// merge the route into the temp peer list
func (this *PeersPool) MergeFromRoutersToTemp(routers pb.Routers, introducer *pb.PeerAddr) {
	payload, err := proto.Marshal(this.localAddr.ToPeerAddress())
	if err != nil {
		log.Error("merge from routers error ", err)
		return
	}
	for _, peerAddress := range routers.Routers {
		if peerAddress.Hash == introducer.Hash {
			continue
		}
		peerAddr := pb.RecoverPeerAddr(peerAddress)

		newpeer := NewPeer(peerAddr, this.localAddr, this.TM, this.CM)
		_, err := newpeer.Connect(payload, pb.Message_ATTEND, false, newpeer.AttendHandler)
		if err != nil {
			log.Error("merge from routers error ", err)
			continue
		}
		this.PutPeerToTemp(*newpeer.PeerAddr, newpeer)
	}
}

// Merge the temp peer into peers list
func (this *PeersPool) MergeTempPeers(peer *Peer) {
	//log.Critical("old节点合并路由表!")
	//使用共识结果进行更新
	//for _, tempPeer := range this.tempPeers {
	//	if tempPeer.RemoteAddr.Hash == address.Hash {
	if peer == nil {
		log.Error("the peer to merge is nil");
		return
	}
	this.peers[peer.PeerAddr.Hash] = peer
	persistAddr, err := proto.Marshal(peer.PeerAddr.ToPeerAddress())
	if err != nil {
		log.Errorf("cannot marshal the marshal Addreass for %v", peer.PeerAddr)
	}
	persist.PutData(peer.PeerAddr.Hash, persistAddr)
	delete(this.tempPeers, peer.PeerAddr.Hash)
	//}
	//}
}

func (this *PeersPool) MergeTempPeersForNewNode() {
	//使用共识结果进行更新
	for _, tempPeer := range this.tempPeers {
		this.peers[tempPeer.PeerAddr.Hash] = tempPeer
		delete(this.tempPeers, tempPeer.PeerAddr.Hash)

	}
}

//reject the temp peer list
func (this *PeersPool) RejectTempPeers() {
	for _, tempPeer := range this.tempPeers {
		delete(this.tempPeers, tempPeer.PeerAddr.Hash)
		this.alivePeers -= 1
	}
}

func (this *PeersPool) DeletePeer(peer *Peer) map[string]pb.PeerAddr {
	this.alivePeers -= 1
	peer.Close()
	delete(this.peers, peer.PeerAddr.Hash)
	//go persist.DelData(peer.PeerAddr.Hash)
	perlist := make(map[string]pb.PeerAddr, 1)
	perlist[peer.PeerAddr.Hash] = *peer.PeerAddr
	return perlist
}

func (this *PeersPool) DeletePeerByHash(hash string) map[string]pb.PeerAddr {
	perlist := make(map[string]pb.PeerAddr, 1)
	for _, peer := range this.GetPeers() {
		if peer.PeerAddr.Hash == hash {
			this.alivePeers -= 1
			delete(this.peers, peer.PeerAddr.Hash)
			//go persist.DelData(peer.PeerAddr.Hash)
			perlist := make(map[string]pb.PeerAddr, 1)
			perlist[peer.PeerAddr.Hash] = *peer.PeerAddr
			break;
		}
	}
	return perlist
}
func (this *PeersPool) SetConnectionByHash(hash string, conn *grpc.ClientConn) error {
	//TODO check error
	if this.peers == nil {
		this.peers = make(map[string]*Peer)
	}
	this.peers[hash].Connection = conn
	return nil
}
func (this *PeersPool) SetClientByHash(hash string, client pb.ChatClient) error {
	//TODO check error
	if this.peers == nil {
		this.peers = make(map[string]*Peer)
	}
	if _, ok := this.peers[hash]; ok {
		this.peers[hash].Client = client
		return nil
	} else {
		panic("cannot set")
	}
	//this.peers[hash].Client = client
}

func (this *PeersPool) Clear() {
	for _, p := range this.GetPeers() {
		p.Connection.Close()
	}
	this.peers = make(map[string]*Peer)
	this.alivePeers = 0
}
