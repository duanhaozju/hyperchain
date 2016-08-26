// author: chenquan
// date: 16-8-25
// last modified: 16-8-25 20:01
// last Modified Author: chenquan
// change log:
//
package peerPool

import (
	peer "hyperchain-alpha/p2p/peer"
	pb "hyperchain-alpha/p2p/peermessage"
	"errors"
	"fmt"
	"time"
	"log"
)


type PeersPool struct {
	peers    map[string]*peer.Peer
	peerAddr map[string]pb.PeerAddress
	peerKeys map[pb.PeerAddress]string
	aliveNodes int
}
// the peers pool instance
var prPoolIns PeersPool

//initialize the peers pool
func init(){
	prPoolIns.peers = make(map[string]*peer.Peer)
	prPoolIns.peerAddr = make(map[string]pb.PeerAddress)
	prPoolIns.peerKeys = make(map[pb.PeerAddress]string)
	prPoolIns.aliveNodes = 0
	//Open a keep alive go routine
	//Set interval to post keep alive event to event manager
	go func(){
		for tick := range time.Tick(15 * time.Second){
			log.Println("Keep alive go routine information")
			log.Println("Keep alive:",tick)
			for nodeName,p := range prPoolIns.peers{
				log.Println(nodeName)
				msg,err := p.Chat(&pb.Message{
					MessageType:pb.Message_HELLO,
					Payload:[]byte("Hello"),
					MsgTimeStamp:time.Now().Unix(),
				})
				if err != nil{
					log.Println("Node:",p.Addr,"is dead and need to reconnect ",err )
					//TODO RECALL THE DEAD NODE
					prPoolIns.aliveNodes -= 1
				}else{
					log.Println("Node:",p.Addr,"Message",msg.MessageType)
					prPoolIns.aliveNodes += 1
				}

			}
		}
	}()
}


func NewPeerPool(isNewInstance bool) PeersPool {
	if isNewInstance{
		var newPrPoolIns PeersPool
		newPrPoolIns.peers = make(map[string]*peer.Peer)
		newPrPoolIns.peerAddr = make(map[string]pb.PeerAddress)
		newPrPoolIns.peerKeys = make(map[pb.PeerAddress]string)
		//Open a keep alive go routine
		//Set interval to post keep alive event to event manager
		go func(){
			for tick := range time.Tick(15 * time.Second){
				log.Println("Keep alive go routine information")
				log.Println("Keep alive:",tick)
				for nodeName,p := range newPrPoolIns.peers{
					log.Println(nodeName)
					msg,err := p.Chat(&pb.Message{
						MessageType:pb.Message_HELLO,
						Payload:[]byte("Hello"),
						MsgTimeStamp:time.Now().Unix(),
					})
					if err != nil{
						log.Fatal("Node:",p.Addr,"is dead need to recall ",err )
						//TODO RECALL THE DEAD NODE
					}
					log.Println("Node:",p.Addr,"Message",msg.MessageType)
				}
			}
		}()
		return newPrPoolIns
	}else{
		return prPoolIns
	}
}

func (this *PeersPool)PutPeer(addr pb.PeerAddress,client *peer.Peer)( *peer.Peer,error){
	addrString := addr.String()
	fmt.Println(addrString)
	if _,ok := this.peerKeys[addr];ok{
		// the pool already has this client
		return this.peers[addrString],errors.New("The client already in")
	}else{
		this.peerKeys[addr] = addrString
		this.peerAddr[addrString]=addr
		this.peers[addrString]=client
		return client,nil
	}

}

func (this *PeersPool)GetPeer(addr pb.PeerAddress) *peer.Peer{
	if clientName,ok := this.peerKeys[addr];ok{
		client := this.peers[clientName]
		return client
	}else{
		return nil
	}
}

func (this *PeersPool)GetAliveNodeNum() int{
	return this.aliveNodes
}


// GetPeerByString get peer by address string
//func GetPeerByString(addr string)*client.ChatClient{
//	address := strings.Split(addr,":")
//	p,err := strconv.Atoi(address[1])
//	if err != nil{
//		log.Fatalln("the string is not like localhost:8888")
//	}
//	pAddr := pb.PeerAddress{
//		Ip:address[0],
//		Port:int32(p),
//	}
//	if peerAddr,ok := prPoolIns.peerAddr[pAddr.String()];ok{
//		return prPoolIns.peers[prPoolIns.peerKeys[peerAddr]]
//	}else{
//		return nil
//	}
//}
// GetPeers  get peers from the peer pool
func (this *PeersPool)GetPeers()[]*peer.Peer{
	var clients []*peer.Peer
	for _,cl := range this.peers {
		clients = append(clients,cl)
	}
	return clients
}

func DelPeer(addr pb.PeerAddress){
	delete(prPoolIns.peers,prPoolIns.peerKeys[addr])
}
