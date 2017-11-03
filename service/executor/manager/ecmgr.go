package manager

import (
	"errors"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/common/interface"
	"hyperchain/namespace"
	er "hyperchain/service/executor/errors"
	"io/ioutil"
	"os"
	"strings"
	"sync"
    "hyperchain/core/ledger/bloom"
)

var logger *logging.Logger

const (
	DEFAULT_NAMESPACE  = "global"
	NS_CONFIG_DIR_ROOT = "namespace.config_root_dir"
)

type ExecutorManager interface {
	Start(namespace string) error

	Stop(namespace string) error

	GetExecutorServiceByName(name string) executorService

	// ProcessRequest dispatches received requests to corresponding namespace
	// processor.
	// Requests are sent from RPC layer, so responses are returned to RPC layer
	// with the certain namespace.
	ProcessRequest(namespace string, request interface{}) interface{}

	// GetNamespaceProcessorName returns the namespace instance by name.
	GetNamespaceProcessor(name string) intfc.NamespaceProcessor
}

type ecManagerImpl struct {

	// manager the connect, it can be included in the ExecutorRemote when it add
	services map[string]executorService

	jvmManager *namespace.JvmManager

    // bloom filter is the transaction bloom filter, helps to do transaction
    // duplication checking
    bloomFilter bloom.TxBloomFilter

	// conf is the global config file of the system, contains global configs
	// of the node
	conf *common.Config

	stopEm chan bool

	restartEm chan bool

	rwLock *sync.RWMutex

	jvmStatus *Status
}

func newExecutorManager(conf *common.Config, stopEm chan bool, restartEm chan bool) *ecManagerImpl {
	js := &Status{
		state: newed,
		desc:  "newed",
		lock:  new(sync.RWMutex),
	}

	em := &ecManagerImpl{
		services:   make(map[string]executorService),
		jvmManager: namespace.NewJvmManager(conf),
		bloomFilter: bloom.NewBloomFilterCache(conf),
		conf:       conf,
		stopEm:     stopEm,
		restartEm:  restartEm,
	}
	em.rwLock = new(sync.RWMutex)
	em.jvmStatus = js
	return em
}

func GetExecutorMgr(conf *common.Config, stopEm chan bool, restartEM chan bool) *ecManagerImpl {
	logger = common.GetLogger(common.DEFAULT_LOG, "executorMgr")
	return newExecutorManager(conf, stopEm, restartEM)
}

func (em *ecManagerImpl) Start(namespace string) error {
	configRootDir := em.conf.GetString(NS_CONFIG_DIR_ROOT)
	logger.Critical("namespace configRootDir:", configRootDir)
	if configRootDir == "" {
		return errors.New("Namespace config root dir is not valid ")
	}
	dirs, err := ioutil.ReadDir(configRootDir)
	if err != nil {
		return err
	}

    // start bloom filter
    em.bloomFilter.Start()

	//start the specify executor service

	hasConfig := func(dirs []os.FileInfo, namespace string) (bool, string) {
		for _, d := range dirs {
			if d.IsDir() {
				name := d.Name()
				if strings.Compare(name, namespace) == 0 {
					return true, name
				}
			} else {
				logger.Errorf("Invalid folder %v", d)
			}
		}
		logger.Errorf("No config file for namespace %s", namespace)
		return false, ""
	}

	// start executor service
	if ok, name := hasConfig(dirs, namespace); ok {
		if _, ok = em.services[name]; ok {
			logger.Warning("Namespace %s for executor already exist.", name)
			return nil
		}
		conf, err := em.getConfig(name)
		service := NewExecutorService(name, conf)
		err = service.Start()
		if err != nil {
			logger.Error(err)
			return err
		}
		em.jvmManager.LedgerProxy().RegisterDB(name, service.executor.FetchStateDb())
		em.services[name] = service
	} else {
		return er.ConfigNotFoundErr("Config not found for namespace %s.", namespace)
	}
	logger.Info("Start executor for namespace %s ok.", namespace)

	//start jvm
	if em.jvmStatus.getState() != running {
		if em.conf.GetBool(common.C_JVM_START) == true {
			if err := em.jvmManager.Start(); err != nil {
				logger.Error(err)
				return err
			}
			em.jvmStatus.setState(running)
			logger.Info("Start JVM ok.")
		}
	}
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
	conf.Set(common.INTERNAL_PORT, em.conf.GetInt(common.INTERNAL_PORT))
	conf.Set(common.NAMESPACE, name)
	conf.Set(common.JVM_PORT, em.conf.GetInt(common.JVM_PORT))
	conf.Set(common.C_JVM_START, em.conf.GetBool(common.C_JVM_START))
	return conf, nil
}

func (em *ecManagerImpl) Stop(namespace string) error {
	// stop all executor service
	if _, ok := em.services[namespace]; ok {
		err := em.services[namespace].Stop()
		if err != nil {
			logger.Error(err)
			return err
		}
		delete(em.services, namespace)
	} else {
		return er.StopNamespaceErr("namespace %s is not running", namespace)
	}

	// stop jvm
	if len(em.services) == 0 {
		if err := em.jvmManager.Stop(); err != nil {
			logger.Errorf("Stop hyperjvm error %v", err)
			return err
		}
	}

	// close bloom filter
	em.bloomFilter.Close()

	return nil
}

func (em *ecManagerImpl) ProcessRequest(namespace string, request interface{}) interface{} {
	np := em.GetNamespaceProcessor(namespace)
	if np == nil {
		logger.Noticef("no namespace found for name: %s", namespace)
		return nil
	}
	return np.ProcessRequest(request)
}

func (em *ecManagerImpl) GetNamespaceProcessor(name string) intfc.NamespaceProcessor {
	em.rwLock.RLock()
	defer em.rwLock.RUnlock()
	logger.Critical("services : %v", em.services)
	if es, ok := em.services[name]; ok {
		return es
	}
	return nil
}

func (em *ecManagerImpl) GetExecutorServiceByName(name string) executorService {
	em.rwLock.RLock()
	defer em.rwLock.RUnlock()
	logger.Critical("services : %v", em.services)
	if es, ok := em.services[name]; ok {
		return es
	}
	return nil
}
