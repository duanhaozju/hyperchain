package controller

import (
	"github.com/hyperchain/hyperchain/admittance"
	"github.com/hyperchain/hyperchain/api"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/common/service/client"
	pb "github.com/hyperchain/hyperchain/common/service/protos"
	"github.com/hyperchain/hyperchain/hyperdb"
	"github.com/hyperchain/hyperchain/manager/event"
	"github.com/hyperchain/hyperchain/manager/filter"
	"github.com/hyperchain/hyperchain/namespace/rpc"

	"github.com/hyperchain/hyperchain/core/ledger/chain"
	"github.com/hyperchain/hyperchain/service/executor/controller/executor"
	"github.com/hyperchain/hyperchain/service/executor/handler"
	"github.com/op/go-logging"
	"sync"
)

type executorService interface {
	//Start start executor.
	Start() error

	//Stop stop executor.
	Stop() error

	//Namespace this executor belong to.
	Name() string

	// ProcessRequest process request under this namespace.
	ProcessRequest(request interface{}) interface{}

	//GetCAManager fetch CA manager.
	GetCAManager() *admittance.CAManager
}

type executorServiceImpl struct {
	namespace string
	executor  *executor.Executor

	client *client.ServiceClient // maintains connection to the order

	conf   *common.Config
	logger *logging.Logger
	status *Status

	rpc       rpc.RequestProcessor //TODO: for what
	caManager *admittance.CAManager

	eventMux  *event.TypeMux
	filterMux *event.TypeMux

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
	es.logger.Infof("init executor client for namespace %s", es.namespace)

	// 1. init DB for current executor client.
	err := chain.InitDBForNamespace(es.conf, es.namespace)
	if err != nil {
		es.logger.Errorf("Init db for namespace: %s error, %v", es.namespace, err)
		return err
	}

	// 2. initial client client
	//TODO(Xiaoyi Wang): fix fixed host
	es.client, err = client.New(es.conf.GetInt(common.INTERNAL_PORT), "127.0.0.1", client.EXECUTOR, es.namespace)
	if err != nil {
		es.logger.Errorf("Init client client for namespace %s error, %v", es.namespace, err)
		return err
	}

	// 3. filter system
	es.filterSystem = filter.NewEventSystem(es.filterMux)

	// 4. initial executor
	executor, err := executor.NewExecutor(es.namespace, es.conf, es.eventMux, es.filterMux)
	if err != nil {
		es.logger.Errorf("Init executor client for namespace %s error, %v", es.namespace, err)
		return err
	}
	executor.CreateInitBlock(es.conf)
	es.executor = executor

	h := handler.New(es.namespace, executor)
	es.client.AddHandler(h)

	es.caManager, err = admittance.NewCAManager(es.conf)
	if err != nil {
		es.logger.Errorf("Init executor client for camanager %s error, %v", es.namespace, err)
		return err
	}

	// 5. add jsonrpc processor
	es.rpc = rpc.NewJsonRpcProcessorImpl(es.namespace, es.GetExecutorApis(es.namespace))

	// 6. initialized status
	es.status.setState(initialized)
	return nil
}

func (es *executorServiceImpl) Start() error {
	es.logger.Noticef("Start executor client for namespace: %s", es.namespace)

	state := es.status.getState()
	if state < initialized {
		err := es.init()
		if err != nil {
			es.logger.Errorf("Executor client for namespace %s initialization failed %v", es.namespace, err)
			return err
		}
	}

	if es.status.getState() == running {
		es.logger.Errorf("Executor client for namespace %s is already running", es.namespace)
		return nil
	}

	//TODO(Xiaoyi Wang): add restart flag judge.
	//err := chain.InitDBForNamespace(es.conf, es.namespace)
	//if err != nil {
	//	es.logger.Errorf("Init db for namespace: %s error, %v", es.namespace, err)
	//	return err
	//}

	es.logger.Noticef("start db for namespace: %s successful", es.namespace)

	// 2. start executor
	err := es.executor.Start()
	if err != nil {
		es.logger.Errorf("StartExecutorServiceByName executor for namespace %s error, %v", es.namespace, err)
		return err
	}

	es.status.setState(running)

	// 3. establish connection
	err = es.client.Connect()
	if err != nil {
		es.logger.Errorf("Establish connection for namespace %s error, %v", es.namespace, err)
		return err
	}

	_, err = es.client.Register(0, pb.FROM_EXECUTOR, &pb.RegisterMessage{ // TODO: Fix id
		Namespace: es.namespace,
	})
	if err != nil {
		es.logger.Errorf("Executor client register failed for namespace %s error, %v", es.namespace, err)
		return err
	}

	es.status.setState(running)

	//// 8. start rpc processor
	if err = es.rpc.Start(); err != nil {
		return err
	}

	return nil

}

func (es *executorServiceImpl) Stop() error {
	es.logger.Noticef("try to stop namespace: %s", es.namespace)
	state := es.status.getState()
	if state != running {
		es.logger.Criticalf("Executor client for namespace: %s not running now, need not to stop", es.namespace)
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
	es.logger.Noticef("Executor client for namespace: %s stopped!", es.namespace)
	return nil
}

func (es *executorServiceImpl) ProcessRequest(request interface{}) interface{} {
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

func (es *executorServiceImpl) GetCAManager() *admittance.CAManager {
	return es.caManager
}

func (es *executorServiceImpl) Name() string {
	return es.namespace
}

// GetExecutorApis returns the RPC api of specified namespace.
func (es *executorServiceImpl) GetExecutorApis(namespace string) map[string]*api.API {
	return map[string]*api.API{
		"execTx": {
			Svcname: "tx",
			Version: "1.5",
			Service: api.NewPublicTransactionExecAPI(namespace, es.conf),
			Public:  true,
		},
		"execBlock": {
			Svcname: "block",
			Version: "1.5",
			Service: api.NewPublicBlockAPI(namespace),
			Public:  true,
		},
		"execAccount": {
			Svcname: "account",
			Version: "1.5",
			Service: api.NewPublicAccountAPI(namespace, nil, es.conf),
			Public:  true,
		},
		"execContract": {
			Svcname: "contract",
			Version: "1.5",
			Service: api.NewPublicContractExecAPI(namespace, es.conf),
			Public:  true,
		},
		"execCert": {
			Svcname: "cert",
			Version: "1.5",
			Service: api.NewCertAPI(namespace, es.caManager),
			Public:  true,
		},
		"execSub": {
			Svcname: "sub",
			Version: "1.5",
			Service: api.NewFilterAPI(namespace, es.filterSystem, es.conf),
		},
		"execArchive": {
			Svcname: "archive",
			Version: "1.5",
			Service: api.NewPublicArchiveAPI(namespace, nil, es.conf),
		},
	}
}
