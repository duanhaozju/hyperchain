package manager

import (
	"github.com/op/go-logging"
	hapi "hyperchain/api"
	"hyperchain/common"
	pb "hyperchain/common/protos"
	"hyperchain/core/executor"
	"hyperchain/hyperdb"
	"hyperchain/namespace/rpc"
	"sync"
	"hyperchain/common/client"
	"hyperchain/admittance"
    "hyperchain/core/ledger/chain"
    "hyperchain/service/executor/handler"
)

type executorService interface {
	Start() error

	Stop() error

	// ProcessRequest process request under this namespace.
	ProcessRequest(request interface{}) interface{}

	// Name returns the name of current namespace.
	Name() string

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
	}
}

func (es *executorServiceImpl) init() error {
	es.logger.Criticalf("Init executor service for namespace %s", es.namespace)

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


	// 3. initial executor
	executor, err := executor.NewExecutor(es.namespace, es.conf, nil, nil, es.service)
	if err != nil {
		es.logger.Errorf("Init executor service for namespace %s error, %v", es.namespace, err)
		return err
	}
	executor.CreateInitBlock(es.conf)
	es.executor = executor

    h := handler.New(executor)
    service.AddHandler(h)

	// 4. add jsonrpc processor
	es.rpc = rpc.NewJsonRpcProcessorImpl(es.namespace, es.GetApis(es.namespace))
	es.rpc.Start()

	// 5. initialized status
	es.status.setState(initialized)

	return nil
}

func (es *executorServiceImpl) Start() error {
	es.logger.Noticef("Try to start executor service for namespace: %s", es.namespace)

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
	//    es.logger.Errorf("Start database for namespace %s error, %v", es.namespace, err)
	//    return err
	//}
	//es.logger.Noticef("start db for namespace: %s successful", es.namespace)

	// 2. start executor
	err := es.executor.Start()
	if err != nil {
		es.logger.Errorf("Start executor for namespace %s error, %v", es.namespace, err)
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

	err = es.service.Register(pb.FROM_EXECUTOR, &pb.RegisterMessage{
		Namespace: es.namespace,
	})
	if err != nil {
		es.logger.Errorf("Executor service register failed for namespace %s error, %v", es.namespace, err)
		return err
	}

	es.status.setState(running)
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
		es.logger.Errorf("Stop executor for namespace %s error, %v", es.namespace, err)
		return err
	}

	// 2. close related database.
	err = hyperdb.StopDatabase(es.namespace)
	if err != nil {
		es.logger.Errorf("Stop database for namespace %s error, %v", es.namespace, err)
		return err
	}

	es.status.setState(closed)
	es.logger.Noticef("Executor service for namespace: %s stopped!", es.namespace)
	return nil
}

func (es *executorServiceImpl) ProcessRequest(request interface{}) interface{} {
	//TODO Check finish logic
	logger.Critical("request : %v", request)
	logger.Critical("executor stauts: %v", es.status.getState())
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
		"txdb":{
			Svcname: "txdb",
			Version: "1,5",
			Service: hapi.NewDBTransactionAPI(namespace, es.conf),
			Public: true,
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
		//"sub": {
		//	//TODO: Inplements the webSocket subscription
		//},
		"archive": {
			Svcname: "archive",
			Version: "1.5",
			Service: hapi.NewPublicArchiveAPI(namespace, es.conf),
		},
		//TODO: implements cert API for module2
	}
}


func (es *executorServiceImpl) GetCAManager() *admittance.CAManager{
	//TODO: add CAManager to the struct.
	cm, err := admittance.NewCAManager(es.conf)
	if err != nil {
		es.logger.Error(err)
		panic("Cannot initialize the CAManager!")
	}
	es.caManager = cm
	return cm
}

func (es *executorServiceImpl) Name() string {
	return es.namespace
}