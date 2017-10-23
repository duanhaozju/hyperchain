package manager

import (
	"github.com/op/go-logging"
	hapi "hyperchain/api"
	"hyperchain/common"
	pb "hyperchain/common/protos"
	"hyperchain/core/executor"
	"hyperchain/core/ledger/chain"
	"hyperchain/hyperdb"
	"hyperchain/namespace/rpc"
	"sync"
	"hyperchain/common/client"
)

type executorService interface {
	Start() error

	Stop() error

	// ProcessRequest process request under this namespace.
	ProcessRequest(request interface{}) interface{}

	// Name returns the name of current namespace.
	Name() string
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
	es.logger.Criticalf("Init executor service %s", es.namespace)

	// 1. init DB for current executor service.
	err := chain.InitDBForNamespace(es.conf, es.namespace)
	if err != nil {
		es.logger.Errorf("Init db for namespace: %s error, %v", es.namespace, err)
		return err
	}

	// 2. initial service client
	service, err := client.New(50071, "127.0.0.1", client.EXECUTOR, es.namespace)
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


	// 5. add jsonrpc processor
	es.rpc = rpc.NewJsonRpcProcessorImpl(es.namespace, es.GetApis(es.namespace))

	return nil
}

func (es *executorServiceImpl) Start() error {
	es.logger.Noticef("try to start namespace: %s", es.namespace)
	err := es.init()
	if err != nil {
		es.logger.Errorf("Executor service initialization failed %v", err)
		return err
	}
	// 1. start executor and service client
	//err = hyperdb.StartDatabase(es.conf, es.namespace)
	//if err != nil {
	//    es.logger.Errorf("Start database for namespace %s error, %v", es.namespace, err)
	//    return err
	//}
	//es.logger.Noticef("start db for namespace: %s successful", es.namespace)

	// 2. start executor
	err = es.executor.Start()
	if err != nil {
		es.logger.Errorf("Start executor for namespace %s error, %v", es.namespace, err)
		return err
	}

    //es.executorApi = api.NewExecutorApi(es.executor, es.namespace)

	// 3. establish connection
	err = es.service.Connect()
	if err != nil {
		es.logger.Errorf("Establish connection for namespace %s error, %v", es.namespace, err)
		return err
	}

	// 4. register the namespace
	err = es.service.Register(pb.FROM_EXECUTOR, &pb.RegisterMessage{
		Namespace: es.namespace,
	})
	if err != nil {
		es.logger.Errorf("Executor service register failed for namespace %s error, %v", es.namespace, err)
		return err
	}
	return nil
}

func (es *executorServiceImpl) Stop() error {
	es.logger.Noticef("try to stop namespace: %s", es.namespace)

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

	es.logger.Noticef("namespace: %s stopped!", es.namespace)
	return nil
}

func (es *executorServiceImpl) ProcessRequest(request interface{}) interface{} {
	//TODO Check finish logic
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
	}
}

func (es *executorServiceImpl) Name() string  {
	return es.namespace
}
