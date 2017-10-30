package controller

import (
	"github.com/op/go-logging"
	"hyperchain/admittance"
	hapi "hyperchain/api"
	"hyperchain/common"
	pb "hyperchain/common/protos"
	"hyperchain/common/service/client"
	"hyperchain/core/executor"
	"hyperchain/core/ledger/chain"
	"hyperchain/hyperdb"
	"hyperchain/manager/filter"
	"hyperchain/namespace/rpc"
	"hyperchain/service/hypexec/handler"
	"sync"

	"hyperchain/manager/event"
	"strings"
)

type executorService interface {
	Start() error

	Stop() error

	// Name returns the name of current namespace.
	Name() string

	// ProcessRequest process request under this namespace.
	ProcessRequest(request interface{}) interface{}

	GetCAManager() *admittance.CAManager
}

type executorServiceImpl struct {
	// namespace
	namespace string
	// real executor object
	executor *executor.Executor
	// manager the connection with service
	service *client.ServiceClient
	// config
	conf *common.Config
	// logger
	logger *logging.Logger

	//executorApi *api.ExecutorApi

	status *Status

	rpc rpc.RequestProcessor

	caManager *admittance.CAManager

	eventMux *event.TypeMux

	filterMux *event.TypeMux

	// filter system for subscription
	filterSystem *filter.EventSystem
}

type EsState int

const (
	newed EsState = 1 << iota
	initialized
	running
	closed
)

// Status describes the dynamic state of current namespace.
type Status struct {
	lock  *sync.RWMutex
	state EsState
	desc  string
}

// setState sets the current namespace status, and update the description.
func (s *Status) setState(state EsState) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.state = state
	s.setDescription()
}

// getState returns the current namespace status.
func (s *Status) getState() EsState {
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

func NewExecutorService(ns string, conf *common.Config) *executorServiceImpl {

	// init hyper logger for executor service
	conf.Set(common.NAMESPACE, ns)
	if err := common.InitHyperLogger(ns, conf); err != nil {
		return nil
	}

	// new status
	status := &Status{
		state: newed,
		desc:  "newed",
		lock:  new(sync.RWMutex),
	}

	return &executorServiceImpl{
		namespace: ns,
		conf:      conf,
		logger:    common.GetLogger(ns, "executor_service"),
		status:    status,
		eventMux:  new(event.TypeMux),
		filterMux: new(event.TypeMux),
	}
}

func (es *executorServiceImpl) init() error {
	es.logger.Infof("init executor service for namespace %s", es.namespace)

	// 1. init DB for current executor service.
	err := chain.InitExecutorDBForNamespace(es.conf, es.namespace)
	if err != nil {
		es.logger.Errorf("Init db for namespace: %s error, %v", es.namespace, err)
		return err
	}

	// 2. initial service client
	service, err := client.New(es.conf.GetInt(common.INTERNAL_PORT), "127.0.0.1", client.EXECUTOR, es.namespace)
	if err != nil {
		es.logger.Errorf("Init service client for namespace %s error, %v", es.namespace, err)
		return err
	}
	es.service = service

	// 3. filter system
	es.filterSystem = filter.NewEventSystem(es.filterMux)

	// 4. initial executor
	executor, err := executor.NewExecutor(es.namespace, es.conf, es.eventMux, es.filterMux, es.service)
	if err != nil {
		es.logger.Errorf("Init executor service for namespace %s error, %v", es.namespace, err)
		return err
	}
	executor.CreateInitBlock(es.conf)
	es.executor = executor

	h := handler.New(es.namespace, executor)
	service.AddHandler(h)

	// 5. add jsonrpc processor
	es.rpc = rpc.NewJsonRpcProcessorImpl(es.namespace, es.GetApis(es.namespace), es.GetRemoteApis(es.namespace),
	strings.Split(es.conf.GetString(common.EXECUTOR_HOST_ADDR), ":")[0], es.conf.GetInt(common.JSON_RPC_PORT))

	// 6. initialized status
	es.status.setState(initialized)

	return nil
}

func (es *executorServiceImpl) Start() error {
	es.logger.Noticef("Start executor service for namespace: %s", es.namespace)

	state := es.status.getState()
	if state < initialized {
		err := es.init()
		if err != nil {
			es.logger.Errorf("Executor service for namespace %s initialization failed %v", es.namespace, err)
			return err
		}
	}

	if es.status.getState() == running {
		es.logger.Errorf("Executor service for namespace %s is already running", es.namespace)
		return nil
	}

	// 1. start executor and service client
	//err = hyperdb.StartDatabase(es.conf, es.namespace)
	//if err != nil {
	//    es.logger.Errorf("StartExecutorServiceByName database for namespace %s error, %v", es.namespace, err)
	//    return err
	//}
	//es.logger.Noticef("start db for namespace: %s successful", es.namespace)

	// 2. start executor
	err := es.executor.Start()
	if err != nil {
		es.logger.Errorf("StartExecutorServiceByName executor for namespace %s error, %v", es.namespace, err)
		return err
	}

	//append: to satisfy apiserver tests.
	es.status.setState(running)

	//es.executorApi = api.NewExecutorApi(es.executor, es.namespace)

	// 3. establish connection
	err = es.service.Connect()
	if err != nil {
		es.logger.Errorf("Establish connection for namespace %s error, %v", es.namespace, err)
		return err
	}

	_, err = es.service.Register(0, pb.FROM_EXECUTOR, &pb.RegisterMessage{ // TODO: Fix id
		Namespace: es.namespace,
	})
	if err != nil {
		es.logger.Errorf("Executor service register failed for namespace %s error, %v", es.namespace, err)
		return err
	}

	es.status.setState(running)

	// 8. start rpc processor
	if err = es.rpc.Start(); err != nil {
		return err
	}

	return nil

}

func (es *executorServiceImpl) Stop() error {
	es.logger.Noticef("try to stop namespace: %s", es.namespace)
	state := es.status.getState()
	if state != running {
		es.logger.Criticalf("Executor service for namespace: %s not running now, need not to stop", es.namespace)
		return nil
	}

	// 1. stop executor.
	err := es.executor.Stop()
	if err != nil {
		es.logger.Errorf("StopExecutorServiceByName executor for namespace %s error, %v", es.namespace, err)
		return err
	}

	// 2. close related database.
	err = hyperdb.StopDatabase(es.namespace)
	if err != nil {
		es.logger.Errorf("StopExecutorServiceByName database for namespace %s error, %v", es.namespace, err)
		return err
	}

	es.status.setState(closed)
	es.logger.Noticef("Executor service for namespace: %s stopped!", es.namespace)
	return nil
}

func (es *executorServiceImpl) ProcessRequest(request interface{}) interface{} {
	//TODO Check finish logic
	es.logger.Critical("request : %v", request)
	es.logger.Critical("executor status: %v", es.status.getState())
	if es.status.getState() == running {
		if request != nil {
			switch r := request.(type) {
			case *common.RPCRequest:
				return es.handleJsonRequest(r)
			default:
				es.logger.Errorf("event not supported %v", r)
			}
		}
	}
	es.logger.Errorf("Process request error, namespace %s is not running now!", es.namespace)
	return nil
}

func (es *executorServiceImpl) handleJsonRequest(request *common.RPCRequest) *common.RPCResponse {
	return es.rpc.ProcessRequest(request)
}

func (es *executorServiceImpl) GetApis(namespace string) map[string]*hapi.API {
	//TODO need to add more APIS
	return map[string]*hapi.API{
		"block": {
			Svcname: "block",
			Version: "1.5",
			Service: hapi.NewPublicBlockAPI(namespace),
			Public:  true,
		},
		"txdb": {
			Svcname: "txdb",
			Version: "1.5",
			Service: hapi.NewDBTransactionAPI(namespace, es.conf),
			Public:  true,
		},
		"accountdb": {
			Svcname: "accountdb",
			Version: "1.5",
			Service: hapi.NewPublicAccountExecutorAPI(namespace, es.conf),
			Public:  true,
		},
		"contractExe": {
			Svcname: "contractExe",
			Version: "1.5",
			Service: hapi.NewContarctExAPI(namespace, es.conf),
			Public:  true,
		},
		//"cert": {
		//	Svcname: "cert",
		//	Version: "1.5",
		//	Service: hapi.NewCertAPI(namespace, es.caManager),
		//	Public:  true,
		//},
		"sub": {
			Svcname: "sub",
			Version: "1.5",
			Service: hapi.NewFilterAPI(namespace, es.filterSystem, es.conf),
		},
		"archive": {
			Svcname: "archive",
			Version: "1.5",
			Service: hapi.NewPublicArchiveAPI(namespace, es.conf),
			Public:  true,
		},
		//TODO: implements cert API for module2
	}
}

func (es *executorServiceImpl) GetRemoteApis(namespace string) map[string]*hapi.API {
	return map[string]*hapi.API{
		"tx": {
			Svcname: "tx",
			Version: "1.5",
			Service: hapi.NewPublicTransactionAPI(namespace, nil, es.conf),
			Public:  true,
		},
		"node": {
			Svcname: "node",
			Version: "1.5",
			Service: hapi.NewPublicNodeAPI(namespace, nil),
			Public:  true,
		},
		"account": {
			Svcname: "account",
			Version: "1.5",
			Service: hapi.NewPublicAccountAPI(namespace, nil, es.conf),
			Public:  true,
		},
		"contract": {
			Svcname: "contract",
			Version: "1.5",
			Service: hapi.NewPublicContractAPI(namespace, nil, es.conf),
			Public:  true,
		},
		"cert": {
			Svcname: "cert",
			Version: "1.5",
			Service: hapi.NewCertAPI(namespace, es.caManager),
			Public:  true,
		},
		"sub": {
			Svcname: "sub",
			Version: "1.5",
			Service: hapi.NewFilterAPI(namespace, nil, es.conf),
		},
	}
}


func (es *executorServiceImpl) GetCAManager() *admittance.CAManager{
	return es.caManager
}

func (es *executorServiceImpl) Name() string {
	return es.namespace
}
