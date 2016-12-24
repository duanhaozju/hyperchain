package p2p

import (
	pb "hyperchain/p2p/peermessage"
	"hyperchain/p2p/transport"
	"errors"
	"sort"
)

type PeersPoolIml struct {
	peers        map[string]*Peer
	peerAddr     map[string]pb.PeerAddr
	peerKeys     map[pb.PeerAddr]string
	tempPeers    map[string]*Peer
	tempPeerAddr map[string]pb.PeerAddr
	tempPeerKeys map[pb.PeerAddr]string
	TEM          transport.TransportEncryptManager
	alivePeers   int
	localAddr    *pb.PeerAddr
}

// the peers pool instance
//var prPoolIns PeersPool

// NewPeerPool get a new peer pool instance
func NewPeerPoolIml(TEM transport.TransportEncryptManager,localAddr *pb.PeerAddr) (newPrPoolIns *PeersPoolIml) {
	newPrPoolIns.peers = make(map[string]*Peer)
	newPrPoolIns.peerAddr = make(map[string]pb.PeerAddr)
	newPrPoolIns.peerKeys = make(map[pb.PeerAddr]string)
	newPrPoolIns.localAddr = localAddr
	newPrPoolIns.tempPeers = make(map[string]*Peer)
	newPrPoolIns.tempPeerAddr = make(map[string]pb.PeerAddr)
	newPrPoolIns.tempPeerKeys = make(map[pb.PeerAddr]string)
	newPrPoolIns.TEM = TEM
	newPrPoolIns.alivePeers = 0
	return
}

// PutPeer put a peer into the peer pool and get a peer point
func (this *PeersPoolIml) PutPeer(addr pb.PeerAddr, client *Peer) error {
	addrString := addr.Hash
	//log.Println("Add a peer:",addrString)
	if _, ok := this.peerKeys[addr]; ok {
		// the pool already has this client
		log.Error(addr.IP, addr.Port, "The client already in")
		return this.peers[addrString], errors.New("The client already in")

	} else {
		this.alivePeers += 1
		this.peerKeys[addr] = addrString
		this.peerAddr[addrString] = addr
		this.peers[addrString] = client
		return client, nil
	}

}

// PutPeer put a peer into the peer pool and get a peer point
func (this *PeersPoolIml) PutPeerToTemp(addr pb.PeerAddr, client *Peer) (*Peer, error) {
	addrString := addr.Hash
	//log.Println("Add a peer:",addrString)
	if _, ok := this.tempPeerKeys[addr]; ok {
		// the pool already has this client
		log.Error(addr.IP, addr.Port, "The client already in temp")
		return this.tempPeers[addrString], errors.New("The client already in")

	} else {
		this.alivePeers += 1
		this.tempPeerKeys[addr] = addrString
		this.tempPeerAddr[addrString] = addr
		this.tempPeers[addrString] = client
		return client, nil
	}

}

// GetPeerByHash
func (this *PeersPoolIml) GetPeerByHash(hash string) *Peer{
	if _,ok := this.peerAddr[hash];ok{
		peerAddr := this.peerAddr[hash]
		return this.peers[peerAddr]
	}
	return nil
}

// GetPeer get a peer point by the peer address
func (this *PeersPoolIml) GetPeer(addr pb.PeerAddr) *Peer {
	if clientName, ok := this.peerKeys[addr]; ok {
		client := this.peers[clientName]
		return client
	} else {
		return nil
	}
}



// GetAliveNodeNum get all alive node num
func (this *PeersPoolIml) GetAliveNodeNum() int {
	return this.alivePeers
}


// GetPeers  get peers from the peer pool
func (this *PeersPoolIml) GetPeers() []*Peer {
	var clients []*Peer
	for _, cl := range this.peers {
		clients = append(clients, cl)
		//log.Critical("取得路由表:", cl)
	}

	return clients
}

// GetPeers  get peers from the peer pool
func (this *PeersPoolIml) GetPeersWithTemp() []*Peer {
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
func (this *PeersPoolIml)ToRoutingTable() pb.Routers {
	peers := this.GetPeers()
	var routers pb.Routers

	for _, pers := range peers {
		routers.Routers = append(routers.Routers, pers.PeerAddr.ToPeerAddress())
	}
	//需要进行排序
	//sort.Sort(routers)
	return routers
}
// get routing table without specificToRoutingTableWithout hash
func (this *PeersPoolIml)ToRoutingTableWithout(hash string)pb.Routers{
	peers := this.GetPeers()
	var routers pb.Routers

	for _, pers := range peers {
		if pers.LocalAddr.Hash == hash{
			continue
		}
		routers.Routers = append(routers.Routers, pers.PeerAddr.ToPeerAddress())
	}
	//加入自己
	routers.Routers = append(routers.Routers,this.localAddr)
	//需要进行排序
	sort.Sort(routers)
	for idx,_ := range routers.Routers{
		routers.Routers[idx].ID = uint64(idx+1)
	}
	return routers
}


// merge the route into the temp peer list
func (this *PeersPoolIml)MergeFormRoutersToTemp(routers pb.Routers) {
	for _, peerAddress := range routers.Routers {
		peerAddr := pb.RecoverPeerAddr(peerAddress)
		newPeer, err := NewPeer(peerAddr,this.localAddr,this.TEM)
		if err != nil {
			log.Error("merge from routers error ", err)
		}
		this.PutPeerToTemp(*newPeer.PeerAddr, newPeer)
	}
}
// Merge the temp peer into peers list
func (this *PeersPoolIml) MergeTempPeers(peer *Peer) {
	//log.Critical("old节点合并路由表!!!!!!!!!!!!!!!!!!!")
	//使用共识结果进行更新
	//for _, tempPeer := range this.tempPeers {
	//	if tempPeer.RemoteAddr.Hash == address.Hash {
	this.peers[peer.PeerAddr.Hash] = peer
	this.peerAddr[peer.PeerAddr.Hash] = *peer.PeerAddr
	this.peerKeys[*peer.PeerAddr] = peer.PeerAddr.Hash
	delete(this.tempPeers, peer.PeerAddr.Hash)
	//}
	//}
}

func (this *PeersPoolIml) MergeTempPeersForNewNode() {
	//使用共识结果进行更新
	for _, tempPeer := range this.tempPeers {
		this.peers[tempPeer.PeerAddr.Hash] = tempPeer
		this.peerAddr[tempPeer.PeerAddr.Hash] = *tempPeer.PeerAddr
		this.peerKeys[*tempPeer.PeerAddr] = tempPeer.PeerAddr.Hash
		delete(this.tempPeers, tempPeer.PeerAddr.Hash)

	}
}

//reject the temp peer list
func (this *PeersPoolIml)RejectTempPeers() {
	for _, tempPeer := range this.tempPeers {
		delete(this.tempPeers, tempPeer.PeerAddr.Hash)
		this.alivePeers -= 1
	}
}

func (this *PeersPoolIml)DeletePeer(p *Peer){
	this.alivePeers -= 1
	p.Connection.Close()
	delete(this.peers, p.PeerAddr.Hash)
	delete(this.peerAddr, p.PeerAddr.Hash)
	delete(this.peerKeys, *p.PeerAddr)
}

