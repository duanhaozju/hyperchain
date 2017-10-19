package manager

import (
	"errors"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/namespace"
	"io/ioutil"
	"os"
	"sync"
)

var logger *logging.Logger

const (
	DEFAULT_NAMESPACE  = "global"
	NS_CONFIG_DIR_ROOT = "namespace.config_root_dir"
)

type ExecutorManager interface {
	Start() error

	Stop() error

	ProcessRequest(namespace string, request interface{}) interface{}

	GetExecutorServiceByName(name string) executorService
}

type ecManagerImpl struct {

	// manager the connect, it can be included in the ExecutorRemote when it add
	services map[string]executorService

	jvmManager *namespace.JvmManager
	// conf is the global config file of the system, contains global configs
	// of the node
	conf *common.Config

	stopEm chan bool

	restartEm chan bool

	rwLock *sync.RWMutex
}

func newExecutorManager(conf *common.Config, stopEm chan bool, restartEm chan bool) *ecManagerImpl {
	em := &ecManagerImpl{
		services:   make(map[string]executorService),
		jvmManager: namespace.NewJvmManager(conf),
		conf:       conf,
		stopEm:     stopEm,
		restartEm:  restartEm,
	}
	em.rwLock = new(sync.RWMutex)
	return em
}

func GetExecutorMgr(conf *common.Config, stopEm chan bool, restartEM chan bool) *ecManagerImpl {
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

			// start each executor service
			conf, err := em.getConfig(name)
			service := NewExecutorService(name, conf)
			err = service.Start()
			if err != nil {
				logger.Error(err)
			}
			em.services[name] = service
		} else {
			logger.Errorf("Invalid folder %v", d)
		}
	}

	// start jvm
	//if em.conf.GetBool(common.C_JVM_START) == true {
	//	if err := em.jvmManager.Start(); err != nil {
	//		logger.Error(err)
	//		return err
	//	}
	//}
	return nil
}

func (em *ecManagerImpl) getConfig(name string) (*common.Config, error) {
	configRootDir := em.conf.GetString(NS_CONFIG_DIR_ROOT)
	if configRootDir == "" {
		return nil, errors.New("Namespace config root dir is not valid ")
	}
	nsRootPath := configRootDir + "/" + name
	if _, err := os.Stat(nsRootPath); os.IsNotExist(err) {
		logger.Errorf("namespace [%s] root path doesn't exist!", name)
	}
	nsConfigDir := nsRootPath + "/config"

	// init namespace configuration(namespace.toml)
	nsConfigPath := nsConfigDir + "/namespace.toml"
	if _, err := os.Stat(nsConfigPath); os.IsNotExist(err) {
		logger.Error("namespace config file doesn't exist!")
	}
	conf := common.NewConfig(nsConfigPath)
	return conf, nil
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

func (em *ecManagerImpl) ProcessRequest(namespace string, request interface{}) interface{} {
	es := em.GetExecutorServiceByName(namespace)
	if es == nil {
		logger.Noticef("no namespace found for name: %s", namespace)
		return nil
	}
	return es.ProcessRequest(request)
}

func (em *ecManagerImpl) GetExecutorServiceByName(name string) executorService {
	em.rwLock.RLock()
	defer em.rwLock.RUnlock()
	if es, ok := em.services[name]; ok {
		return es
	}
	return nil
}
