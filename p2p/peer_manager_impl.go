package p2p

import (
	"fmt"
	"github.com/op/go-logging"
	"github.com/orcaman/concurrent-map"
	"github.com/pkg/errors"
	"github.com/terasum/viper"
	"hyperchain/common"
	"hyperchain/manager/event"
	"hyperchain/p2p/hts"
	"hyperchain/p2p/hts/secimpl"
	"hyperchain/p2p/info"
	pb "hyperchain/p2p/message"
	"hyperchain/p2p/msg"
	"hyperchain/p2p/network"
	"hyperchain/p2p/peerevent"
	"hyperchain/p2p/random_stack"
	"hyperchain/p2p/threadsafe"
	"math"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"time"
)

const (
	PEERTYPE_VP = iota
	PEERTYPE_NVP
)

type peerManagerImpl struct {
	namespace string
	// here use the concurrent map to keep thread safe
	// use this map will lose some tps, and idHostMap functional are same as the old version peersPool
	// idHostMap -> map[string]string => map[id]hostname
	hyperNet *network.HyperNet
	node     *Node
	peerPool *PeersPool

	eventHub  *event.TypeMux
	blackHole chan interface{}
	nodeNum   int
	selfID    int

	isonline *threadsafe.SpinLock

	isnew bool

	isVP  bool
	isOrg bool

	n int

	isRec bool

	delchan chan bool

	logger *logging.Logger

	pts *PeerTriples
	// this is for persist the config
	peercnf *peerCnf

	hts *hts.HTS

	//connection controller typemux
	peerMgrEv      *event.TypeMux
	peerMgrEvClose chan interface{}
	peerMgrSub     cmap.ConcurrentMap

	pendingChan  chan struct{}
	pendingSkMap cmap.ConcurrentMap
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
	if isvp && (selfID-1 > N || selfID <= 0) {
		return nil, errors.New(fmt.Sprintf("invalid self id: %d", selfID))
	}
	if !isvp && selfID != 0 {
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
	if !common.FileExist(caconf) {
		return nil, errors.New(fmt.Sprintf("caconfig file is not exist, please check it %s \n", caconf))
	}
	h, err := hts.NewHTS(namespace, secimpl.NewSecuritySelector(caconf), caconf)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("hts initlized failed: %s", err.Error()))
	}
	pmi := &peerManagerImpl{
		namespace:    namespace,
		eventHub:     ev,
		hyperNet:     net,
		selfID:       selfID,
		node:         NewNode(namespace, selfID, selfHostname, net),
		nodeNum:      N,
		blackHole:    make(chan interface{}),
		peerMgrEv:    new(event.TypeMux),
		peerMgrSub:   cmap.New(),
		isonline:     new(threadsafe.SpinLock),
		isnew:        isnew,
		isOrg:        isorg,
		isRec:        isrec,
		delchan:      delChan,
		logger:       logger,
		isVP:         isvp,
		n:            N,
		hts:          h,
		peercnf:      newPeerCnf(peercnf),
		pendingChan:  make(chan struct{}, 1),
		pendingSkMap: cmap.New(),
	}
	return pmi, nil
}

func (pmi *peerManagerImpl) configure() error {
	er := pmi.peercnf.viper().ReadInConfig()
	if er != nil {
		return er
	}
	nodes := pmi.peercnf.viper().Get("nodes").([]interface{})
	pts, err := QuickParsePeerTriples(pmi.namespace, nodes)
	if err != nil {
		return err
	}
	pmi.pts = pts
	//peer pool
	pmi.peerPool = NewPeersPool(pmi.namespace, pmi.peerMgrEv, pmi.pts, pmi.peercnf)
	//set vp information
	if !pmi.isVP {
		pmi.node.info.SetNVP()
	}
	//original
	if pmi.isOrg {
		pmi.node.info.SetOrg()
	}
	//reconnect
	if pmi.isRec {
		pmi.node.info.SetRec()
	}
	pmi.isonline.TryLock()
	sort.Sort(pmi.pts)
	//pmi.prepare()
	//ReView After success start up, config org should be set false ,and rec should be set to true
	pmi.peercnf.Lock()
	pmi.peercnf.vip.Set("self.org", false)
	pmi.peercnf.vip.Set("self.new", false)
	pmi.peercnf.vip.Set("self.rec", true)
	pmi.peercnf.vip.WriteConfig()
	pmi.peercnf.Unlock()
	return nil
}

func (pmi *peerManagerImpl) prepare() error {
	pmi.logger.Notice("prepare for ns", pmi.namespace)
	// MUST ensure the channel is already closed or the peerMgrEvClose is nil
	serverHTS, err := pmi.hts.GetServerHTS(pmi.peerMgrEv)
	if err != nil {
		return err
	}
	clientHelloHandler := msg.NewClientHelloHandler(serverHTS, pmi.peerMgrEv, pmi.node.info, pmi.isOrg, pmi.logger)
	pmi.node.Bind(pb.MsgType_CLIENTHELLO, clientHelloHandler)

	clientAcceptHandler := msg.NewClientAcceptHandler(serverHTS, pmi.logger)
	pmi.node.Bind(pb.MsgType_CLIENTACCEPT, clientAcceptHandler)

	sessionHandler := msg.NewSessionHandler(pmi.blackHole, pmi.eventHub, pmi.peerMgrEv, serverHTS, pmi.logger)
	pmi.node.Bind(pb.MsgType_SESSION, sessionHandler)

	helloHandler := msg.NewHelloHandler(pmi.blackHole, pmi.eventHub)
	pmi.node.Bind(pb.MsgType_HELLO, helloHandler)

	attendHandler := msg.NewAttendHandler(pmi.blackHole, pmi.eventHub, serverHTS)
	pmi.node.Bind(pb.MsgType_ATTEND, attendHandler)

	nvpAttendHandler := msg.NewNVPAttendHandler(pmi.blackHole, pmi.eventHub, pmi.peerMgrEv, pmi.logger)
	pmi.node.Bind(pb.MsgType_NVPATTEND, nvpAttendHandler)

	nvpDeleteHandler := msg.NewNVPDeleteHandler(pmi.blackHole, pmi.eventHub, pmi.peerMgrEv, serverHTS)
	pmi.node.Bind(pb.MsgType_NVPDELETE, nvpDeleteHandler)

	nvpExitHandler := msg.NewNVPExitHandler(pmi.blackHole, pmi.peerMgrEv, serverHTS)
	pmi.node.Bind(pb.MsgType_NVPEXIT, nvpExitHandler)
	//peer manager event subscribe
	pmi.peerMgrSub.Set(strconv.Itoa(peerevent.EV_VPCONNECT), pmi.peerMgrEv.Subscribe(peerevent.S_VPConnect{}))
	pmi.peerMgrSub.Set(strconv.Itoa(peerevent.EV_NVPCONNECT), pmi.peerMgrEv.Subscribe(peerevent.S_NVPConnect{}))
	pmi.peerMgrSub.Set(strconv.Itoa(peerevent.EV_VPDELETE), pmi.peerMgrEv.Subscribe(peerevent.S_DELETE_VP{}))
	pmi.peerMgrSub.Set(strconv.Itoa(peerevent.EV_NVPDELETE), pmi.peerMgrEv.Subscribe(peerevent.S_DELETE_NVP{}))
	pmi.peerMgrSub.Set(strconv.Itoa(peerevent.EV_NVPEXIT), pmi.peerMgrEv.Subscribe(peerevent.S_NVP_EXIT{}))
	pmi.peerMgrSub.Set(strconv.Itoa(peerevent.EV_UPDATE_SESSION_KEY), pmi.peerMgrEv.Subscribe(peerevent.S_UPDATE_SESSION_KEY{}))
	pmi.peerMgrSub.Set(strconv.Itoa(peerevent.EV_REBIND), pmi.peerMgrEv.Subscribe(peerevent.S_ReBind{}))

	return nil
}

func (pmgr *peerManagerImpl) unSubscribe() {
	pmgr.logger.Critical("Unsubscribe event...")
	for t := range pmgr.peerMgrSub.IterBuffered() {
		t.Val.(event.Subscription).Unsubscribe()
		pmgr.logger.Critical("Unsubscribe event... for", t.Key)
		pmgr.peerMgrSub.Remove(t.Key)
	}
}

// initialize the peerManager which is for init the local node
func (pmgr *peerManagerImpl) Start() error {
	if e := pmgr.configure(); e != nil {
		return e
	}
	if e := pmgr.prepare(); e != nil {
		return e
	}
	if pmgr.isRec {
		pmgr.node.info.SetRec()
	}
	pmgr.peerMgrEvClose = make(chan interface{})
	pmgr.listening()
	go pmgr.linking()
	go pmgr.updating()
	for pmgr.pts.HasNext() {
		pt := pmgr.pts.Pop()
		//Here should control the permission
		pmgr.bind(PEERTYPE_VP, pt.namespace, pt.id, pt.hostname, "")
	}
	pmgr.SetOnline()
	//new attend Process
	pmgr.logger.Notice("waiting...")
	<-pmgr.pendingChan
	if pmgr.isnew && pmgr.isVP {
		pmgr.logger.Critical("NEW PEER CONNECT")
		//this should wait until all nodes reverse connect to self.
		pmgr.broadcast(pb.MsgType_ATTEND, []byte(pmgr.GetLocalAddressPayload()))
		pmgr.eventHub.Post(event.AlreadyInChainEvent{})
	} else if !pmgr.isVP {
		if pmgr.isRec {
			pmgr.broadcast(pb.MsgType_NVPATTEND, []byte("True"))
		} else {
			pmgr.broadcast(pb.MsgType_NVPATTEND, []byte("False"))
		}
	}
	pmgr.logger.Infof("SELF hash: %s", pmgr.node.info.Hash)
	// after all connection
	return nil
}

func (pmgr *peerManagerImpl) listening() {
	pmgr.logger.Notice("start listening...")
	//Listening should listening all connection request, and handle it
	for subitem := range pmgr.peerMgrSub.IterBuffered() {
		go func(closechan chan interface{}, t string, s event.Subscription) {
			for {
				select {
				case <-closechan:
					{
						pmgr.logger.Debug("Listening sub goroutine stopped.")
						return
					}
				case ev := <-s.Chan():
					{
						// distribute all event to handlers
						if ev != nil {
							pmgr.distribute(t, ev.Data)
						}
					}
				}
			}
		}(pmgr.peerMgrEvClose, subitem.Key, subitem.Val.(event.Subscription))
	}
}

//distribute the event and payload
func (pmgr *peerManagerImpl) distribute(t string, ev interface{}) {
	switch ev.(type) {
	case peerevent.S_VPConnect:
		{
			conev := ev.(peerevent.S_VPConnect)
			if conev.ID > pmgr.nodeNum {
				pmgr.logger.Warning("Invalid peer connect: ", conev.ID, conev.Namespace, conev.Hostname)
				return
			}
			pmgr.logger.Infof("Got VP connected %s EVENT", conev.Hostname)
			// here how to connect to hypernet layer, generally, hypernet should already connect to
			// remote node by Inneraddr, here just need to bind the hostname.
			pmgr.bind(PEERTYPE_VP, conev.Namespace, conev.ID, conev.Hostname, "")

		}
	case peerevent.S_NVPConnect:
		{
			conev := ev.(peerevent.S_NVPConnect)
			pmgr.logger.Info("GOT NVP CONNECT EVENT", conev.Hostname)
			pmgr.bind(PEERTYPE_NVP, conev.Namespace, 0, conev.Hostname, conev.Hash)
		}
	case peerevent.S_DELETE_NVP:
		{
			conev := ev.(peerevent.S_DELETE_NVP)
			pmgr.logger.Info("Got a EV_DELETE_NVP for %s", conev.Hash)
			peer := pmgr.peerPool.GetNVPByHash(conev.Hash)
			if peer == nil {
				pmgr.logger.Warningf("This NVP(%s) not connect to this VP ignored.", conev.Hash)
				return
			}
			pmgr.logger.Critical("SEND TO NVP=>>", pmgr.node.info.Hash)
			m := pb.NewMsg(pb.MsgType_NVPDELETE, common.Hex2Bytes(pmgr.node.info.Hash))
			_, err := peer.Chat(m)
			if err != nil {
				pmgr.logger.Errorf("cannot delete NVP peer, reason: %s ", err.Error())
			} else {
				pmgr.peerPool.DeleteNVPPeer(conev.Hash)
				pmgr.logger.Noticef("delete NVP peer, hash %s, vp pool size(%d) nvp pool size(%d)", conev.Hash, pmgr.peerPool.GetVPNum(), pmgr.peerPool.GetNVPNum())
			}
		}
	case peerevent.S_DELETE_VP:
		{
			if pmgr.isVP {
				pmgr.logger.Warning("As A VP Node, this process cannot be invoked")
				return
			}
			conev := ev.(peerevent.S_DELETE_VP)
			pmgr.logger.Noticef("GOT a EV_DELETE_VP %s", conev.Hash)
			err := pmgr.peerPool.DeleteVPPeerByHash(conev.Hash)
			if err != nil {
				pmgr.logger.Errorf("cannot delete vp peer, reason: %s", err.Error())
			}
			if err == nil && !pmgr.isVP {
				if pmgr.peerPool.GetVPNum() == 0 {
					pmgr.logger.Warning("ALL Validate Peer are disconnect with this Non-Validate Peer")
					pmgr.logger.Warning("This Peer Will quit automaticlly after 3 seconds")
					<-time.After(3 * time.Second)
					pmgr.delchan <- true
				}
			}

		}
	case peerevent.S_UPDATE_SESSION_KEY:
		{
			conev := ev.(peerevent.S_UPDATE_SESSION_KEY)
			pmgr.logger.Infof("Got update SESSION key EVNET for %s", conev.NodeHash)
			if num, ok := pmgr.pendingSkMap.Get(conev.NodeHash); !ok {
				pmgr.pendingSkMap.Set(conev.NodeHash, 0)
			} else {
				pmgr.pendingSkMap.Set(conev.NodeHash, num.(int)+1)
			}
		}
	case peerevent.S_NVP_EXIT:
		{
			conev := ev.(peerevent.S_NVP_EXIT)
			pmgr.logger.Infof("Got NVP EXIT EVENT for %s", conev.NVPHash)
			er := pmgr.peerPool.DeleteNVPPeer(conev.NVPHash)
			if er != nil {
				pmgr.logger.Errorf("delete NVP failed, reason: %s", er.Error())
			}
		}
	default:
		pmgr.logger.Warningf("cannot determin the event type %v", reflect.TypeOf(ev))

	}
}

func (pmgr *peerManagerImpl) updating() {
	pmgr.logger.Notice("start updating process..")
	for {
		select {
		case <-pmgr.peerMgrEvClose:
			return
		case <-time.Tick(3 * time.Second):
			{
				if !pmgr.peerPool.Ready() {
					pmgr.logger.Warning("failed to update the session key, the peers pool has not prepared yet.")
					continue
				}
				waitList := make([]string, 0)
				for t := range pmgr.pendingSkMap.IterBuffered() {
					waitList = append(waitList, t.Key)
				}
				for _, hash := range waitList {
					pmgr.logger.Info("Update the session key for", hash)
					peer := pmgr.peerPool.GetPeerByHash(hash)
					if peer == nil {
						pmgr.logger.Warningf("Cannot find a VP peer By hash %s", hash)
						peer = pmgr.peerPool.GetNVPByHash(hash)
						if peer == nil {
							pmgr.logger.Errorf("Cannot find a NVP/VP peer By hash %s", hash)
							continue
						}
						pmgr.logger.Noticef("find a NVP peer by hash %s", hash)
					}
					err := peer.clientHello(false, false)
					if err != nil {
						pmgr.logger.Warningf("failed to update the session key, hash: %s, reason: %s", hash, err.Error())
						continue
					}
					pmgr.pendingSkMap.Remove(hash)
				}
			}
		}
	}
}

func (pmgr *peerManagerImpl) linking() {
	pmgr.logger.Notice("start linking process..")
	var flag = false
	for {
		select {
		case <-pmgr.peerMgrEvClose:
			return
		case <-time.Tick(3 * time.Second):
			{
				if pmgr.peerPool.Ready() && !flag && (pmgr.peerPool.GetVPNum() > int(math.Floor(float64(pmgr.n)/3.00))) {
					pmgr.pendingChan <- struct{}{}
					flag = true
				}
				waitList := make([]peerevent.S_ReBind, 0)
				for t := range pmgr.peerPool.pendingMap.IterBuffered() {
					conev := t.Val.(peerevent.S_ReBind)
					waitList = append(waitList, conev)
				}

				for _, wait := range waitList {
					pmgr.logger.Debugf("linking to peer [%s](%s) waiting queue size %d", wait.Namespace, wait.Hostname, pmgr.peerPool.pendingMap.Count())
					chts, err := pmgr.hts.GetAClientHTS()
					if err != nil {
						pmgr.logger.Errorf("get chts failed %s", err.Error())
						if len(waitList) <= 1 {
							break
						}
						continue
					}
					//the exist peer
					if wait.PeerType == PEERTYPE_VP {
						if p, ok := pmgr.peerPool.GetPeersByHostname(wait.Hostname); ok {
							err := p.clientHello(false, false)
							if err != nil {
								pmgr.logger.Warning("connect to vp %s failed, retry after 3 second", wait.Hostname)
								if len(waitList) <= 1 {
									break
								}
								continue
							}
						}
					} else {
						if p, ok := pmgr.peerPool.GetNVPByHostname(wait.Hostname); ok {
							pmgr.logger.Critical("Say hello to hostname:", wait.Hostname)
							err := p.clientHello(false, false)
							if err != nil {
								pmgr.logger.Warning("connect to nvp %s failed, retry after 3 second", wait.Hostname)
								if len(waitList) <= 1 {
									break
								}
								continue
							}
						}
					}
					//TODO here has a problem: generally, this method will be invoked before than
					//TODO the network persist the new hostname, it will return a error: the host hasn't been initialized.
					//TODO so the node may cannot connect to new peer, bind will be failed.
					//TODO to quick fix this: before create a new peer, sleep 1 second.
					<-time.After(time.Second)
					newPeer, err := NewPeer(wait.Namespace, wait.Hostname, wait.Id, pmgr.node.info, pmgr.hyperNet, chts, pmgr.eventHub)
					if err != nil {
						pmgr.logger.Warningf("cannot establish connection to %s, reason: %s ", wait.Hostname, err.Error())
						if len(waitList) <= 1 {
							break
						}
						continue
					}
					if wait.PeerType == PEERTYPE_VP {
						err = pmgr.peerPool.AddVPPeer(wait.Id, newPeer)
					} else {
						err = pmgr.peerPool.AddNVPPeer(wait.Hash, newPeer)
					}
					if err != nil {
						pmgr.logger.Warningf("add peer %s failed, retry after 3 second, err: %s", wait.Hostname, err.Error())
						if len(waitList) <= 1 {
							break
						}
						continue
					}
					pmgr.logger.Noticef("successful connect to [%s] (%s)", wait.Namespace, wait.Hostname)
					pmgr.peerPool.pendingMap.Remove(wait.Namespace + ":" + wait.Hostname)
				}
			}
		}
	}
}

//bind the namespace+id -> hostname
func (pmgr *peerManagerImpl) bind(peerType int, namespace string, id int, hostname string, hash string) {
	pmgr.logger.Debugf("bind to peer [%s](%s)", namespace, hostname)
	_, ok1 := pmgr.peerPool.GetPeersByHostname(hostname)
	_, ok2 := pmgr.peerPool.GetNVPByHostname(hostname)
	_, ok3 := pmgr.peerPool.pendingMap.Get(namespace + ":" + hostname)
	if !ok1 && !ok2 && !ok3 {
		pmgr.peerPool.pendingMap.Set(namespace+":"+hostname,
			peerevent.S_ReBind{
				PeerType:  peerType,
				Namespace: namespace,
				Id:        id,
				Hostname:  hostname,
				Hash:      hash,
			})
	}

	if peerType == PEERTYPE_VP {
		if vp, exist := pmgr.peerPool.GetPeersByHostname(hostname); exist {
			pmgr.logger.Debugf("rebind to peer [%s](%s) hash: %s", namespace, hostname, vp.info.Hash)
			pmgr.peerMgrEv.Post(peerevent.S_UPDATE_SESSION_KEY{vp.info.Hash})
		}

	} else if peerType == PEERTYPE_NVP {
		if nvp, exist := pmgr.peerPool.GetNVPByHostname(hostname); exist {
			pmgr.logger.Debugf("rebind to peer [%s](%s) hash: %s", namespace, hostname, nvp.info.Hash)
			pmgr.peerMgrEv.Post(peerevent.S_UPDATE_SESSION_KEY{nvp.info.Hash})
		}
	}

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
	if !pmgr.isOnline() || !pmgr.IsVP() {
		return
	}
	//TODO here can be improved, such as pre-calculate the peers' hash
	//TODO utils.GetHash will new a hasher every time, this waste of time.
	peerList := pmgr.peerPool.GetPeers()
	size := len(peerList)
	for _, id := range peers {
		//REVIEW here should ensure `>=` to avoid index out of range
		if int(id-1) >= size {
			continue
		}
		// this judge is for judge the max id is out of range or not
		if int(id) > pmgr.peerPool.MaxID() || id <= 0 {
			return
		}
		// REVIEW here should compare with local node id, shouldn't sub one
		if id == uint64(pmgr.node.info.GetID()) {
			pmgr.logger.Debugf("skip message send to self", pmgr.node.info.GetID())
			continue
		}
		// REVIEW  peers pool low layer struct is priority queue,
		// REVIEW this can ensure the node id order.
		// REVIEW avoid out of range
		// REVIEW here may cause a bug:
		// if a node hasn't connect such as node 3 connect before than node2
		// the peer list will  be [1,,3,4]
		// if send message to 2, the message will be send to node 3 actually
		//peer := peerList[int(id)-1]
		for _, peer := range peerList {
			// here do not need to judge peer is self, because self node has been skipped.
			if peer.info.Id != int(id) {
				continue
			}
			m := pb.NewMsg(msgType, payload)
			go func(p *Peer) {
				pmgr.logger.Debug("send message to", p.info.Id, p.info.Hostname)
				_, err := p.Chat(m)
				if err != nil {
					if ok, _ := regexp.MatchString("cannot get session Key", err.Error()); ok {
						go pmgr.peerMgrEv.Post(peerevent.S_UPDATE_SESSION_KEY{NodeHash: peer.info.Hash})
					}
					pmgr.logger.Warningf("Send Messahge failed, to (%s), reason: %s", p.hostname, err.Error())
				}
			}(peer)
		}
	}
}

func (pmgr *peerManagerImpl) Broadcast(payload []byte) {
	pmgr.broadcast(pb.MsgType_SESSION, payload)
}

//Broadcast message to all binding hosts
func (pmgr *peerManagerImpl) broadcast(msgType pb.MsgType, payload []byte) {
	if !pmgr.isOnline() {
		pmgr.logger.Warningf("Broadcast failed, local node has not prepared (isonline: %v, isvp: %v)", pmgr.isonline, pmgr.isVP)
		return
	}
	// use IterBuffered for better performance
	peerList := pmgr.peerPool.GetPeers()
	for _, p := range peerList {
		pmgr.logger.Debug("Broadcast to VP peer", p.hostname, msgType)
		go func(peer *Peer) {
			if peer.hostname == peer.local.Hostname {
				return
			}
			// this is un thread safe, because this,is a pointer
			m := pb.NewMsg(msgType, payload)
			_, err := peer.Chat(m)
			if err != nil {
				pmgr.logger.Warningf("hostname [target: %s](local: %s) chat err: send self %s ", peer.hostname, peer.local.Hostname, err.Error())
				if ok, _ := regexp.MatchString(".+decrypt.+?", err.Error()); ok {
					pmgr.logger.Warningf("update the session key because of cannot decrypt msg [%s](%s)", peer.namespace, peer.hostname)
					pmgr.peerMgrEv.Post(peerevent.S_UPDATE_SESSION_KEY{peer.info.Hash})
				}
			}
		}(p)

	}
}

// set peer managers primary peer and node
func (pmgr *peerManagerImpl) SetPrimary(_id uint64) error {
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
func (pmgr *peerManagerImpl) GetLocalNodeHash() string {
	return pmgr.node.info.Hash
}

//GetVPPeers return all vp peers
func (pmgr *peerManagerImpl) GetVPPeers() []*Peer {
	return pmgr.peerPool.GetPeers()
}

func (pmgr *peerManagerImpl) Stop() {
	select {
	case <-pmgr.peerMgrEvClose:
		pmgr.logger.Infof("the peer manager channel already closed.")
	default:
		close(pmgr.peerMgrEvClose)

	}
	pmgr.broadcast(pb.MsgType_NVPEXIT, []byte(pmgr.GetLocalNodeHash()))
	pmgr.logger.Criticalf("Unbind all slots...")
	pmgr.SetOffline()
	pmgr.unSubscribe()
	pmgr.peercnf.vip.Set("self.rec", true)
	err := pmgr.peercnf.vip.WriteConfig()
	pmgr.isRec = true
	pmgr.isnew = false
	if err != nil {
		pmgr.logger.Errorf("cannot stop peermanager, reason %s", err.Error())
	}
	pmgr.node.UnBindAll()
}

// AddNode
// update routing table when new peer's join request is accepted
func (pmgr *peerManagerImpl) UpdateRoutingTable(payLoad []byte) {
	//unmarshal info
	i := info.InfoUnmarshal(payLoad)
	pmgr.bind(PEERTYPE_VP, i.Namespace, i.Id, i.Hostname, "")
	<-time.After(4 * time.Second)
	pmgr.nodeNum++
	for _, p := range pmgr.peerPool.GetPeers() {
		pmgr.logger.Info("update table", p.hostname)
	}
	err := pmgr.peerPool.PersistList()
	if err != nil {
		pmgr.logger.Errorf("cannot persisit peer config reason: %s", err.Error())
		return
	}
}

func (pmgr *peerManagerImpl) GetLocalAddressPayload() []byte {
	return pmgr.node.info.Serialize()
}

func (pmgr *peerManagerImpl) SetOnline() {
	pmgr.isonline.UnLock()
}

func (pmgr *peerManagerImpl) isOnline() bool {
	return !pmgr.isonline.IsLocked()
}

func (pmgr *peerManagerImpl) SetOffline() {
	pmgr.isonline.TryLock()
}

//GetRouterHashifDelete returns after delete specific peer, the router table hash , self new id and the delete id
func (pmgr *peerManagerImpl) GetRouterHashifDelete(hash string) (afterDelRouterHash string, selfNewId uint64, delID uint64) {
	afterDelRouterHash, selfNewId, delID, err := pmgr.peerPool.TryDelete(pmgr.GetLocalNodeHash(), hash)
	if err != nil {
		pmgr.logger.Errorf("cannot try del peer, error: %s", err.Error())
	}
	return
}

//DeleteNode delete the specific hash node, if the node hash is self, this node will stoped.
func (pmgr *peerManagerImpl) DeleteNode(hash string) error {
	pmgr.logger.Critical("DELETE NODE", hash)
	if pmgr.node.info.Hash == hash {
		pmgr.Stop()
		pmgr.logger.Critical(" WARNING!! THIS NODE HAS BEEN DELETED!")
		pmgr.logger.Critical(" THIS NODE WILL STOP IN 3 SECONDS")
		<-time.After(3 * time.Second)
		pmgr.logger.Critical("EXIT..")
		pmgr.delchan <- true

	}
	err := pmgr.peerPool.DeleteVPPeerByHash(hash)
	if err != nil {
		return err
	}
	return pmgr.peerPool.PersistList()
}

func (pmgr *peerManagerImpl) DeleteNVPNode(hash string) error {
	pmgr.logger.Critical("Delete None Validate peer hash:", hash)
	if pmgr.node.info.Hash == hash {
		pmgr.logger.Critical("Please do not send delete NVP command to nvp")
		return nil
	}
	ev := peerevent.S_DELETE_NVP{
		Hash: hash,
	}
	go pmgr.peerMgrEv.Post(ev)
	return nil
}

// InfoGetter get the peer info to manager
// get the all peer list to broadcast
func (pmgr *peerManagerImpl) GetAllPeers() []*Peer {
	return pmgr.peerPool.GetPeers()
}

// get local node instance
//GetLocalNode() *Node
// Get local node id
func (pmgr *peerManagerImpl) GetNodeId() int {
	return pmgr.node.info.GetID()
}

func (pmgr *peerManagerImpl) GetN() int {
	return pmgr.n
}

//get the peer information of all nodes.
func (pmgr *peerManagerImpl) GetPeerInfo() PeerInfos {

	var peerInfos PeerInfos
	peers := pmgr.GetAllPeers()
	sHostName := pmgr.node.info.Hostname

	for _, p := range peers {

		dHostName := p.info.GetHostName()
		ip, port := pmgr.hyperNet.GetDNS(dHostName)
		peerInfo := PeerInfo{
			ID:        p.info.GetID(),
			Namespace: p.info.GetNameSpace(),
			Hash:      p.info.GetHash(),
			Hostname:  dHostName,
			IsPrimary: p.info.GetPrimary(),
			IsVP:      p.info.GetVP(),
			IP:        ip,
			Port:      port,
		}

		start := time.Now().UnixNano()
		resp, err := p.net.Discuss(dHostName, pb.NewPkg([]byte("ping"), pb.ControlType_KeepAlive))
		if err != nil {
			peerInfo.Status = STOP
		} else if resp.Type == pb.ControlType_Response {
			peerInfo.Status = ALIVE
		}

		if dHostName != sHostName {
			peerInfo.Delay = time.Now().UnixNano() - start
		}

		peerInfos = append(peerInfos, peerInfo)
	}

	return peerInfos
}

// use by new peer when join the chain dynamically only
func (pmgr *peerManagerImpl) GetRouters() []byte {
	b, e := pmgr.peerPool.Serlize()
	if e != nil {
		pmgr.logger.Errorf("cannot serialize the peerpool,err:%s ", e.Error())
		return nil
	}
	return b
}

// random select a VP and send msg to it
func (pmgr *peerManagerImpl) SendRandomVP(payload []byte) error {
	peers := pmgr.peerPool.GetPeers()
	randomStack := random_stack.NewStack()
	for _, peer := range peers {
		randomStack.Push(peer)
	}
	var err error
	for !randomStack.Empty() {
		speer := randomStack.RandomPop().(*Peer)
		m := pb.NewMsg(pb.MsgType_SESSION, payload)
		_, err = speer.Chat(m)
		if err == nil {
			break
		}
	}
	return err
}

// broadcast information to NVP peers
func (pmgr *peerManagerImpl) BroadcastNVP(payLoad []byte) error {
	return pmgr.broadcastNVP(pb.MsgType_SESSION, payLoad)
}
func (pmgr *peerManagerImpl) broadcastNVP(msgType pb.MsgType, payload []byte) error {
	if !pmgr.isOnline() || !pmgr.IsVP() {
		return nil
	}
	// use IterBuffered for better performance
	for t := range pmgr.peerPool.nvpPool.Iter() {
		pmgr.logger.Debugf("(NVP) send message to %s", t.Key)
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
				pmgr.logger.Warningf("hostname [target: %s](local: %s) chat err: send self %s ", peer.hostname, peer.local.Hostname, err.Error())
				if ok, _ := regexp.MatchString("cannot find the filed msg slot", err.Error()); ok {
					pmgr.peerPool.DeleteNVPPeer(t.Key)
					pmgr.logger.Noticef("delete NVP [%s](%s)", peer.namespace, peer.hostname)
				}
			}
		}(p)
	}
	return nil
}

// send a message to specific NVP peer (by nvp hash) UNICAST
func (pmgr *peerManagerImpl) SendMsgNVP(payLoad []byte, nvpList []string) error {
	return pmgr.sendMsgNVP(pb.MsgType_SESSION, payLoad, nvpList)
}

func (pmgr *peerManagerImpl) sendMsgNVP(msgType pb.MsgType, payLoad []byte, nvpList []string) error {
	if !pmgr.isOnline() || !pmgr.IsVP() {
		return nil
	}
	// use IterBuffered for better performance
	for t := range pmgr.peerPool.nvpPool.Iter() {
		for _, nvphash := range nvpList {
			if nvphash != t.Key {
				continue
			}
			pmgr.logger.Debugf("(VP) send message to %s", t.Key)
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
func (pmgr *peerManagerImpl) IsVP() bool {
	return pmgr.isVP
}
