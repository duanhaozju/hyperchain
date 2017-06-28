package p2p

import (
	"fmt"
	"github.com/pkg/errors"
	"hyperchain/p2p/network"
	pb "hyperchain/p2p/message"
	"github.com/spf13/viper"
	"hyperchain/manager/event"
	"hyperchain/p2p/msg"
	"hyperchain/p2p/info"
	"hyperchain/p2p/threadsafelinkedlist"
)


type peerManagerImpl struct {
	namespace string
	// here use the concurrent map to keep thread safe
	// use this map will lose some tps, and idHostMap functional are same as the old version peersPool
	// idHostMap -> map[string]string => map[id]hostname
	hyperNet  *network.HyperNet

	node      *Node
	peerPool  *PeersPool

	eventHub  *event.TypeMux
	blackHole chan interface{}
	nodeNum   int
	selfID    int

	initType  chan int

	isonline *threadsafelinkedlist.SpinLock
	isnew bool
}

//todo rename new function
func NewPeerManagerImpl(namespace string,peercnf *viper.Viper,ev *event.TypeMux, net *network.HyperNet) (*peerManagerImpl, error) {
	if net == nil{
		return nil,errors.New("the P2P manager hasn't initlized.")
	}
	N := peercnf.GetInt("self.N")
	if N < 4 {
		return nil, errors.New(fmt.Sprintf("invalid N: %d", N))
	}
	// cnf check logic
	selfID := peercnf.GetInt("self.id")
	if selfID > N || selfID <= 0{
		return nil, errors.New(fmt.Sprintf("invalid self id: %d", selfID))
	}
	selfHostname := peercnf.GetString("self.hostname")
	isnew:= peercnf.GetBool("self.new")
	if selfHostname == ""{
		return nil, errors.New(fmt.Sprintf("invalid self hostname: %s", selfHostname))
	}

	pmi :=  &peerManagerImpl{
		namespace: namespace,
		eventHub:ev,
		hyperNet:net,
		selfID:selfID,
		peerPool:NewPeersPool(namespace),
		node:NewNode(namespace,selfID,selfHostname,net),
		nodeNum:N,
		blackHole:make(chan interface{}),
		initType:make(chan int,1),
		isonline:new(threadsafelinkedlist.SpinLock),
		isnew:isnew,
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
	fmt.Println("now vp peers pool is ")
	for _,p := range pmi.peerPool.GetPeers(){
		fmt.Println(p.info.GetID())
	}
	return pmi,nil
}

// initialize the peerManager which is for init the local node
func (pmgr *peerManagerImpl)Start() error{
	//todo this for test
	pmgr.Listening()
	sessionHandler := msg.NewSessionHandler(pmgr.blackHole,pmgr.eventHub)
	helloHandler := msg.NewHelloHandler(pmgr.blackHole,pmgr.eventHub)
	attendHandler := msg.NewAttendHandler(pmgr.blackHole,pmgr.eventHub)
	pmgr.node.Bind(pb.MsgType_SESSION,sessionHandler)
	pmgr.node.Bind(pb.MsgType_HELLO,helloHandler)
	pmgr.node.Bind(pb.MsgType_ATTEND,attendHandler)
	pmgr.initType <- START_NORMAL
	pmgr.SetOnline()
	//new attend Process
	if pmgr.isnew{
		//this should wait until all nodes reverse connect to self.
		pmgr.broadcast(pb.MsgType_ATTEND,[]byte(pmgr.GetLocalAddressPayload()))
		pmgr.eventHub.Post(event.AlreadyInChainEvent{})
	}
	return nil
}

func (pmgr *peerManagerImpl)Listening(){
	fmt.Println("hello im listening msg")
	go func(){
		for ev := range pmgr.blackHole {
			fmt.Printf("OUTER GOT A Message %v \n",ev)
		}
	}()
}

//bind the namespace+id -> hostname
func (pmgr *peerManagerImpl) bind(namespace string, id int, hostname string) error {
	newPeer := NewPeer(namespace,hostname,id,pmgr.node.info,pmgr.hyperNet)
	err := pmgr.peerPool.AddVPPeer(id,newPeer)
	if err != nil{
		return err
	}
	return nil
}

//unBindAll clients, because of error occur.
func (pmgr *peerManagerImpl) unBindAll() error {
	peers := pmgr.peerPool.GetPeers()
	for _,p:=range peers{
		pmgr.peerPool.DeleteVPPeer(p.info.GetID())
	}
	return nil
}

func (pmgr *peerManagerImpl) SendMsg(payload []byte, peers []uint64) {
	pmgr.sendMsg(pb.MsgType_SESSION,payload,peers)
}

//SendMsg send  message to specific hosts
func (pmgr *peerManagerImpl) sendMsg(msgType pb.MsgType,payload []byte, peers []uint64) {
	if !pmgr.isonline.IsLocked(){
		return
	}
	//TODO here can be improved, such as pre-calculate the peers' hash
	//TODO utils.GetHash will new a hasher every time, this waste of time.
	peerList := pmgr.peerPool.GetPeers()
	size := len(peerList)
	for _,id := range peers {
		//REVIEW here should ensure >= to avoid index out of range
		if int(id-1) >= size {
			continue
		}
		if id == uint64(pmgr.node.info.GetID()){
			continue
		}
		peer := peerList[int(id-1)]
		if peer.info.Hostname == pmgr.node.info.Hostname{
			continue
		}
		msg := pb.NewMsg(msgType,payload)
		peer.Chat(msg)
	}

}

func (pmgr *peerManagerImpl)Broadcast(payload []byte){
	pmgr.broadcast(pb.MsgType_SESSION,payload)
}
//Broadcast message to all binding hosts
func (pmgr *peerManagerImpl) broadcast(msgType pb.MsgType,payload []byte) {
	if !pmgr.isonline.IsLocked(){
		return
	}
	// use IterBuffered for better performance
	peerList := pmgr.peerPool.GetPeers()
	for _,p := range peerList{
		go func(peer *Peer){
			if peer.hostname == peer.local.Hostname{
				return
			}
			// this is un thread safe, because this,is a pointer
			msg := pb.NewMsg(msgType,payload)
			_,err := peer.Chat(msg)
			if err != nil{
				fmt.Printf("hostname [%s](%s) chat err: send self %s \n",peer.hostname,peer.local.Hostname,err.Error())
			}
		}(p)

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
	for _,peer := range pmgr.peerPool.GetPeers(){
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

// DeleteNode interface
func (pmgr *peerManagerImpl)GetLocalNodeHash() string {
	return pmgr.node.info.GetHash()
}
//GetVPPeers return all vp peers
func (pmgr *peerManagerImpl)GetVPPeers() []*Peer {
	return pmgr.peerPool.GetPeers()
}

/////////////////////////////
// TODO UnImplements
////////////////////////////
func (pmgr *peerManagerImpl)Stop(){
	//TODO
}

func (pmgr *peerManagerImpl)GetInitType() <-chan int{
	c := make(chan int)
	return c
}

// AddNode
// update routing table when new peer's join request is accepted
func (pmgr *peerManagerImpl)UpdateRoutingTable(payLoad []byte){
	//unmarshal info
	i := info.InfoUnmarshal(payLoad)
	fmt.Println(i.Hostname)
	fmt.Println(i.Namespace)
	fmt.Println("update the route table")
	err := pmgr.bind(i.Namespace,i.Id,i.Hostname)
	if err !=nil{
		fmt.Errorf("cannot bind a new peer: %s",err.Error())
		return
	}

	for _,p := range pmgr.peerPool.GetPeers(){
		fmt.Println("update table", p.hostname)
	}
}
//This method is DEPRECIATE.
func (pmgr *peerManagerImpl)UpdateAllRoutingTable(routerPayload []byte){
	//because new add node, it needn't negotiate view
	return
}

func (pmgr *peerManagerImpl)GetLocalAddressPayload() []byte{
	return pmgr.node.info.Serialize()
}

func (pmgr *peerManagerImpl)SetOnline(){
	pmgr.isonline.TryLock()
}

func (pmgr *peerManagerImpl)GetRouterHashifDelete(hash string) (string, uint64, uint64){
	return "",1,1
}

func (pmgr *peerManagerImpl)DeleteNode(hash string) error { // if self {...} else{...}
	return nil
}

// InfoGetter get the peer info to manager
// get the all peer list to broadcast
func (pmgr *peerManagerImpl)GetAllPeers() []*Peer {
	return pmgr.peerPool.GetPeers()
}

// get local node instance
//GetLocalNode() *Node
// Get local node id
func (pmgr *peerManagerImpl)GetNodeId() int{
	return pmgr.node.info.GetID()
}

//get the peer information of all nodes.
func (pmgr *peerManagerImpl)GetPeerInfo() PeerInfos{
	return PeerInfos{}
}

// use by new peer when join the chain dynamically only
func (pmgr *peerManagerImpl)GetRouters() []byte{
	b,e := pmgr.peerPool.Serlize()
	if e != nil{
		fmt.Errorf("cannot serialize the peerpool,err:%s \n",e.Error())
		return nil
	}
	return b
}


