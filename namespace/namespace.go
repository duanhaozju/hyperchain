//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package namespace

import (
	"github.com/hyperchain/hyperchain/admittance"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/consensus"
	"github.com/hyperchain/hyperchain/consensus/csmgr"
	"github.com/hyperchain/hyperchain/core/executor"
	"github.com/hyperchain/hyperchain/core/ledger/chain"
	"github.com/hyperchain/hyperchain/hyperdb"
	"github.com/hyperchain/hyperchain/manager"
	"github.com/hyperchain/hyperchain/manager/event"
	"github.com/hyperchain/hyperchain/manager/protos"
	"github.com/hyperchain/hyperchain/namespace/rpc"
	"github.com/hyperchain/hyperchain/p2p"
	"github.com/op/go-logging"
	"sync"
	"time"
)

// This file defines the Namespace interface, which manages all the
// operations related to a certain namespace.
// There are 7 components in one namespace:
// 1. DB: Database is used to store data about this namespace.
// 2. CAManager: Authority management is used to authenticate identity
//    in network.
// 3. PeerManager: Network p2p management is used to establish connection
// 	  then deliver connection and consensus messages.
// 4. Consenter: Consensus component is used to order the coming requests
//    and guarantee the consistency of all consensus nodes.
// 5. Executor: Executor is mainly used to validate and commit blocks.
// 6. EventHub: The component is used to help internal components to
//    interact with each other.
// 7. JsonRpcProcess: Requests sent from clients are dispatched by
//    NamespaceManager to corresponding namespace processor first, then
//    JsonRpcProcess will actually process the request.

// Data and messages in different namespaces are isolated from each
// other, so don't worry if operations on a namespace will influence the
// others or not.

// Namespace manages a certain namespace instance.
type Namespace interface {
	// Start initializes and starts all services under this namespace.
	Start() error

	// Stop stops all services under this namespace.
	Stop() error

	// Restart restarts services under this namespace.
	Restart() error

	// Status returns the current namespace status, which may be:
	// 1. newed: after newed this namespace instance, before initialize
	// 2. initialized: after initialized
	// 3. running: after Start
	// 4. closed: after Close
	Status() *Status

	// Info returns the basic information of current namespace.
	Info() *NamespaceInfo

	// ProcessRequest process request under this namespace.
	ProcessRequest(request interface{}) interface{}

	// Name returns the name of current namespace.
	Name() string

	// GetCAManager returns the CAManager of current namespace.
	GetCAManager() *admittance.CAManager

	// GetExecutor returns the executor module of current namespace.
	GetExecutor() *executor.Executor
}

type NsState int

const (
	newed NsState = 1 << iota
	initialized
	running
	closed
)

// Status describes the dynamic state of current namespace.
type Status struct {
	lock  *sync.RWMutex
	state NsState
	desc  string
}

// setState sets the current namespace status, and update the description.
func (s *Status) setState(state NsState) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.state = state
	s.setDescription()
}

// getState returns the current namespace status.
func (s *Status) getState() NsState {
	s.lock.RLock()
	defer s.lock.RUnlock()
	state := s.state
	return state
}

// setDescription updates the current description by current state.
func (s *Status) setDescription() {
	switch s.state {
	case newed:
		s.desc = "newed"
	case initialized:
		s.desc = "initialized"
	case running:
		s.desc = "running"
	case closed:
		s.desc = "closed"
	default:
		s.desc = "Unknown state"
	}
}

// namespaceImpl implements the Namespace interface.
type namespaceImpl struct {
	logger *logging.Logger

	// registry-subscription service
	eventMux  *event.TypeMux
	filterMux *event.TypeMux

	consenter consensus.Consenter
	caMgr     *admittance.CAManager
	eh        *manager.EventHub
	peerMgr   p2p.PeerManager
	executor  *executor.Executor
	rpc       rpc.RequestProcessor

	nsInfo  *NamespaceInfo
	status  *Status
	conf    *common.Config
	restart bool
	delFlag chan bool
}

// newNamespaceImpl returns a newed Namespace instance with
// the given name and config
func newNamespaceImpl(namespace string, conf *common.Config, delFlag chan bool) (*namespaceImpl, error) {
	conf.Set(common.NAMESPACE, namespace)
	if err := common.InitHyperLogger(namespace, conf); err != nil {
		return nil, err
	}

	status := &Status{
		state: newed,
		desc:  "newed",
		lock:  new(sync.RWMutex),
	}
	ppath := common.GetPath(namespace, conf.GetString(common.PEER_CONFIG_PATH))
	nsInfo, err := NewNamespaceInfo(ppath, namespace, common.GetLogger(namespace, "namespace"))

	if err != nil {
		return nil, err
	}
	ns := &namespaceImpl{
		nsInfo:    nsInfo,
		status:    status,
		conf:      conf,
		eventMux:  new(event.TypeMux),
		filterMux: new(event.TypeMux),
		restart:   false,
		delFlag:   delFlag,
	}
	ns.logger = common.GetLogger(namespace, "namespace")
	return ns, nil
}

// init initializes the namespace by init all the components
// one by one.
func (ns *namespaceImpl) init() error {
	ns.logger.Criticalf("Init namespace %s", ns.Name())

	// 1. init DB for current namespace.
	err := chain.InitDBForNamespace(ns.conf, ns.Name())
	if err != nil {
		ns.logger.Errorf("Init db for namespace: %s error, %v", ns.Name(), err)
		return err
	}

	// 2. init CaManager to manage account identity.
	cm, err := admittance.NewCAManager(ns.conf)
	if err != nil {
		ns.logger.Error(err)
		panic("Cannot initialize the CAManager!")
	}
	ns.caMgr = cm

	peerconf := common.GetPath(ns.Name(), ns.conf.GetString(common.PEER_CONFIG_PATH))
	if !common.FileExist(peerconf) {
		panic("Cannot find the peer config!")
	}

	// 3. init peerManager to start grpc server and client.
	ns.logger.Warning("GetPeerManager for", ns.Name())
	peerMgr, err := p2p.GetPeerManager(ns.Name(), peerconf, ns.eventMux, ns.delFlag)
	if err != nil {
		ns.logger.Error(err)
		return err
	}
	ns.peerMgr = peerMgr

	// 4. init consensus module to order requests.
	consenter, err := csmgr.Consenter(ns.Name(), ns.conf, ns.eventMux, ns.filterMux, peerMgr.GetN())
	if err != nil {
		ns.logger.Errorf("init Consenter for namespace %s error, %v", ns.Name(), err)
		return err
	}
	ns.consenter = consenter

	// 5. init Executor to validate and commit block.
	executor, err := executor.NewExecutor(ns.Name(), ns.conf, ns.eventMux, ns.filterMux, peerMgr.GetLocalNodeHash())
	if err != nil {
		ns.logger.Errorf("init Executor for namespace %s error, %v", ns.Name(), err)
		return err
	}

	executor.CreateInitBlock(ns.conf)
	ns.executor = executor

	// 6. init Eventhub to coordinate message delivery between local modules.
	eh := manager.New(ns.Name(), ns.eventMux, ns.filterMux, executor, ns.peerMgr, consenter, cm)
	ns.eh = eh

	// 7. init JsonRpcProcessor to process incoming requests.
	ns.rpc = rpc.NewJsonRpcProcessorImpl(ns.Name(), ns.GetApis(ns.Name()))

	ns.status.setState(initialized)
	return nil
}

// GetNamespace returns the Namespace instance of the given name.
func GetNamespace(name string, conf *common.Config, delFlag chan bool) (Namespace, error) {
	ns, err := newNamespaceImpl(name, conf, delFlag)
	if err != nil {
		ns.logger.Errorf("namespace %s init error", name)
		return ns, err
	}
	err = ns.init()
	return ns, err
}

// Start starts all services under this namespace.
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
		ns.logger.Criticalf("namespace: %s is already running", ns.Name())
		return nil
	}
	// 1. start db service
	if ns.restart {
		err := hyperdb.StartDatabase(ns.conf, ns.Name())
		if err != nil {
			ns.logger.Error(err)
			return err
		}
		ns.logger.Noticef("start db for namespace: %s successful", ns.Name())
	}

	// 2. start consenter
	ns.consenter.Start()

	// 3. start executor
	err := ns.executor.Start()
	if err != nil {
		return err
	}

	// 4. start event hub
	ns.eh.Start()

	// 5. start grpc manager
	err = ns.peerMgr.Start()
	if err != nil {
		return err
	}

	// 6. consensus the routers
	ns.passRouters()

	// 7. start negotiateView
	if ns.peerMgr.IsVP() {
		ns.negotiateView()
	}

	// 8. start rpc processor
	if err = ns.rpc.Start(); err != nil {
		return err
	}
	ns.status.setState(running)
	ns.logger.Noticef("namespace: %s start successful", ns.Name())
	ns.restart = true
	return nil
}

// negotiateView sends negotiate view event to consensus module.
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

// Stop stops all services under this namespace.
func (ns *namespaceImpl) Stop() error {
	ns.logger.Noticef("try to stop namespace: %s", ns.Name())
	state := ns.status.getState()
	if state != running {
		ns.logger.Criticalf("namespace: %s not running now, need not to stop", ns.Name())
		return nil
	}
	// 1. stop request processor.
	err := ns.rpc.Stop()
	if err != nil {
		ns.logger.Error(err)
	}

	// 2. stop eventhub.
	ns.eh.Stop()

	// 3. stop executor.
	err = ns.executor.Stop()
	if err != nil {
		ns.logger.Error(err)
	}

	// 4. stop consensus service.
	ns.consenter.Stop()

	// 5. stop peerManager.
	ns.peerMgr.Stop()

	ns.status.setState(closed)

	// 6. close related database.
	err = hyperdb.StopDatabase(ns.Name())
	if err != nil {
		ns.logger.Error(err)
	}

	ns.logger.Noticef("namespace: %s stopped!", ns.Name())
	return nil
}

// Restart restarts all services under this namespace.
func (ns *namespaceImpl) Restart() error {
	err := ns.Stop()
	if err != nil {
		return err
	}
	return ns.Start()
}

// Status returns the current namespace status.
func (ns *namespaceImpl) Status() *Status {
	return ns.status
}

// Info returns basic information of this namespace.
func (ns *namespaceImpl) Info() *NamespaceInfo {
	return ns.nsInfo
}

// Name returns the name of this namespace.
func (ns *namespaceImpl) Name() string {
	return ns.nsInfo.name
}

// GetCAManager returns the CAManager of this namespace.
func (ns namespaceImpl) GetCAManager() *admittance.CAManager {
	return ns.caMgr
}

// GetExecutor returns the executor of this namespace.
func (ns namespaceImpl) GetExecutor() *executor.Executor {
	return ns.executor
}

// ProcessRequest processes request under this namespace, and dispatch request
// to corresponding handler(now support json request only).
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
	ns.logger.Errorf("Process request error, namespace %s is not running now!", ns.Name())
	return nil
}
