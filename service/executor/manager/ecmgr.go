package manager

import (
	"hyperchain/common"
	"hyperchain/core/executor"
    "hyperchain/common/service"
    "io/ioutil"
    "errors"
    "github.com/op/go-logging"
    "hyperchain/namespace"
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

	// manager the connect, it can be included in the ExecutorRemote when it add
	services map[string]executorService

    jvmManager *namespace.JvmManager
	// conf is the global config file of the system, contains global configs
	// of the node
	conf *common.Config

	stopEm    chan bool
	restartEm chan bool
}


func newExecutorManager(conf *common.Config, stopEm chan bool, restartEm chan bool) *ecManagerImpl {
	em := &ecManagerImpl{
		services:  make(map[string]*executorService),
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
    configRootDir := em.conf.GetString(NS_CONFIG_DIR_ROOT)
    if configRootDir == "" {
        return errors.New("Namespace config root dir is not valid ")
    }
    dirs, err := ioutil.ReadDir(configRootDir)
    if err != nil {
        return err
    }

    // start all executor service
    for _, d := range dirs {
        if d.IsDir() {
            name := d.Name()
            start := em.conf.GetBool(common.START_NAMESPACE + name)
            if !start {
                continue
            }
            service := NewExecutorService(name, em.conf)
            err := service.Start()
            if err != nil {
                logger.Error(err)
            }
            em.services[name] = service
        } else {
            logger.Errorf("Invalid folder %v", d)
        }
    }

    // start jvm
    if em.conf.GetBool(common.C_JVM_START) == true {
        if err := em.jvmManager.Start(); err != nil {
            logger.Error(err)
            return err
        }
    }
    return nil
}

func (em *ecManagerImpl) Stop() error {
    // stop all executor service
    for ns := range em.services {
        err := em.services[ns].Stop()
        if err != nil {
            logger.Error(err)
        }
    }
    // stop jvm
    if err := em.jvmManager.Stop(); err != nil {
        logger.Errorf("Stop hyperjvm error %v", err)
    }
    return nil
}