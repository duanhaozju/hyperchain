//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package namespace

import (
	"errors"
	"hyperchain/accounts"
	"hyperchain/admittance"
	"hyperchain/api/jsonrpc/core"
	"hyperchain/common"
	"hyperchain/consensus"
	"hyperchain/consensus/csmgr"
	"hyperchain/core/db_utils"
	"hyperchain/core/executor"
	"hyperchain/event"
	"hyperchain/manager"
	"hyperchain/p2p"
)

//Namespace represent the namespace instance
type Namespace interface {
	//Start start services under this namespace.
	Start() error
	//Stop stop services under this namespace.
	Stop() error
	//Status return current namespace status.
	Status() *Status
	//Info return basic information of this namespace.
	Info() *NamespaceInfo
	//ProcessRequest process request under this namespace.
	ProcessRequest(request interface{}) interface{}

	//Name of current namespace.
	Name() string
}

const (
	STARTTING = iota
	STARTTED
	RUNNING
	CLOSING
	CLOSED
)

//Status dynamic state of current namespace.
type Status struct {
	state int
	desc  string
}

//NamespaceInfo basic information of this namespace.
type NamespaceInfo struct {
	name    string
	members []string //member ips list
	desc    string
}

//namespaceImpl implementation of Namespace
type namespaceImpl struct {
	nsInfo *NamespaceInfo
	status *Status

	conf      *common.Config
	eventMux  *event.TypeMux
	consenter consensus.Consenter
	caMgr     *admittance.CAManager
	am        *accounts.AccountManager
	eh        *manager.EventHub
	grpcMgr   *p2p.GRPCPeerManager
	executor  *executor.Executor
}

func newNamespaceImpl(name string, conf *common.Config) (*namespaceImpl, error) {

	ninfo := &NamespaceInfo{
		name: name,
	}
	status := &Status{
		state: STARTTING,
		desc:  "startting",
	}
	ns := &namespaceImpl{
		nsInfo:   ninfo,
		status:   status,
		conf:     conf,
		eventMux: new(event.TypeMux),
	}
	return ns, nil
}

func (ns *namespaceImpl) init() error {
	logger.Criticalf("Init namespace %s", ns.Name())

	//1.init DB
	err := db_utils.InitDBForNamespace(ns.conf, ns.Name())
	if err != nil {
		logger.Errorf("init db for namespace %s error, %v", ns.Name(), err)
		return err
	}

	//2.init CaManager
	cm, cmerr := admittance.GetCaManager(ns.conf)
	if cmerr != nil {
		panic("cannot initliazied the camanager")
	}
	ns.caMgr = cm

	//3. init peer manager to start grpc server and client
	grpcPeerMgr := p2p.NewGrpcManager(ns.conf)
	ns.grpcMgr = grpcPeerMgr

	//4.init pbft consensus
	consenter, err := csmgr.Consenter(ns.Name(), ns.conf, ns.eventMux)
	if err != nil {
		logger.Errorf("init Consenter for namespace %s error, %v", ns.Name(), err)
		return err
	}
	consenter.Start()
	ns.consenter = consenter

	//5.init account manager
	am := accounts.NewAccountManager(ns.conf)
	am.UnlockAllAccount(ns.conf.GetString(common.KEY_STORE_DIR))
	ns.am = am

	//6.init block pool to save block
	executor := executor.NewExecutor(ns.Name(), ns.conf, ns.eventMux)
	if executor == nil {
		return errors.New("Initialize BlockPool failed")
	}

	executor.CreateInitBlock(ns.conf)
	executor.Initialize()

	//7. init peer manager
	eh := manager.New(ns.Name(), ns.eventMux, executor, ns.grpcMgr, consenter, am, cm)
	ns.eh = eh
	return nil
}

func GetNamespace(name string, conf *common.Config) (Namespace, error) {
	ns, err := newNamespaceImpl(name, conf)
	if err != nil {
		logger.Errorf("namespace %s init error", name)
		return ns, err
	}
	err = ns.init()
	return ns, err
}

//Start start services under this namespace.
func (ns *namespaceImpl) Start() error {
	logger.Criticalf("namespace %s startting", ns.Name())
	//TODO: start this namespace service
	//TODO: provide start method for every components
	ns.status.state = STARTTED
	return nil
}

//Stop stop services under this namespace.
func (ns *namespaceImpl) Stop() error {
	//TODO: stop a namespace service
	//TODO: to provide Stop method for every components
	ns.status.state = CLOSED
	return nil
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

//ProcessRequest process request under this namespace
func (ns *namespaceImpl) ProcessRequest(request interface{}) interface{} {
	switch r := request.(type) {
	case *jsonrpc.JSONRequest:
		return ns.handleJsonRequest(r)
	default:
		logger.Errorf("event not supportted %v", r)
	}
	return nil
}
