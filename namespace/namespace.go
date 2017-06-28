//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package namespace

import (
	"errors"
	"github.com/op/go-logging"
	"hyperchain/accounts"
	"hyperchain/admittance"
	"hyperchain/common"
	"hyperchain/consensus"
	"hyperchain/consensus/csmgr"
	//"hyperchain/core/db_utils"
	"hyperchain/core/db_utils"
	"hyperchain/core/executor"
	"hyperchain/hyperdb"
	"hyperchain/manager"
	"hyperchain/manager/event"
	"hyperchain/manager/protos"
	"hyperchain/namespace/rpc"
	"hyperchain/p2p"
	"sync"
	"time"
	"fmt"
	"github.com/terasum/lettor/utils"
)

//Namespace represent the namespace instance
type Namespace interface {
	//Start start services under this namespace.
	Start() error
	//Stop stop services under this namespace.
	Stop() error
	//Restart restart services under this namespace.
	Restart() error
	//Status return current namespace status.
	Status() *Status
	//Info return basic information of this namespace.
	Info() *NamespaceInfo
	//ProcessRequest process request under this namespace.
	ProcessRequest(request interface{}) interface{}
	//Name of current namespace.
	Name() string
	//GetCAManager get CAManager by namespace name.
	GetCAManager() *admittance.CAManager
	//GetExecutor fetch executor module
	GetExecutor() *executor.Executor
}
type NsState int

const (
	initnew NsState = 1 << iota
	initialized
	running
	closed
)

//Status dynamic state of current namespace.
type Status struct {
	lock  *sync.RWMutex
	state NsState
	desc  string
}

func (s *Status) setState(state NsState) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.state = state
}

func (s *Status) getState() NsState {
	s.lock.RLock()
	defer s.lock.RUnlock()
	state := s.state
	return state
}

//namespaceImpl implementation of Namespace
type namespaceImpl struct {
	logger    *logging.Logger

	nsInfo    *NamespaceInfo
	status    *Status

	conf      *common.Config
	eventMux  *event.TypeMux
	consenter consensus.Consenter
	caMgr     *admittance.CAManager
	am        *accounts.AccountManager
	eh        *manager.EventHub
	peerMgr   p2p.PeerManager
	executor  *executor.Executor

	rpc       rpc.RequestProcessor
	restart   bool
}

type API struct {
	Srvname string      // srvname under which the rpc methods of Service are exposed
	Version string      // api version for DApp's
	Service interface{} // receiver instance which holds the methods
	Public  bool        // indication if the methods must be considered safe for public use
}

func newNamespaceImpl(name string, conf *common.Config) (*namespaceImpl, error) {
	// Init Hyperlogger
	if _, err := common.InitHyperLogger(conf); err != nil {
		return nil, err
	}

	status := &Status{
		state: initnew,
		desc:  "startting",
		lock:  new(sync.RWMutex),
	}

	nsInfo, err := NewNamespaceInfo(conf.GetString(common.PEER_CONFIG_PATH), name, common.GetLogger(name, "namespace"))
	nsInfo.PrintInfo()
	if err != nil {
		return nil, err
	}
	ns := &namespaceImpl{
		nsInfo:   nsInfo,
		status:   status,
		conf:     conf,
		eventMux: new(event.TypeMux),
		restart:  false,
	}
	ns.logger = common.GetLogger(name, "namespace")
	return ns, nil
}

func (ns *namespaceImpl) init() error {
	ns.logger.Criticalf("Init namespace %s", ns.Name())

	//1.init DB
	err := db_utils.InitDBForNamespace(ns.conf, ns.Name())
	if err != nil {
		ns.logger.Errorf("init db for namespace: %s error, %v", ns.Name(), err)
		return err
	}

	//2.init CaManager
	cm, err := admittance.NewCAManager(ns.conf)
	if err != nil {
		logger.Error(err)
		panic("cannot initliazied the camanager")
	}
	ns.caMgr = cm


	peerconf := ns.conf.GetString(common.PEER_CONFIG13_PATH)
	if !utils.Exist(peerconf) {
		panic("cannot find the peer config")
	}
	fmt.Println("peerconf path is",peerconf)
	//peerconf := "/home/chenquan/Workspace/go/src/hyperchain/p2p/test/peerconfig.yaml"

	//3. init peer manager to start grpc server and client
	logger.Warning("getPeerManager for",ns.Name())
	peerMgr, err := p2p.GetPeerManager(ns.Name(),peerconf,ns.eventMux)
	if err != nil {
		ns.logger.Error(err)
		return err
	}
	// here should be GetP2PManager(ns.conf,ns.eventMux,cm)
	ns.peerMgr = peerMgr

	//4.init pbft consensus
	consenter, err := csmgr.Consenter(ns.Name(), ns.conf, ns.eventMux)
	if err != nil {
		logger.Errorf("init Consenter for namespace %s error, %v", ns.Name(), err)
		return err
	}
	ns.consenter = consenter

	//5.init account manager
	am := accounts.NewAccountManager(ns.conf)
	am.UnlockAllAccount(ns.conf.GetString(common.KEY_STORE_DIR))
	ns.am = am

	//6.init block pool to save block
	executor := executor.NewExecutor(ns.Name(), ns.conf, ns.eventMux)
	if executor == nil {
		return errors.New("Initialize Executor failed")
	}

	executor.CreateInitBlock(ns.conf)
	ns.executor = executor

	//7. init eventhub
	eh := manager.New(ns.Name(), ns.eventMux, executor, ns.peerMgr, consenter, am, cm)
	ns.eh = eh
	ns.status.setState(initialized)

	// 8. init JsonRpcProcessor
	ns.rpc = rpc.NewJsonRpcProcessorImpl(ns.Name(), ns.GetApis(ns.Name()))
	return nil
}

func GetNamespace(name string, conf *common.Config) (Namespace, error) {
	ns, err := newNamespaceImpl(name, conf)
	if err != nil {
		ns.logger.Errorf("namespace %s init error", name)
		return ns, err
	}
	err = ns.init()
	return ns, err
}

//Start start services under this namespace.
func (ns *namespaceImpl) Start() error {
	ns.logger.Noticef("try to start namespace: %s", ns.Name())
	state := ns.status.getState()
	if state < initialized {
		err := ns.init()
		if err != nil {
			return err
		}
	}

	if ns.status.getState() == running {
		logger.Criticalf("namespace: %s is already running", ns.Name())
		return nil
	}
	//0.init db
	if ns.restart {
		err := hyperdb.StartDatabase(ns.conf, ns.Name())
		if err != nil {
			ns.logger.Error(err)
			return err
		}
		ns.logger.Noticef("start db for namespace: %s successful", ns.Name())
	}

	//1.start consenter
	ns.consenter.Start()
	//2.start executor
	ns.executor.Start()
	//3.start event hub
	ns.eh.Start()
	//4.start grpc manager
	ns.peerMgr.Start()
	//5 consensue the routers
	ns.passRouters()
	//6. negotiateView
	ns.negotiateView()

	ns.rpc.Start()
	ns.status.setState(running)
	ns.logger.Noticef("namespace: %s start successful", ns.Name())
	ns.restart = true
	return nil
}

func (ns *namespaceImpl) negotiateView() {
	ns.logger.Debug("negotiate view")
	negoView := &protos.Message{
		Type:      protos.Message_NEGOTIATE_VIEW,
		Timestamp: time.Now().UnixNano(),
		Payload:   nil,
		Id:        0,
	}
	ns.consenter.RecvLocal(negoView)
}

func (ns *namespaceImpl) passRouters() {
	router := ns.peerMgr.GetRouters()
	msg := protos.RoutersMessage{Routers: router}
	ns.consenter.RecvLocal(msg)
}

//Stop stop services under this namespace.
func (ns *namespaceImpl) Stop() error {
	ns.logger.Noticef("try to stop namespace: %s", ns.Name())
	state := ns.status.getState()
	if state != running {
		ns.logger.Criticalf("namespace: %s not running now, need not to stop", ns.Name())
	}
	//1.stop request processor
	ns.rpc.Stop()

	//2.stop eventhub
	ns.eh.Stop()

	//3.stop executor
	ns.executor.Stop()

	//4.stop consensus service
	ns.consenter.Close()

	//5.stop peer manager
	go ns.peerMgr.Stop()

	ns.status.setState(closed)
	//ns.logger.Notice()
	//close related database
	hyperdb.StopDatabase(ns.Name())

	ns.logger.Noticef("namespace: %s stopped!", ns.Name())
	return nil
}

//Restart restart services under this namespace.
func (ns *namespaceImpl) Restart() error {
	err := ns.Stop()
	if err != nil {
		return err
	}
	return ns.Start()
}

//Status return current namespace status.
func (ns *namespaceImpl) Status() *Status {
	return ns.status
}

//Info return basic information of this namespace.
func (ns *namespaceImpl) Info() *NamespaceInfo {
	return ns.nsInfo
}

func (ns *namespaceImpl) Name() string {
	return ns.nsInfo.name
}

//GetCAManager get CAManager by namespace name.
func (ns namespaceImpl) GetCAManager() *admittance.CAManager {
	return ns.caMgr
}
func (ns namespaceImpl) GetExecutor() *executor.Executor {
	return ns.executor
}

//ProcessRequest process request under this namespace
func (ns *namespaceImpl) ProcessRequest(request interface{}) interface{} {
	if ns.status.getState() == running {
		if request != nil {
			switch r := request.(type) {
			case *common.RPCRequest:
				return ns.handleJsonRequest(r)
			default:
				ns.logger.Errorf("event not supportted %v", r)
			}
		}
	}
	logger.Errorf("Process request error, namespace %s is not running now!", ns.Name())
	return nil
}
