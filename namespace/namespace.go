//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package namespace

import (
	"sync"
	"time"

	"github.com/hyperchain/hyperchain/admittance"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/consensus"
	"github.com/hyperchain/hyperchain/consensus/csmgr"
	"github.com/hyperchain/hyperchain/manager"
	"github.com/hyperchain/hyperchain/manager/event"
	"github.com/hyperchain/hyperchain/manager/protos"
	"github.com/hyperchain/hyperchain/namespace/rpc"
	"github.com/hyperchain/hyperchain/p2p"

	"github.com/hyperchain/hyperchain/common/service/server"
	"github.com/hyperchain/hyperchain/core/fiber"
	"github.com/hyperchain/hyperchain/core/fiber/executor"
	"github.com/hyperchain/hyperchain/core/oplog"
	"github.com/hyperchain/hyperchain/core/oplog/kvlog"
	"github.com/op/go-logging"
)

// This file defines the Namespace interface, which manages all the
// operations related to a certain namespace.
// There are 5 components in one namespace:
// 1. CAManager: Authority management is used to authenticate identity
//    in network.
// 2. PeerManager: Network p2p management is used to establish connection
// 	  then deliver connection and consensus messages.
// 3. Consenter: Consensus component is used to order the coming requests
//    and guarantee the consistency of all consensus nodes.
// 4. EventHub: The component is used to help internal components to
//    interact with each other.
// 5. JsonRpcProcess: Requests sent from clients are dispatched by
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

	// ProcessRequest process request under this namespace.
	ProcessRequest(request interface{}) interface{}

	// Name returns the name of current namespace.
	Name() string

	// GetCAManager returns the CAManager of current namespace.
	GetCAManager() *admittance.CAManager
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
	rpc       rpc.RequestProcessor
	opLog     oplog.OpLog

	fiber   fiber.Fiber
	is      *server.InternalServer
	name    string
	status  *Status
	conf    *common.Config
	restart bool
	delFlag chan bool
}

// newNamespaceImpl returns a newed Namespace instance with
// the given name and config
func newNamespaceImpl(namespace string, conf *common.Config, delFlag chan bool, is *server.InternalServer) (*namespaceImpl, error) {
	conf.Set(common.NAMESPACE, namespace)
	if err := common.InitHyperLogger(namespace, conf); err != nil {
		return nil, err
	}

	status := &Status{
		state: newed,
		desc:  "newed",
		lock:  new(sync.RWMutex),
	}

	ns := &namespaceImpl{
		name:      namespace,
		status:    status,
		conf:      conf,
		eventMux:  new(event.TypeMux),
		filterMux: new(event.TypeMux),
		restart:   false,
		delFlag:   delFlag,
		is:        is,
	}
	ns.logger = common.GetLogger(namespace, "namespace")
	return ns, nil
}

// init initializes the namespace by init all the components
// one by one.
func (ns *namespaceImpl) init() error {
	ns.logger.Noticef("Init namespace %s", ns.Name())

	// 1. init CaManager to manage account identity.
	cm, err := admittance.NewCAManager(ns.conf)
	if err != nil {
		ns.logger.Errorf("Init CA manager failed: %s", err)
		return err
	}
	ns.caMgr = cm

	// 2. init peerManager to start grpc server and client.
	peerConfPath := common.GetPath(ns.Name(), ns.conf.GetString(common.PEER_CONFIG_PATH))
	if !common.FileExist(peerConfPath) {
		ns.logger.Errorf("Cannot find the peer config for namespace: %s", ns.Name())
		return ErrNonExistConfig
	}
	peerMgr, err := p2p.GetPeerManager(ns.Name(), peerConfPath, ns.eventMux, ns.delFlag)
	if err != nil {
		ns.logger.Errorf("Get peer manager failed: %s", err)
		return err
	}
	ns.peerMgr = peerMgr

	// init opLog
	ns.opLog = kvlog.New(ns.conf)

	nss := ns.is.ServerRegistry().Namespace(ns.name)
	if ns.fiber, err = executor.NewFiber(ns.conf, nss, ns.opLog); err != nil {
		return err
	}

	// 3. init consensus module to order requests.
	consenter, err := csmgr.Consenter(ns.Name(), ns.conf, ns.opLog, ns.eventMux, ns.filterMux, peerMgr.GetN())
	if err != nil {
		ns.logger.Errorf("Init Consenter for namespace %s error: %s", ns.Name(), err)
		return err
	}
	ns.consenter = consenter

	// 4. init Eventhub to coordinate message delivery between local modules.
	eh := manager.New(ns.Name(), ns.eventMux, ns.filterMux, ns.peerMgr, consenter, cm)
	ns.eh = eh

	// 5. init JsonRpcProcessor to process incoming requests.
	ns.rpc = rpc.NewJsonRpcProcessorImpl(ns.Name(), ns.GetApis(ns.Name()))

	ns.status.setState(initialized)
	return nil
}

// GetNamespace returns the Namespace instance of the given name.
func GetNamespace(name string, conf *common.Config, delFlag chan bool, is *server.InternalServer) (Namespace, error) {
	ns, err := newNamespaceImpl(name, conf, delFlag, is)
	if err != nil {
		ns.logger.Errorf("New namespace %s failed: %s", name, err)
		return ns, err
	}
	err = ns.init()
	return ns, err
}

// Start starts all services under this namespace.
func (ns *namespaceImpl) Start() error {
	ns.logger.Noticef("Try to start namespace %s", ns.Name())

	var err error

	state := ns.status.getState()
	if state < initialized {
		err := ns.init()
		if err != nil {
			return err
		}
	}

	if ns.status.getState() == running {
		ns.logger.Warningf("Namespace %s is already running", ns.Name())
		return nil
	}

	// 1. start consenter
	err = ns.consenter.Start()
	if err != nil {
		return err
	}

	// 2. start event hub
	ns.eh.Start()

	// 3. start peer manager
	err = ns.peerMgr.Start()
	if err != nil {
		return err
	}

	// 4. consensus the routers
	ns.passRouters()

	// 5. start negotiateView
	if ns.peerMgr.IsVP() {
		ns.negotiateView()
	}

	// 6. start rpc processor
	if err = ns.rpc.Start(); err != nil {
		return err
	}
	ns.status.setState(running)
	ns.logger.Noticef("Namespace %s start successfully", ns.Name())
	ns.restart = true

	go ns.fiber.Start()
	return nil
}

// Stop stops all services under this namespace.
func (ns *namespaceImpl) Stop() error {
	ns.logger.Noticef("Try to stop namespace %s", ns.Name())
	state := ns.status.getState()
	if state != running {
		ns.logger.Warningf("Namespace %s is not running now, need not to stop", ns.Name())
		return nil
	}
	// 1. stop request processor.
	err := ns.rpc.Stop()
	if err != nil {
		ns.logger.Error(err)
	}

	// 2. stop eventhub.
	ns.eh.Stop()

	// 3. stop consensus service.
	ns.consenter.Stop()

	// 4. stop peerManager.
	ns.peerMgr.Stop()

	ns.status.setState(closed)
	ns.logger.Noticef("Namespace %s stopped!", ns.Name())
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

// Name returns the name of this namespace.
func (ns *namespaceImpl) Name() string {
	return ns.name
}

// GetCAManager returns the CAManager of this namespace.
func (ns namespaceImpl) GetCAManager() *admittance.CAManager {
	return ns.caMgr
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
				ns.logger.Errorf("Not supported event %v", r)
			}
		}
	}
	ns.logger.Errorf("Process request error: namespace %s is not running now!", ns.Name())
	return nil
}

// negotiateView sends negotiate view event to consensus module.
func (ns *namespaceImpl) negotiateView() {
	negoView := &protos.Message{
		Type:      protos.Message_NEGOTIATE_VIEW,
		Timestamp: time.Now().UnixNano(),
		Payload:   nil,
		Id:        0,
	}
	ns.consenter.RecvLocal(negoView)
}

// passRouters passes network router to consensus module.
func (ns *namespaceImpl) passRouters() {
	router := ns.peerMgr.GetRouters()
	msg := protos.RoutersMessage{Routers: router}
	ns.consenter.RecvLocal(msg)
}
