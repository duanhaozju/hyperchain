//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package namespace

import (
	"errors"
	"hyperchain/accounts"
	"hyperchain/admittance"
	"hyperchain/common"
	"hyperchain/consensus"
	"hyperchain/consensus/csmgr"
	"hyperchain/core/db_utils"
	"hyperchain/core/executor"
	"hyperchain/event"
	"hyperchain/manager"
	"hyperchain/p2p"
	"hyperchain/namespace/rpc"
	"hyperchain/api"
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
type NamespaceImpl struct {
	nsInfo       *NamespaceInfo
	status       *Status

	Conf         *common.Config
	eventMux     *event.TypeMux
	consenter    consensus.Consenter
	CaMgr        *admittance.CAManager
	am           *accounts.AccountManager
	eh           *manager.EventHub
	grpcMgr      *p2p.GRPCPeerManager
	executor     *executor.Executor

	rpcProcesser rpc.RPCProcesser
}

type API struct {
	Srvname string        // srvname under which the rpc methods of Service are exposed
	Version   string      // api version for DApp's
	Service   interface{} // receiver instance which holds the methods
	Public    bool        // indication if the methods must be considered safe for public use
}

func newNamespaceImpl(name string, conf *common.Config) (*NamespaceImpl, error) {

	ninfo := &NamespaceInfo{
		name: name,
	}
	status := &Status{
		state: STARTTING,
		desc:  "startting",
	}
	ns := &NamespaceImpl{
		nsInfo:   ninfo,
		status:   status,
		Conf:     conf,
		eventMux: new(event.TypeMux),
	}
	return ns, nil
}

func (ns *NamespaceImpl) init() error {
	logger.Criticalf("Init namespace %s", ns.Name())

	//1.init DB
	err := db_utils.InitDBForNamespace(ns.Conf, ns.Name())
	if err != nil {
		logger.Errorf("init db for namespace %s error, %v", ns.Name(), err)
		return err
	}

	//2.init CaManager
	cm, cmerr := admittance.GetCaManager(ns.Conf)
	if cmerr != nil {
		panic("cannot initliazied the camanager")
	}
	ns.CaMgr = cm

	//3. init peer manager to start grpc server and client
	grpcPeerMgr := p2p.NewGrpcManager(ns.Conf)
	ns.grpcMgr = grpcPeerMgr

	//4.init pbft consensus
	consenter := csmgr.Consenter(ns.Name(), ns.Conf, ns.eventMux)
	consenter.Start()
	ns.consenter = consenter

	//5.init account manager
	am := accounts.NewAccountManager(ns.Conf)
	am.UnlockAllAccount(ns.Conf.GetString(common.KEY_STORE_DIR))
	ns.am = am

	//6.init block pool to save block
	executor := executor.NewExecutor(ns.Name(), ns.Conf, ns.eventMux)
	if executor == nil {
		return errors.New("Initialize BlockPool failed")
	}

	executor.CreateInitBlock(ns.Conf)
	executor.Initialize()

	//7. init peer manager
	eh := manager.New(ns.Name(), ns.eventMux, executor, ns.grpcMgr, consenter, am, cm)
	ns.eh = eh

	// 8. init JsonRpcProcessor

	ns.rpcProcesser = rpc.NewRPCProcessorImpl(ns.Name(), ns.GetApis())
	ns.rpcProcesser.Start()
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
func (ns *NamespaceImpl) Start() error {
	logger.Criticalf("namespace %s startting", ns.Name())
	//TODO: start this namespace service
	//TODO: provide start method for every components
	ns.status.state = STARTTED
	return nil
}

//Stop stop services under this namespace.
func (ns *NamespaceImpl) Stop() error {
	//TODO: stop a namespace service
	//TODO: to provide Stop method for every components
	ns.status.state = CLOSED
	return nil
}

//Status return current namespace status.
func (ns *NamespaceImpl) Status() *Status {
	return ns.status
}

//Info return basic information of this namespace.
func (ns *NamespaceImpl) Info() *NamespaceInfo {
	return ns.nsInfo
}

func (ns *NamespaceImpl) Name() string {
	return ns.nsInfo.name
}

//ProcessRequest process request under this namespace
func (ns *NamespaceImpl) ProcessRequest(request interface{}) interface{} {
	switch r := request.(type) {
	case *common.RPCRequest:
		return ns.handleJsonRequest(r)
	default:
		logger.Errorf("event not supportted %v", r)
	}
	return nil
}

func (ns *NamespaceImpl) GetApis() []hpc.API {

	return []hpc.API{
		{
			Srvname: "tx",
			Version:   "0.4",
			Service:   hpc.NewPublicTransactionAPI("global", ns.eventMux, ns.eh, ns.Conf),
			Public:    true,
		},
		{
			Srvname: "node",
			Version:   "0.4",
			Service:   hpc.NewPublicNodeAPI(ns.eh),
			Public:    true,
		},
		{
			Srvname: "block",
			Version:   "0.4",
			Service:   hpc.NewPublicBlockAPI("global"),
			Public:    true,
		},
		{
			Srvname: "account",
			Version:   "0.4",
			Service:   hpc.NewPublicAccountAPI("global", ns.eh, ns.Conf),
			Public:    true,
		},
		{
			Srvname: "contract",
			Version:   "0.4",
			Service:   hpc.NewPublicContractAPI("global", ns.eventMux, ns.eh, ns.Conf),
			Public:    true,
		},
		{
			Srvname: "cert",
			Version:   "0.4",
			Service:   hpc.NewPublicCertAPI(ns.CaMgr),
			Public:    true,
		},
	}

}