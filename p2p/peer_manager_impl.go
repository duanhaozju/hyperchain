package p2p

import (
	"fmt"
	"github.com/orcaman/concurrent-map"
	"github.com/pkg/errors"
	"hyperchain/p2p/network"
	"hyperchain/p2p/utils"
	"hyperchain/p2p/message"
	"github.com/spf13/viper"
	"hyperchain/manager/event"
)

type peerManagerImpl struct {
	namespace string
	// here use the concurrent map to keep thread safe
	// use this map will lose some tps, and idHostMap functional are same as the old version peersPool
	// idHostMap -> map[string]string => map[id]hostname
	idHostMap cmap.ConcurrentMap
	hyperNet  *network.HyperNet

	node      *Node
	peerPool  *PeersPool

	eventHub  *event.TypeMux
}
//todo rename new function
func NewPeerManagerImpl(namespace string,peercnf *viper.Viper,ev *event.TypeMux, net *network.HyperNet) (*peerManagerImpl, error) {
	pmi :=  &peerManagerImpl{
		namespace:       namespace,
		idHostMap:     cmap.New(),
		eventHub:ev,
		hyperNet:net,
	}
	nodemaps  := peercnf.Get("nodes").([]interface{})
	for _,item := range nodemaps{
		var err error
		var id int
		var hostname string
		node  := item.(map[interface{}]interface{})
		for key,value := range node{
			if key.(string) == "id"{
				id  = value.(int)
			}
			if key.(string) == "hostname"{
				hostname = value.(string)
			}
		}
		err = pmi.bind(namespace,id,hostname)
		if err != nil{
			fmt.Errorf("cannot bind client,err: %v",err)
			//this is for atomic bind operation, if a error occurs,
			//that should unbind all the already binding clients.
			pmi.unBindAll()
			return nil,err
		}
	}

	return pmi,nil
}

//bind the namespace+id -> hostname
func (pmgr *peerManagerImpl) bind(namespace string, id int, hostname string) error {
	hash := utils.GetPeerHash(namespace, id)
	if _, ok := pmgr.idHostMap.Get(hash); ok {
		return errors.New("this peeer already has been binded.")
	}
	pmgr.idHostMap.Set(hash, hostname)
	return nil
}

//unBindAll clients, because of error occur.
func (pmgr *peerManagerImpl) unBindAll() error {
	if pmgr.idHostMap.Count() == 0{
		return errors.New("the bind table is nil")
	}
	items := pmgr.idHostMap.IterBuffered()
	for item := range items{
		pmgr.idHostMap.Remove(item.Key)
	}
	return nil
}

//SendMsg send  message to specific hosts
func (pmgr *peerManagerImpl) SendMsg(payload []byte, peers []uint64) {
	//TODO here can be improved, such as pre-calculate the peers' hash
	//TODO utils.GetHash will new a hasher every time, this waste of time.
	for _, id := range peers {
		hash := utils.GetPeerHash(pmgr.namespace, int(id))
		// review here send to client should be thread safe
		hostname, ok := pmgr.idHostMap.Get(hash)
		if !ok {
			fmt.Errorf("cannot get the peer for peer %d", id)
		}

		msg := &message.Message{
			MessageType:message.Message_SESSION,
			Payload:payload,
		}
		// TODO Here should change as chat method
		pmgr.hyperNet.Greeting(hostname.(string),msg)
	}
}

//Broadcast message to all binding hosts
func (pmgr *peerManagerImpl) Broadcast(payload []byte) {
	// use IterBuffered for better performance
	for item := range pmgr.idHostMap.IterBuffered(){
		hostname := item.Val.(string)
		msg := &message.Message{
			MessageType:message.Message_SESSION,
			Payload:payload,
		}
		pmgr.hyperNet.Greeting(hostname,msg)
	}
}

// set peer managers primary peer and node
func (pmgr *peerManagerImpl)SetPrimary(_id uint64) error{
	//review here conversation is not safe
	id := int(_id)
	flag := false
	if pmgr.node.info.GetID() == id{
		pmgr.node.info.SetPrimary(true)
		flag = true
	}
	for _,peer := range pmgr.peerPool.GetIterator(){
		if peer.info.GetID() == id {
			flag = true
			peer.info.SetPrimary(true)
		}else{
			peer.info.SetPrimary(false)
		}
	}
	if !flag {
		return errors.New("invalid peer id, not found any suitable peer")
	}
	return nil
}

// initialize the peerManager which is for init the local node
func (pmgr *peerManagerImpl)Start() error{
	return nil
}
func (pmgr *peerManagerImpl)Stop(){}
func (pmgr *peerManagerImpl)GetInitType() <-chan int{
	c := make(chan int)
	return c
}


// AddNode
// update routing table when new peer's join request is accepted
func (pmgr *peerManagerImpl)UpdateRoutingTable(payLoad []byte){
	return
}
func (pmgr *peerManagerImpl)UpdateAllRoutingTable(routerPayload []byte){
	return
}
func (pmgr *peerManagerImpl)GetLocalAddressPayload() []byte{
	return nil
}
func (pmgr *peerManagerImpl)SetOnline(){
	return
}

// DeleteNode interface
func (pmgr *peerManagerImpl)GetLocalNodeHash() string{
	return ""
}
func (pmgr *peerManagerImpl)GetRouterHashifDelete(hash string) (string, uint64, uint64){
	return "",1,1
}
func (pmgr *peerManagerImpl)DeleteNode(hash string) error { // if self {...} else{...}
	return nil
}

// InfoGetter get the peer info to manager
// get the all peer list to broadcast
func (pmgr *peerManagerImpl)GetAllPeers() []*peer {
	return nil
}
func (pmgr *peerManagerImpl)GetAllPeersWithTemp() []*peer {
	return nil
}
func (pmgr *peerManagerImpl)GetVPPeers() []*peer {
	return nil
}
// get local node instance
//GetLocalNode() *Node
// Get local node id
func (pmgr *peerManagerImpl)GetNodeId() int{
	return 1;
}
//get the peer information of all nodes.
func (pmgr *peerManagerImpl)GetPeerInfo() PeerInfos{
	return PeerInfos{}
}

// use by new peer when join the chain dynamically only
func (pmgr *peerManagerImpl)GetRouters() []byte{
	return nil
}
