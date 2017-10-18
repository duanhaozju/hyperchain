package manager

import (
	"hyperchain/common"
	"hyperchain/core/executor"
    "hyperchain/common/service"
    "io/ioutil"
    "errors"
    "github.com/op/go-logging"
    "hyperchain/namespace"
    "hyperchain/service/executor/handler"
    pb "hyperchain/common/protos"
)

var logger *logging.Logger

const (
    DEFAULT_NAMESPACE  = "global"
    NS_CONFIG_DIR_ROOT = "namespace.config_root_dir"
)

type ExecutorManager interface {

	Start() error

	Stop() error

}

type ecManagerImpl struct {

	executors map[string]*executor.Executor

	// manager the connect, it can be included in the ExecutorRemote when it add
	services map[string]*service.ServiceClient

    jvmManager *namespace.JvmManager
	// conf is the global config file of the system, contains global configs
	// of the node
	conf *common.Config

	stopEm    chan bool
	restartEm chan bool
}

func newExecutorManager(conf *common.Config, stopEm chan bool, restartEm chan bool) *ecManagerImpl {
	em := &ecManagerImpl{
		executors:  make(map[string]*executor.Executor),
        jvmManager:  namespace.NewJvmManager(conf),
        conf:        conf,
		stopEm:      stopEm,
		restartEm:   restartEm,
	}
	return em
}

func GetExecutorMgr(conf *common.Config, stopEm chan bool, restartEM chan bool) *ecManagerImpl{
    logger = common.GetLogger(common.DEFAULT_LOG, "executorMgr")

    return newExecutorManager(conf, stopEm, restartEM)
}


func (em *ecManagerImpl) Start() error {
    //TODO: 读配置文件,开启多个executor实例
    configRootDir := em.conf.GetString(NS_CONFIG_DIR_ROOT)
    if configRootDir == "" {
        return errors.New("Namespace config root dir is not valid ")
    }
    dirs, err := ioutil.ReadDir(configRootDir)
    if err != nil {
        return err
    }
    for _, d := range dirs {
        if d.IsDir() {
            name := d.Name()
            start := em.conf.GetBool(common.START_NAMESPACE + name)
            if !start {
                continue
            }

            exec, err := executor.NewExecutor(name, em.conf, nil, nil)
            if err != nil {
                logger.Errorf("NewExecutor is fault")
            }
            em.executors[name] = exec

            s, err := service.New(em.conf.GetInt(common.EXECUTOR_PORT), "127.0.0.1", service.EXECUTOR, name)
            if err != nil {
                logger.Errorf("new service failed in %v")
            }
			//establish connection
            err = s.Connect()
            if err != nil {
                logger.Error("service Connect failed")
            }
			//register the namespace
            err = s.Register(service.EXECUTOR, &pb.RegisterMessage{
                Namespace: name,
            })
            if err != nil{
                logger.Error("service Register failed")
            }

            // Add executor handler
            h := handler.New(exec)
            s.AddHandler(h)

            em.services[name] = s

        } else {
            logger.Errorf("Invalid folder %v", d)
        }
    }
    if em.conf.GetBool(common.C_JVM_START) == true {
        if err := em.jvmManager.Start(); err != nil {
            logger.Error(err)
            return err
        }
    }
    return nil
}

func (em *ecManagerImpl) Stop() error {

    // 1. stop executor
    for ns := range em.executors {
        err := em.executors[ns].Stop()
        if err != nil {
            logger.Error(err)
        }
    }

    // 2. stop jvm
    if err := em.jvmManager.Stop(); err != nil {
        logger.Errorf("Stop hyperjvm error %v", err)
    }
    return nil
}