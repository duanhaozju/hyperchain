// author: chenquan
// date: 16-8-25
// last modified: 16-8-29 13:23
// last Modified Author: chenquan
// change log: given a test param into the new peer pool, this is the convenience to test
//
package peerPool

import (
	"errors"
	peer "hyperchain/p2p/peer"
	pb "hyperchain/p2p/peermessage"
	"time"
	"strings"
	"strconv"
	"github.com/op/go-logging"
)

type PeersPool struct {
	peers      map[string]*peer.Peer
	peerAddr   map[string]pb.PeerAddress
	peerKeys   map[pb.PeerAddress]string
	aliveNodes int
}

// the peers pool instance
var prPoolIns PeersPool
var log *logging.Logger // package-level logger

//initialize the peers pool
func init() {
	log = logging.MustGetLogger("p2p/peerPool")
	prPoolIns.peers = make(map[string]*peer.Peer)
	prPoolIns.peerAddr = make(map[string]pb.PeerAddress)
	prPoolIns.peerKeys = make(map[pb.PeerAddress]string)
	prPoolIns.aliveNodes = 0
}

// NewPeerPool get a new peer pool instance
func NewPeerPool(isNewInstance bool,isKeepAlive bool) PeersPool {
	if isNewInstance {
		var newPrPoolIns PeersPool
		newPrPoolIns.peers = make(map[string]*peer.Peer)
		newPrPoolIns.peerAddr = make(map[string]pb.PeerAddress)
		newPrPoolIns.peerKeys = make(map[pb.PeerAddress]string)
		//Open a keep alive go routine
		//Set interval to post keep alive event to event manager
		/////////////////////////////////////////////////////
		//                                                 //
		//       THIS PART IS USED FOR KEEP ALIVE          //
		//                                                 //
		/////////////////////////////////////////////////////
		if isKeepAlive{
			go func() {
				for tick := range time.Tick(15 * time.Second) {
					log.Debug("Keep alive go routine information")
					log.Debug("Keep alive:", tick)
					for nodeName, p := range newPrPoolIns.peers {
						log.Info(nodeName)
						msg, err := p.Chat(&pb.Message{
							MessageType:  pb.Message_HELLO,
							Payload:      []byte("Hello"),
							MsgTimeStamp: time.Now().UnixNano(),
						})
						if err != nil {
							log.Fatal("Node:", p.Addr, "was down and need to recall ", err)
							//TODO RECALL THE DEAD NODE
						}
						log.Info("Node:", p.Addr, "Message", msg.MessageType)
					}
				}
			}()
		}
		return newPrPoolIns
	} else {
		return prPoolIns
	}
}

// PutPeer put a peer into the peer pool and get a peer point
func (this *PeersPool) PutPeer(addr pb.PeerAddress, client *peer.Peer) (*peer.Peer, error) {
	addrString := addr.String()
	//log.Println("Add a peer:",addrString)
	if _, ok := this.peerKeys[addr]; ok {
		// the pool already has this client
		return this.peers[addrString], errors.New("The client already in")
	} else {
		this.aliveNodes += 1
		this.peerKeys[addr] = addrString
		this.peerAddr[addrString] = addr
		this.peers[addrString] = client
		return client, nil
	}

}

// GetPeer get a peer point by the peer address
func (this *PeersPool) GetPeer(addr pb.PeerAddress) *peer.Peer {
	if clientName, ok := this.peerKeys[addr]; ok {
		client := this.peers[clientName]
		return client
	} else {
		return nil
	}
}

// GetAliveNodeNum get all alive node num
func (this *PeersPool) GetAliveNodeNum() int {
	return this.aliveNodes
}

// GetPeerByString get peer by address string
func GetPeerByString(addr string)*peer.Peer{
	address := strings.Split(addr,":")
	p,err := strconv.Atoi(address[1])
	if err != nil{
		log.Error(`given string is not like "localhost:1234", pls check it`)
		return nil
	}
	pAddr := pb.PeerAddress{
		Ip:address[0],
		Port:int32(p),
	}
	if peerAddr,ok := prPoolIns.peerAddr[pAddr.String()];ok{
		return prPoolIns.peers[prPoolIns.peerKeys[peerAddr]]
	}else{
		return nil
	}
}

// GetPeers  get peers from the peer pool
func (this *PeersPool) GetPeers() []*peer.Peer {
	var clients []*peer.Peer
	for _, cl := range this.peers {
		clients = append(clients, cl)
	}
	return clients
}

// DelPeer delete a peer by the given peer address (Not Used)
func DelPeer(addr pb.PeerAddress) {
	delete(prPoolIns.peers, prPoolIns.peerKeys[addr])
}
