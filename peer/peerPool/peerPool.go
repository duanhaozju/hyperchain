package peerPool

import (
	"hyperchain-alpha/peer/client"
	pb "hyperchain-alpha/peer/peermessage"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"fmt"
	//"strings"
	//"strconv"
	//"log"
)


type PeersPool struct {
	peers    map[string]*client.ChatClient
	peerAddr map[string]pb.PeerAddress
	peerKeys map[pb.PeerAddress]string
}
// the peers pool instance
var prPoolIns PeersPool

//initialize the peers pool
func init(){
	prPoolIns.peers = make(map[string]*client.ChatClient)
	prPoolIns.peerAddr = make(map[string]pb.PeerAddress)
	prPoolIns.peerKeys = make(map[pb.PeerAddress]string)
}

func NewPeerPool(isNewInstance bool) PeersPool {
	if isNewInstance{
		var newPrPoolIns PeersPool
		newPrPoolIns.peers = make(map[string]*client.ChatClient)
		newPrPoolIns.peerAddr = make(map[string]pb.PeerAddress)
		newPrPoolIns.peerKeys = make(map[pb.PeerAddress]string)
		return newPrPoolIns
	}else{
		return prPoolIns
	}
}

func (pp *PeersPool)PutPeer(addr pb.PeerAddress,client *client.ChatClient)( *client.ChatClient,error){
	this := pp
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

func (pp *PeersPool)GetPeer(addr pb.PeerAddress) *client.ChatClient{
	this:=pp
	if clientName,ok := this.peerKeys[addr];ok{
		client := this.peers[clientName]
		return client
	}else{
		return nil
	}
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
//GetPeers  get peers from the peer pool
//func GetPeers()[]*client.ChatClient{
//	var clients []*client.ChatClient
//	for _,cl := range prPoolIns.peers {
//		clients = append(clients,cl)
//	}
//	return clients
//}
//func DelPeer(addr pb.PeerAddress){
//	delete(prPoolIns.peers,prPoolIns.peerKeys[addr])
//}
