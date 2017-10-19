package manager

import (
	"hyperchain/common"
	"hyperchain/common/service"
	"hyperchain/core/executor"
	"hyperchain/service/executor/handler"
    pb "hyperchain/common/protos"
    "github.com/op/go-logging"
    "hyperchain/core/ledger/chain"
    "hyperchain/hyperdb"
    "hyperchain/service/executor/api"
)

type executorService interface {
	Start() error
	Stop() error

	// ProcessRequest process request under this namespace.
	ProcessRequest(request interface{}) interface{}
}

type executorServiceImpl struct {
	// namespace
	namespace string
	// real executor object
	executor *executor.Executor
	// manager the connection with service
	service *service.ServiceClient
	// config
	conf *common.Config
	// logger
	logger *logging.Logger

	executorApi *api.ExecutorApi

}

func NewExecutorService(ns string, conf *common.Config) *executorServiceImpl {
    // init hyper logger for executor service
    conf.Set(common.NAMESPACE, ns)
    if err := common.InitHyperLogger(ns, conf); err != nil {
        return nil
    }

    return &executorServiceImpl{
		namespace: ns,
		conf:      conf,
		logger:    common.GetLogger(ns, "executor_service"),
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

    // 2. initial executor
    executor, err := executor.NewExecutor(es.namespace, es.conf, nil, nil)
    if err != nil {
        es.logger.Errorf("Init executor service for namespace %s error, %v", es.namespace, err)
        return err
    }
    executor.CreateInitBlock(es.conf)
    es.executor = executor

    // 3. initial service client
    service, err := service.New(50071, "127.0.0.1", service.EXECUTOR, es.namespace)
    if err != nil {
        es.logger.Errorf("Init service client for namespace %s error, %v", es.namespace, err)
        return err
    }
    es.service = service

    // 4. add executor handler
    h := handler.New(executor)
    service.AddHandler(h)

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

    es.executorApi = api.NewExecutorApi(es.executor, es.namespace)

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
    if err != nil{
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

func (es *executorServiceImpl) ProcessRequest(request interface{}) interface{}{
	//TODO Need to finish logic
	return nil
}
