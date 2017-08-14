package p2p

import (
	"github.com/pkg/errors"
	"hyperchain/p2p/network"
	pb "hyperchain/p2p/message"
	"github.com/terasum/viper"
	"hyperchain/manager/event"
	"hyperchain/p2p/msg"
	"hyperchain/p2p/info"
	"hyperchain/p2p/threadsafe"
	"hyperchain/common"
	"github.com/op/go-logging"
	"time"
	"hyperchain/p2p/random_stack"
	"sort"
	"hyperchain/p2p/hts"
	"hyperchain/p2p/hts/secimpl"
	"github.com/orcaman/concurrent-map"
	"hyperchain/p2p/peerevent"
	"strconv"
	"reflect"
	"fmt"
)

const(
	PEERTYPE_VP = iota
	PEERTYPE_NVP
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

	isonline  *threadsafe.SpinLock

	isnew     bool

	isVP      bool
	isOrg      bool

	n     int

	isRec      bool

	delchan   chan bool

	logger    *logging.Logger

	pts       *PeerTriples
	// this is for persist the config
	peercnf *peerCnf

	hts *hts.HTS

	//connection controller typemux
	peerMgrEv *event.TypeMux
	peerMgrEvClose chan interface{}
	peerMgrSub cmap.ConcurrentMap

}

//todo rename new function
func NewPeerManagerImpl(namespace string, peercnf *viper.Viper, ev *event.TypeMux, net *network.HyperNet, delChan chan bool) (*peerManagerImpl, error) {
	logger := common.GetLogger(namespace, "p2p")
	if net == nil {
		return nil, errors.New("the P2P manager hasn't initlized.")
	}
	isvp := peercnf.GetBool("self.vp")
	N := peercnf.GetInt("self.N")
	if isvp && N < 4 {
		return nil, errors.New(fmt.Sprintf("invalid N: %d", N))
	}
	// cnf check logic
	selfID := peercnf.GetInt("self.id")
	if isvp && (selfID - 1 > N || selfID <= 0) {
		return nil, errors.New(fmt.Sprintf("invalid self id: %d", selfID))
	}
	if !isvp && selfID != 0{
		return nil, errors.New(fmt.Sprintf("invalid self id: %d, nvp id should be zero", selfID))
	}
	selfHostname := peercnf.GetString("self.hostname")
	isnew := peercnf.GetBool("self.new")
	isorg := peercnf.GetBool("self.org")
	isrec := peercnf.GetBool("self.rec")

	if selfHostname == "" {
		return nil, errors.New(fmt.Sprintf("invalid self hostname: %s", selfHostname))
	}
	caconf := common.GetPath(namespace, peercnf.GetString("self.caconf"))
	if !common.FileExist(caconf){
		return nil,errors.New(fmt.Sprintf("caconfig file is not exist, please check it %s \n",caconf))
	}
	h,err := hts.NewHTS(namespace, secimpl.NewSecuritySelector(caconf),caconf)
	if err != nil{
		return nil, errors.New(fmt.Sprintf("hts initlized failed: %s", err.Error()))
	}
	pmi := &peerManagerImpl{
		namespace: namespace,
		eventHub:ev,
		hyperNet:net,
		selfID:selfID,
		node:NewNode(namespace, selfID, selfHostname, net),
		nodeNum:N,
		blackHole:make(chan interface{}),
		peerMgrEv:new(event.TypeMux),
		peerMgrSub:cmap.New(),
		isonline:new(threadsafe.SpinLock),
		isnew:isnew,
		isOrg:isorg,
		isRec:isrec,
		delchan:delChan,
		logger: logger,
		isVP:isvp,
		n:N,
		hts:h,
		peercnf:newPeerCnf(peercnf),
	}
	nodes := peercnf.Get("nodes").([]interface{})
	pmi.pts = QuickParsePeerTriples(pmi.namespace, nodes)
	//peer pool
	pmi.peerPool = NewPeersPool(namespace,pmi.peerMgrEv,pmi.pts,pmi.peercnf)
	//set vp information
	if !pmi.isVP {
		pmi.node.info.SetNVP()
	}

	//original
	if pmi.isOrg{
		pmi.node.info.SetOrg()
	}
	//TODO org and rec cannot both be true
	//reconnect
	if pmi.isRec{
		pmi.node.info.SetRec()
	}
	pmi.isonline.TryLock()
	sort.Sort(pmi.pts)
	pmi.binding()
	//ReView After success start up, config org should be set false ,and rec should be set to true
	pmi.peercnf.Lock()
	pmi.peercnf.vip.Set("self.org",false)
	pmi.peercnf.vip.Set("self.new",false)
	pmi.peercnf.vip.Set("self.rec",true)
	pmi.peercnf.vip.WriteConfig()
	pmi.peercnf.Unlock()
	return pmi, nil
}

func(pmi *peerManagerImpl)binding()error{
	serverHTS,err := pmi.hts.GetServerHTS(pmi.peerMgrEv)
	if err != nil{
		return err
	}
	clientHelloHandler := msg.NewClientHelloHandler(serverHTS,pmi.peerMgrEv,pmi.node.info,pmi.isOrg,pmi.logger)
	pmi.node.Bind(pb.MsgType_CLIENTHELLO, clientHelloHandler)

	clientAcceptHandler := msg.NewClientAcceptHandler(serverHTS,pmi.logger)
	pmi.node.Bind(pb.MsgType_CLIENTACCEPT, clientAcceptHandler)

	sessionHandler := msg.NewSessionHandler(pmi.blackHole, pmi.eventHub,serverHTS,pmi.logger)
	pmi.node.Bind(pb.MsgType_SESSION, sessionHandler)

	helloHandler := msg.NewHelloHandler(pmi.blackHole, pmi.eventHub)
	pmi.node.Bind(pb.MsgType_HELLO, helloHandler)

	attendHandler := msg.NewAttendHandler(pmi.blackHole, pmi.eventHub,serverHTS)
	pmi.node.Bind(pb.MsgType_ATTEND, attendHandler)

	nvpAttendHandler := msg.NewNVPAttendHandler(pmi.blackHole, pmi.eventHub)
	pmi.node.Bind(pb.MsgType_NVPATTEND, nvpAttendHandler)

	nvpDeleteHandler := msg.NewNVPDeleteHandler(pmi.blackHole, pmi.eventHub,pmi.peerMgrEv,serverHTS)
	pmi.node.Bind(pb.MsgType_NVPDELETE,nvpDeleteHandler)

	//peer manager event subscribe
	pmi.peerMgrSub.Set(strconv.Itoa(peerevent.EV_VPCONNECT),pmi.peerMgrEv.Subscribe(peerevent.VPConnect{}))
	pmi.peerMgrSub.Set(strconv.Itoa(peerevent.EV_NVPCONNECT),pmi.peerMgrEv.Subscribe(peerevent.NVPConnect{}))
	pmi.peerMgrSub.Set(strconv.Itoa(peerevent.EV_VPDELETE),pmi.peerMgrEv.Subscribe(peerevent.DELETE_VP{}))
	pmi.peerMgrSub.Set(strconv.Itoa(peerevent.EV_NVPDELETE),pmi.peerMgrEv.Subscribe(peerevent.DELETE_NVP{}))
	pmi.peerMgrSub.Set(strconv.Itoa(peerevent.EV_UPDATESESSIONKey),pmi.peerMgrEv.Subscribe(peerevent.UPDATE_SESSION_KEY{}))

	return nil
}


// initialize the peerManager which is for init the local node
func (pmgr *peerManagerImpl)Start() error {
	if e := pmgr.binding();e!= nil{
		return e
	}
	pmgr.listening()
	for pmgr.pts.HasNext() {
		pt := pmgr.pts.Pop()
		//Here should control the permission
		err := pmgr.bind(PEERTYPE_VP,pt.namespace, pt.id, pt.hostname,"")
		if err != nil {
			pmgr.logger.Errorf("cannot bind client,err: %v", err)
			//this is for atomic bind operation, if a error occurs,
			//that should unbind all the already binding clients.
			pmgr.unBindAll()
			return err
		}
	}
	pmgr.SetOnline()
	//new attend Process
	if pmgr.isnew && pmgr.isVP {
		//this should wait until all nodes reverse connect to self.
		pmgr.broadcast(pb.MsgType_ATTEND, []byte(pmgr.GetLocalAddressPayload()))
		pmgr.eventHub.Post(event.AlreadyInChainEvent{})
	} else if !pmgr.isVP {
		pmgr.broadcast(pb.MsgType_NVPATTEND, []byte(pmgr.GetLocalAddressPayload()))
	}
	pmgr.logger.Infof("SELF hash: %s", pmgr.node.info.Hash)
	// after all connection
	return nil
}

func (pmgr *peerManagerImpl)listening() {
	// every times when listening, should give a new channel
	// MUST ensure the channel is already closed or the peerMgrEvClose is nil
	pmgr.peerMgrEvClose = make(chan interface{})
	pmgr.logger.Info("PeerManager is listening the peer manager event...")
	//Listening should listening all connection request, and handle it
	for subitem := range pmgr.peerMgrSub.IterBuffered(){
		go func (closechan chan interface{},t string,s event.Subscription){
			for{
				select {
				case <- closechan:{
					pmgr.logger.Debug("Listening sub goroutine stopped.")
					return
				}
				case ev := <- s.Chan():{
					// distribute all event to handlers
					pmgr.distribute(t,ev.Data)
				}
				}
			}
		}(pmgr.peerMgrEvClose,subitem.Key,subitem.Val.(event.Subscription))
	}
}

//distribute the event and payload
func (pmgr *peerManagerImpl)distribute(t string,ev interface{}){
	switch ev.(type) {
	case peerevent.VPConnect:{
		conev := ev.(peerevent.VPConnect)
		if conev.ID > pmgr.nodeNum {
			pmgr.logger.Warning("Invalid peer connect: ",conev.ID,conev.Namespace,conev.Hostname)
			return
		}
		pmgr.logger.Notice("vp connected",conev.Hostname)
		// here how to connect to hypernet layer, generally, hypernet should already connect to
		// remote node by Inneraddr, here just need to bind the hostname.
		err := pmgr.bind(PEERTYPE_VP,conev.Namespace,conev.ID,conev.Hostname,"")
		if err != nil{
			pmgr.logger.Warningf("cannot bind the remote VP hostname: reason: %s", err.Error())
			return
		}
	}
	case peerevent.NVPConnect:{
		conev := ev.(peerevent.NVPConnect)
		err := pmgr.bind(PEERTYPE_NVP,conev.Namespace,0,conev.Hostname,conev.Hash)
		if err != nil{
			pmgr.logger.Errorf("cannot bind the remote NVP hostname: reason: %s", err.Error())
			return
		}
	}
	case peerevent.DELETE_NVP:{
		pmgr.logger.Critical("GOT a EV_DELETE_NVP")
		conev := ev.(peerevent.DELETE_NVP)
		peer := pmgr.peerPool.GetNVPByHash(conev.Hash)
		if peer == nil{
			pmgr.logger.Warningf("This NVP(%s) not connect to this VP ignored.",conev.Hash)
			return
		}
		pmgr.logger.Critical("SEND TO NVP=>>",pmgr.node.info.Hash)
		m := pb.NewMsg(pb.MsgType_NVPDELETE,common.Hex2Bytes(pmgr.node.info.Hash))
		rsp,err := peer.Chat(m)
		if err != nil{
			pmgr.logger.Errorf("cannot delete NVP peer, reason: %s ",err.Error())
		}
		if rsp != nil && rsp.Payload != nil{
			pmgr.logger.Infof("delete NVP peer, response: %s ",string(rsp.Payload))
			pmgr.peerPool.DeleteNVPPeer(conev.Hash)
		}
	}
	case peerevent.DELETE_VP:{
		pmgr.logger.Critical("GOT a EV_DELETE_VP")
		if pmgr.isVP{
			pmgr.logger.Critical("As A VP Node, this process cannot be invoked")
			return
		}
		conev := ev.(peerevent.DELETE_VP)
		pmgr.logger.Critical("NVP delete VP Peer By hash",conev.Hash)
		pmgr.peerPool.DeleteVPPeerByHash(conev.Hash)
		if !pmgr.isVP{
				if pmgr.peerPool.GetVPNum() == 0{
					pmgr.logger.Critical("ALL Validate Peer are disconnect with this Non-Validate Peer")
					pmgr.logger.Critical("This Peer Will quit automaticlly after 3 seconds")
					<- time.After(3 * time.Second)
					pmgr.delchan <- true
				}
		}

	}
	case peerevent.UPDATE_SESSION_KEY:{
		pmgr.logger.Debug("GOT a EV_UPDATE_SESSION_KEY")
		conev := ev.(peerevent.UPDATE_SESSION_KEY)
		pmgr.logger.Notice("Update the session key for",conev.NodeHash)
		if !pmgr.peerPool.Ready(){
			pmgr.logger.Error("failed to update the session key, the peers pool has not prepared yet.")
			return
		}
		peer := pmgr.peerPool.GetPeerByHash(conev.NodeHash)
		if peer == nil{
			pmgr.logger.Error("Cannot find a peer By hash",conev.NodeHash)
			return
		}
		err := peer.clientHello(false,true)
		if err != nil{
			pmgr.logger.Error("failed to update the session key",conev.NodeHash)
		}

	}
	default:
		pmgr.logger.Critical("cannot determin the event type",reflect.TypeOf(ev))

	}
}

//bind the namespace+id -> hostname
func (pmgr *peerManagerImpl) bind(peerType int,namespace string, id int, hostname string,hash string) error {
	chts,err := pmgr.hts.GetAClientHTS()
	if err != nil{
		return err
	}
	//TODO here can use the peer hash to quick search
	//the exist peer
	if peerType == PEERTYPE_VP{
		if p := pmgr.peerPool.GetPeersByHostname(hostname);p != nil{
			err := p.clientHello(false,false)
			if err != nil{
				return err
			}
			return nil
		}
	}else{
		if p:= pmgr.peerPool.GetNVPByHostname(hostname);p != nil{
			pmgr.logger.Critical("Say hello to hostname:",hostname)
			err := p.clientHello(false,false)
			if err != nil{
				return err
			}
			return nil
		}
	}
	//TODO here has a problem: generally, this method will be invoked before than
	//TODO the network persist the new hostname, it will return a error: the host hasn't been initialized.
	//TODO so the node may cannot connect to new peer, bind will be failed.
	//TODO to quick fix this: before create a new peer, sleep 1 second.
	<- time.After(time.Second)
	newPeer,err := NewPeer(namespace, hostname, id, pmgr.node.info, pmgr.hyperNet,chts)
	if err != nil {
		pmgr.logger.Errorf("cannot establish connection to %s, reason: %s ",hostname,err.Error())
		return err
	}
	if peerType == PEERTYPE_VP{
		err = pmgr.peerPool.AddVPPeer(id, newPeer)
	}else{
		err = pmgr.peerPool.AddNVPPeer(hash, newPeer)
	}
	if err != nil {
		return err
	}
	return nil
}

//unBindAll clients, because of error occur.
func (pmgr *peerManagerImpl) unBindAll() error {
	peers := pmgr.peerPool.GetPeers()
	for _, p := range peers {
		pmgr.peerPool.DeleteVPPeer(p.info.GetID())
	}
	return nil
}

func (pmgr *peerManagerImpl) SendMsg(payload []byte, peers []uint64) {
	pmgr.sendMsg(pb.MsgType_SESSION, payload, peers)
}

//SendMsg send  message to specific hosts
func (pmgr *peerManagerImpl) sendMsg(msgType pb.MsgType, payload []byte, peers []uint64) {
	if !pmgr.isOnline() || !pmgr.IsVP(){
		return
	}
	//TODO here can be improved, such as pre-calculate the peers' hash
	//TODO utils.GetHash will new a hasher every time, this waste of time.
	peerList := pmgr.peerPool.GetPeers()
	size := len(peerList)
	for _, id := range peers {
		//REVIEW here should ensure `>=` to avoid index out of range
		if int(id - 1) >= size {
			continue
		}
		if id == uint64(pmgr.node.info.GetID()) {
			continue
		}
		//REVIEW  peers pool low layer struct is priority queue,
		// REVIEW this can ensure the node id order.
		// avoid out of range
		if id > uint64(len(peerList)) || id <= 0{
			return
		}
		peer := peerList[int(id)-1]
		if peer.info.Hostname == pmgr.node.info.Hostname {
			continue
		}
		m := pb.NewMsg(msgType, payload)
		peer.Chat(m)
	}

}

func (pmgr *peerManagerImpl)Broadcast(payload []byte) {
	pmgr.broadcast(pb.MsgType_SESSION, payload)
}
//Broadcast message to all binding hosts
func (pmgr *peerManagerImpl) broadcast(msgType pb.MsgType, payload []byte) {
	if !pmgr.isOnline() || !pmgr.IsVP(){
		return
	}
	// use IterBuffered for better performance
	peerList := pmgr.peerPool.GetPeers()
	for _, p := range peerList {
		go func(peer *Peer) {
			if peer.hostname == peer.local.Hostname {
				return
			}
			// this is un thread safe, because this,is a pointer
			m := pb.NewMsg(msgType, payload)
			_, err := peer.Chat(m)
			if err != nil {
				pmgr.logger.Warningf("hostname [target: %s](local: %s) chat err: send self %s ", peer.hostname, peer.local.Hostname, err.Error())
			}
		}(p)

	}
}

// set peer managers primary peer and node
func (pmgr *peerManagerImpl)SetPrimary(_id uint64) error {
	//review here conversation is not safe
	id := int(_id)
	flag := false
	if pmgr.node.info.GetID() == id {
		pmgr.node.info.SetPrimary(true)
		flag = true
	}
	for _, peer := range pmgr.peerPool.GetPeers() {
		if peer.info.GetID() == id {
			flag = true
			peer.info.SetPrimary(true)
		} else {
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
	return pmgr.node.info.Hash
}
//GetVPPeers return all vp peers
func (pmgr *peerManagerImpl)GetVPPeers() []*Peer {
	return pmgr.peerPool.GetPeers()
}

func (pmgr *peerManagerImpl)Stop() {
	close(pmgr.peerMgrEvClose)
	pmgr.logger.Criticalf("Unbind all slots...")
	pmgr.SetOffline()
	pmgr.node.UnBindAll()
}

// AddNode
// update routing table when new peer's join request is accepted
func (pmgr *peerManagerImpl)UpdateRoutingTable(payLoad []byte) {
	//unmarshal info
	i := info.InfoUnmarshal(payLoad)
	err := pmgr.bind(PEERTYPE_VP,i.Namespace, i.Id, i.Hostname,"")
	if err != nil {
		pmgr.logger.Errorf("cannot bind a new peer: %s", err.Error())
		return
	}
	pmgr.nodeNum ++
	for _, p := range pmgr.peerPool.GetPeers() {
		pmgr.logger.Info("update table", p.hostname)
	}
	err = pmgr.peerPool.PersistList()
	if err !=nil{
		pmgr.logger.Errorf("cannot persisit peer config reason: %s", err.Error())
		return
	}
}

func (pmgr *peerManagerImpl)GetLocalAddressPayload() []byte {
	return pmgr.node.info.Serialize()
}

func (pmgr *peerManagerImpl)SetOnline() {
	pmgr.isonline.UnLock()
}

func (pmgr *peerManagerImpl)isOnline()bool{
	return !pmgr.isonline.IsLocked()
}

func (pmgr *peerManagerImpl)SetOffline() {
	pmgr.isonline.TryLock()
}

//GetRouterHashifDelete returns after delete specific peer, the router table hash , self new id and the delete id
func (pmgr *peerManagerImpl)GetRouterHashifDelete(hash string) (afterDelRouterHash string, selfNewId  uint64, delID uint64) {
	afterDelRouterHash, selfNewId, delID, err := pmgr.peerPool.TryDelete(pmgr.GetLocalNodeHash(), hash)
	if err != nil {
		pmgr.logger.Errorf("cannot try del peer, error: %s", err.Error())
	}
	return
}

//DeleteNode delete the specific hash node, if the node hash is self, this node will stoped.
func (pmgr *peerManagerImpl)DeleteNode(hash string) error {
	pmgr.logger.Critical("DELETE NODE", hash)
	if pmgr.node.info.Hash == hash {
		pmgr.Stop()
		pmgr.logger.Critical(" WARNING!! THIS NODE HAS BEEN DELETED!")
		pmgr.logger.Critical(" THIS NODE WILL STOP IN 3 SECONDS")
		<-time.After(3 * time.Second)
		pmgr.logger.Critical("EXIT..")
		pmgr.delchan <- true

	}
	err :=   pmgr.peerPool.DeleteVPPeerByHash(hash)
	if err != nil{
		return err
	}
	return pmgr.peerPool.PersistList()
}

func (pmgr *peerManagerImpl)DeleteNVPNode(hash string) error {
	pmgr.logger.Critical("Delete None Validate peer hash:", hash)
	if pmgr.node.info.Hash == hash {
		pmgr.logger.Critical("Please do not send delete NVP command to nvp")
		return nil
	}
	ev := peerevent.DELETE_NVP{
		Hash:hash,
	}
	go pmgr.peerMgrEv.Post(ev)
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
func (pmgr *peerManagerImpl)GetNodeId() int {
	return pmgr.node.info.GetID()
}

func (pmgr *peerManagerImpl)GetN() int {
	return pmgr.n
}
//get the peer information of all nodes.
func (pmgr *peerManagerImpl)GetPeerInfo() PeerInfos {
	return PeerInfos{}
}

// use by new peer when join the chain dynamically only
func (pmgr *peerManagerImpl)GetRouters() []byte {
	b, e := pmgr.peerPool.Serlize()
	if e != nil {
		pmgr.logger.Errorf("cannot serialize the peerpool,err:%s ", e.Error())
		return nil
	}
	return b
}

// random select a VP and send msg to it
func (pmgr *peerManagerImpl)SendRandomVP(payload []byte) error {
	peers := pmgr.peerPool.GetPeers()
	randomStack := random_stack.NewStack()
	for _,peer :=range peers{
		randomStack.Push(peer)
	}
	var err error
	for !randomStack.Empty(){
		speer := randomStack.RandomPop().(*Peer)
		m := pb.NewMsg(pb.MsgType_SESSION, payload)
		_,err = speer.Chat(m)
		if err == nil{
			break
		}
	}
	return err
}

// broadcast information to NVP peers
func (pmgr *peerManagerImpl)BroadcastNVP(payLoad []byte) error {
	return pmgr.broadcastNVP(pb.MsgType_SESSION,payLoad)
}
func(pmgr *peerManagerImpl)broadcastNVP(msgType pb.MsgType,payload []byte)error{
	if !pmgr.isOnline() || !pmgr.IsVP(){
		return nil
	}
	// use IterBuffered for better performance
	for t := range pmgr.peerPool.nvpPool.Iter(){
		pmgr.logger.Critical("(NVP) send message to ",t.Key)
		p := t.Val.(*Peer)
		//TODO IF send failed. should return a err
		go func(peer *Peer) {
			if peer.hostname == peer.local.Hostname {
				return
			}
			// this is un thread safe, because this,is a pointer
			m := pb.NewMsg(msgType, payload)
			_, err := peer.Chat(m)
			if err != nil {
				pmgr.logger.Errorf("hostname [target: %s](local: %s) chat err: send self %s ", peer.hostname, peer.local.Hostname, err.Error())
			}
		}(p)
	}
	return nil
}

// send a message to specific NVP peer (by nvp hash) UNICAST
func (pmgr *peerManagerImpl)SendMsgNVP(payLoad []byte, nvpList []string) error {
	return pmgr.sendMsgNVP(pb.MsgType_SESSION,payLoad,nvpList)
}

func (pmgr *peerManagerImpl)sendMsgNVP(msgType pb.MsgType,payLoad []byte, nvpList []string) error {
	if !pmgr.isOnline() || !pmgr.IsVP(){
		return nil
	}
	// use IterBuffered for better performance
	for t := range pmgr.peerPool.nvpPool.Iter(){
		for _,nvphash := range nvpList{
			if nvphash != t.Key{
				continue
			}
			pmgr.logger.Critical("(VP) send message to ",t.Key)
			p := t.Val.(*Peer)
			//TODO IF send failed. should return a err
			go func(peer *Peer) {
				if peer.hostname == peer.local.Hostname {
					return
				}
				// this is un thread safe, because this,is a pointer
				m := pb.NewMsg(msgType, payLoad)
				_, err := peer.Chat(m)
				if err != nil {
					pmgr.logger.Errorf("hostname [target: %s](local: %s) chat err: send self %s ", peer.hostname, peer.local.Hostname, err.Error())
				}
			}(p)
		}
	}
	return nil
}
//IsVP return true if this node is vp node
func (pmgr *peerManagerImpl)IsVP() bool {
	return pmgr.isVP
}

