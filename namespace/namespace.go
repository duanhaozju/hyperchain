//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package namespace

import (
	"errors"
	"github.com/op/go-logging"
	"hyperchain/accounts"
	"hyperchain/admittance"
	"hyperchain/api"
	"hyperchain/common"
	"hyperchain/consensus"
	"hyperchain/consensus/csmgr"
	"hyperchain/core/db_utils"
	"hyperchain/core/executor"
	"hyperchain/event"
	"hyperchain/manager"
	"hyperchain/namespace/rpc"
	"hyperchain/p2p"
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
	state NsState
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
	logger *logging.Logger

	nsInfo *NamespaceInfo
	status *Status

	Conf      *common.Config
	eventMux  *event.TypeMux
	consenter consensus.Consenter
	CaMgr     *admittance.CAManager
	am        *accounts.AccountManager
	eh        *manager.EventHub
	grpcMgr   *p2p.GRPCPeerManager
	executor  *executor.Executor

	rpcProcesser rpc.RequestProcesser
}

type API struct {
	Srvname string      // srvname under which the rpc methods of Service are exposed
	Version string      // api version for DApp's
	Service interface{} // receiver instance which holds the methods
	Public  bool        // indication if the methods must be considered safe for public use
}

func newNamespaceImpl(name string, conf *common.Config) (*namespaceImpl, error) {

	ninfo := &NamespaceInfo{
		name: name,
	}
	status := &Status{
		state: initnew,
		desc:  "startting",
	}
	ns := &namespaceImpl{
		nsInfo:   ninfo,
		status:   status,
		Conf:     conf,
		eventMux: new(event.TypeMux),
	}
	ns.logger = common.GetLogger(name, "namespace")
	return ns, nil
}

func (ns *namespaceImpl) init() error {
	ns.logger.Criticalf("Init namespace %s", ns.Name())

	//1.init DB
	err := db_utils.InitDBForNamespace(ns.Conf, ns.Name())
	if err != nil {
		ns.logger.Errorf("init db for namespace: %s error, %v", ns.Name(), err)
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
	consenter, err := csmgr.Consenter(ns.Name(), ns.Conf, ns.eventMux)
	if err != nil {
		logger.Errorf("init Consenter for namespace %s error, %v", ns.Name(), err)
		return err
	}
	consenter.Start()
	ns.consenter = consenter

	//5.init account manager
	am := accounts.NewAccountManager(ns.Conf)
	am.UnlockAllAccount(ns.Conf.GetString(common.KEY_STORE_DIR))
	ns.am = am

	//6.init block pool to save block
	executor := executor.NewExecutor(ns.Name(), ns.Conf, ns.eventMux)
	if executor == nil {
		return errors.New("Initialize Executor failed")
	}

	executor.CreateInitBlock(ns.Conf)
	executor.Initialize()

	//7. init peer manager
	eh := manager.New(ns.Name(), ns.eventMux, executor, ns.grpcMgr, consenter, am, cm)
	ns.eh = eh
	ns.status.state = initialized

	// 8. init JsonRpcProcessor

	ns.rpcProcesser = rpc.NewRPCProcessorImpl(ns.Name(), ns.GetApis())
	ns.rpcProcesser.Start()
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
	state := ns.status.state
	if state < initialized {
		err := ns.init()
		if err != nil {
			return err
		}
	}

	if state == running {
		logger.Criticalf("namespace: %s is already running", ns.Name())
		return nil
	}

	//TODO: add start component logic here
	ns.status.state = running
	ns.logger.Noticef("namespace: %s start successful", ns.Name())
	return nil
}

//Stop stop services under this namespace.
func (ns *namespaceImpl) Stop() error {
	ns.logger.Noticef("try to stop namespace: %s", ns.Name())
	state := ns.status.state

	if state != running {
		ns.logger.Criticalf("namespace: %s not running now, need not to stop", ns.Name())
	}
	//TODO: to provide Stop method for every components

	ns.status.state = closed

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
	return ns.CaMgr
}

//ProcessRequest process request under this namespace
func (ns *namespaceImpl) ProcessRequest(request interface{}) interface{} {
	switch r := request.(type) {
	case *common.RPCRequest:
		return ns.handleJsonRequest(r)
	default:
		ns.logger.Errorf("event not supportted %v", r)
	}
	return nil
}

func (ns *namespaceImpl) GetApis() []hpc.API {
	return []hpc.API{
		{
			Srvname: "tx",
			Version: "0.4",
			Service: hpc.NewPublicTransactionAPI("global", ns.eventMux, ns.eh, ns.Conf),
			Public:  true,
		},
		{
			Srvname: "node",
			Version: "0.4",
			Service: hpc.NewPublicNodeAPI(ns.eh),
			Public:  true,
		},
		{
			Srvname: "block",
			Version: "0.4",
			Service: hpc.NewPublicBlockAPI("global"),
			Public:  true,
		},
		{
			Srvname: "account",
			Version: "0.4",
			Service: hpc.NewPublicAccountAPI("global", ns.eh, ns.Conf),
			Public:  true,
		},
		{
			Srvname: "contract",
			Version: "0.4",
			Service: hpc.NewPublicContractAPI("global", ns.eventMux, ns.eh, ns.Conf),
			Public:  true,
		},
		{
			Srvname: "cert",
			Version: "0.4",
			Service: hpc.NewPublicCertAPI(ns.CaMgr),
			Public:  true,
		},
	}

}
