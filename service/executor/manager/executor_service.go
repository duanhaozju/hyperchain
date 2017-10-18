package manager

import (
	"hyperchain/common"
	"hyperchain/common/service"
	"hyperchain/core/executor"
	"hyperchain/service/executor/handler"
    "github.com/pkg/errors"
    pb "hyperchain/common/protos"
)

type executorService interface {
	Start() error
	Stop() error
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
}

func NewExecutorService(ns string, conf *common.Config) *executorServiceImpl {
	return &executorServiceImpl{
		namespace: ns,
		conf:      conf,
	}
}

func (es *executorServiceImpl) Start() error {
    // initial executor
	exec, err := executor.NewExecutor(es.namespace, es.conf, nil, nil)
	if err != nil {
		return errors.New("NewExecutor is fault")
	}
	es.executor = exec

	// initial service client
	s, err := service.New(es.conf.GetInt(common.EXECUTOR_PORT), "127.0.0.1", service.EXECUTOR, es.namespace)
	if err != nil {
        return errors.New("new service failed in %v")
	}
	es.service = s

    // establish connection
    err = s.Connect()
	if err != nil {
		return errors.New("service Connect failed")
	}
	// Add executor handler
	h := handler.New(exec)
	s.AddHandler(h)

    //register the namespace
    err = s.Register(pb.FROM_EXECUTOR, &pb.RegisterMessage{
        Namespace: es.namespace,
    })
    if err != nil{
        logger.Error("service Register failed")
    }
	return nil
}

func (es *executorServiceImpl) Stop() error {
    es.executor.Stop()
    return nil
}
