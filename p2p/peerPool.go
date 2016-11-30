//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package p2p

import (
	pb "hyperchain/p2p/peermessage"
	"hyperchain/p2p/transport"
	"strconv"
	"strings"
	"errors"
	"hyperchain/p2p/peerComm"
	"google.golang.org/grpc/peer"
)

type PeersPool struct {
	peers        map[string]*Peer
	peerAddr     map[string]pb.PeerAddress
	peerKeys     map[pb.PeerAddress]string
	tempPeers    map[string]*Peer
	tempPeerAddr map[string]pb.PeerAddress
	tempPeerKeys map[pb.PeerAddress]string
	TEM          transport.TransportEncryptManager
	alivePeers   int
	localNode    pb.PeerAddress
}

// the peers pool instance
var prPoolIns PeersPool

// NewPeerPool get a new peer pool instance
func NewPeerPool(TEM transport.TransportEncryptManager) *PeersPool {
	var newPrPoolIns PeersPool
	newPrPoolIns.peers = make(map[string]*Peer)
	newPrPoolIns.peerAddr = make(map[string]pb.PeerAddress)
	newPrPoolIns.peerKeys = make(map[pb.PeerAddress]string)

	newPrPoolIns.tempPeers = make(map[string]*Peer)
	newPrPoolIns.tempPeerAddr = make(map[string]pb.PeerAddress)
	newPrPoolIns.tempPeerKeys = make(map[pb.PeerAddress]string)

	newPrPoolIns.TEM = TEM
	newPrPoolIns.alivePeers = 0
	return &newPrPoolIns

}

// PutPeer put a peer into the peer pool and get a peer point
func (this *PeersPool) PutPeer(addr pb.PeerAddress, client *Peer) (*Peer, error) {
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
func (this *PeersPool) PutPeerToTemp(addr pb.PeerAddress, client *Peer) (*Peer, error) {
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




// GetPeer get a peer point by the peer address
func (this *PeersPool) GetPeer(addr pb.PeerAddress) *Peer {
	if clientName, ok := this.peerKeys[addr]; ok {
		client := this.peers[clientName]
		return client
	} else {
		return nil
	}
}



// GetAliveNodeNum get all alive node num
func (this *PeersPool) GetAliveNodeNum() int {
	return this.alivePeers
}

// GetPeerByString get peer by address string
func GetPeerByString(addr string) *Peer {
	address := strings.Split(addr, ":")
	p, err := strconv.Atoi(address[1])
	if err != nil {
		log.Error(`given string is not like "localhost:1234", pls check it`)
		return nil
	}
	pAddr := pb.PeerAddress{
		IP:   address[0],
		Port: int64(p),
	}
	if peerAddr, ok := prPoolIns.peerAddr[pAddr.String()]; ok {
		return prPoolIns.peers[prPoolIns.peerKeys[peerAddr]]
	} else {
		return nil
	}
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

// DelPeer delete a peer by the given peer address (Not Used)
func DelPeer(addr pb.PeerAddress) {
	delete(prPoolIns.peers, prPoolIns.peerKeys[addr])
}

//将peerspool转换成能够传输的列表
func (this *PeersPool)ToRoutingTable() pb.Routers {
	peers := this.GetPeers()
	var routers pb.Routers

	for _, pers := range peers {
		routers.Routers = append(routers.Routers, pers.RemoteAddr)
	}
	//需要进行排序
	//sort.Sort(routers)
	return routers
}
// get routing table with specific hash
func (this *PeersPool)ToRoutingTableWithout(hash string)pb.Routers{
	peers := this.GetPeers()
	var routers pb.Routers

	for _, pers := range peers {
		if pers.Addr.Hash == hash{
			continue
		}
		routers.Routers = append(routers.Routers, pers.RemoteAddr)
	}
	//需要进行排序
	//sort.Sort(routers)
	return routers
}


// merge the route into the temp peer list
func (this *PeersPool)MergeFormRoutersToTemp(routers pb.Routers) {
	for _, peerAddress := range routers.Routers {
		newPeer, err := NewPeerByIpAndPort(peerAddress.IP, peerAddress.Port, uint64(this.alivePeers + 1), this.TEM, &this.localNode,this)
		if err != nil {
			log.Error("merge from routers error ", err)
		}
		this.PutPeerToTemp(*newPeer.RemoteAddr, newPeer)
	}
}
// Merge the temp peer into peers list
func (this *PeersPool) MergeTempPeers(peer *Peer) {
	//log.Critical("old节点合并路由表!!!!!!!!!!!!!!!!!!!")
	//使用共识结果进行更新
	//for _, tempPeer := range this.tempPeers {
	//	if tempPeer.RemoteAddr.Hash == address.Hash {
	this.peers[peer.RemoteAddr.Hash] = peer
	this.peerAddr[peer.RemoteAddr.Hash] = *peer.RemoteAddr
	this.peerKeys[*peer.RemoteAddr] = peer.RemoteAddr.Hash
	delete(this.tempPeers, peer.RemoteAddr.Hash)
	//}
	//}
}

func (this *PeersPool) MergeTempPeersForNewNode() {
	//使用共识结果进行更新
	for _, tempPeer := range this.tempPeers {
		this.peers[tempPeer.RemoteAddr.Hash] = tempPeer
		this.peerAddr[tempPeer.RemoteAddr.Hash] = *tempPeer.RemoteAddr
		this.peerKeys[*tempPeer.RemoteAddr] = tempPeer.RemoteAddr.Hash
		delete(this.tempPeers, tempPeer.RemoteAddr.Hash)

	}
}

//reject the temp peer list
func (this *PeersPool)RejectTempPeers() {
	for _, tempPeer := range this.tempPeers {
		delete(this.tempPeers, tempPeer.RemoteAddr.Hash)
		this.alivePeers -= 1
	}
}

func (this *PeersPool)DeletePeer(p *Peer){
	this.alivePeers -= 1
	delete(this.peers, p.RemoteAddr.Hash)
	delete(this.peerAddr, p.RemoteAddr.Hash)
	delete(this.peerKeys, p.RemoteAddr)
}
