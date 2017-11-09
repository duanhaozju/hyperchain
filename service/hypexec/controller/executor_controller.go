package controller

import (
	"errors"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/common/interface"
	"hyperchain/core/ledger/bloom"
	"hyperchain/namespace"
	er "hyperchain/service/hypexec/errors"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

const (
	default_namespace  = "global"
	ns_config_dir_root = "namespace.config_root_dir"
)

type ExecutorController interface {
	Start() error

	Stop() error

	StartAllExecutorSrvs() error

	StartExecutorServiceByName(namespace string) error

	StopExecutorServiceByName(namespace string) error

	GetExecutorServiceByName(namespace string) executorService

	//ProcessRequest process request by namespace
	ProcessRequest(namespace string, request interface{}) interface{}

	// GetNamespaceProcessor returns the namespace processor by namespace.
	GetNamespaceProcessor(namespace string) intfc.NamespaceProcessor
}

type execControllerImpl struct {
	rwLock   *sync.RWMutex
	services map[string]executorService

	jvmManager  *namespace.JvmManager
	jvmStatus   *Status
	bloomFilter bloom.TxBloomFilter

	conf *common.Config

	stopEm    chan bool
	restartEm chan bool

	logger *logging.Logger
}

func newExecutorController(conf *common.Config, stopEm chan bool, restartEm chan bool) *execControllerImpl {
	js := &Status{
		state: newed,
		desc:  "newed",
		lock:  new(sync.RWMutex),
	}

	ec := &execControllerImpl{
		services:    make(map[string]executorService),
		jvmManager:  namespace.NewJvmManager(conf),
		bloomFilter: bloom.NewBloomFilterCache(conf),
		conf:        conf,
		stopEm:      stopEm,
		restartEm:   restartEm,
		logger:      common.GetLogger(common.DEFAULT_LOG, "executor-controller"),
	}
	ec.rwLock = new(sync.RWMutex)
	ec.jvmStatus = js
	return ec
}

func GetExecutorCtl(conf *common.Config, stopEm chan bool, restartEM chan bool) *execControllerImpl {
	return newExecutorController(conf, stopEm, restartEM)
}

func (ec *execControllerImpl) StartExecutorServiceByName(namespace string) error {
	ec.logger.Infof("try to start executor service for namespace %s", namespace)

	configRootDir := ec.conf.GetString(ns_config_dir_root)
	if configRootDir == "" {
		return errors.New("Namespace config root dir is not valid ")
	}

	dirs, err := ioutil.ReadDir(configRootDir)
	if err != nil {
		return err
	}

	hasConfig := func(dirs []os.FileInfo, namespace string) (bool, string) {
		for _, d := range dirs {
			if d.IsDir() {
				name := d.Name()
				if strings.Compare(name, namespace) == 0 {
					return true, name
				}
			} else {
				ec.logger.Errorf("Invalid folder %v", d)
			}
		}
		ec.logger.Errorf("No config file for namespace %s", namespace)
		return false, ""
	}

	// start executor service
	if ok, name := hasConfig(dirs, namespace); ok {
		if _, ok = ec.services[name]; ok {
			ec.logger.Warning("Namespace %s for executor already exist.", name)
			return nil
		}
		conf, err := ec.getConfig(name)
		service := NewExecutorService(name, conf)
		err = service.Start()
		if err != nil {
			ec.logger.Error(err)
			return err
		}
		ec.jvmManager.LedgerProxy().RegisterDB(name, service.executor.FetchStateDb())
		ec.services[name] = service
		if err := ec.bloomFilter.Register(name); err != nil {
			ec.logger.Error("register bloom filter failed", err.Error())
			return err
		}
	} else {
		return er.ConfigNotFoundErr("config not found for namespace %s.", namespace)
	}

	ec.logger.Info("start executor for namespace %s successful ", namespace)

	return nil
}

func (ec *execControllerImpl) getConfig(name string) (*common.Config, error) {
	configRootDir := ec.conf.GetString(ns_config_dir_root)
	if configRootDir == "" {
		return nil, errors.New("Namespace config root dir is not valid ")
	}
	nsRootPath := configRootDir + "/" + name
	if _, err := os.Stat(nsRootPath); os.IsNotExist(err) {
		ec.logger.Errorf("namespace [%s] root path doesn't exist!", name)
	}
	nsConfigDir := nsRootPath + "/config"

	// init namespace configuration(namespace.toml)
	nsConfigPath := nsConfigDir + "/namespace.toml"
	if _, err := os.Stat(nsConfigPath); os.IsNotExist(err) {
		ec.logger.Error("namespace config file doesn't exist!")
	}
	conf := common.NewConfig(nsConfigPath)
	conf.Set(common.INTERNAL_PORT, ec.conf.GetInt(common.INTERNAL_PORT))
	conf.Set(common.NAMESPACE, name)
	conf.Set(common.JVM_PORT, ec.conf.GetInt(common.JVM_PORT))
	conf.Set(common.C_JVM_START, ec.conf.GetBool(common.C_JVM_START))
	return conf, nil
}

func (ec *execControllerImpl) StopExecutorServiceByName(namespace string) error {
	if _, ok := ec.services[namespace]; ok {
		err := ec.services[namespace].Stop()
		if err != nil {
			ec.logger.Error(err)
			return err
		}
		delete(ec.services, namespace)
	} else {
		return er.StopNamespaceErr("namespace %s is not running", namespace)
	}
	return nil
}

func (ec *execControllerImpl) ProcessRequest(namespace string, request interface{}) interface{} {
	np := ec.GetNamespaceProcessor(namespace)
	if np == nil {
		ec.logger.Noticef("no namespace found for name: %s", namespace)
		return nil
	}
	return np.ProcessRequest(request)
}

func (ec *execControllerImpl) GetNamespaceProcessor(name string) intfc.NamespaceProcessor {
	ec.rwLock.RLock()
	defer ec.rwLock.RUnlock()
	ec.logger.Debugf("services : %v", ec.services)
	if es, ok := ec.services[name]; ok {
		return es
	}
	return nil
}

func (ec *execControllerImpl) GetExecutorServiceByName(name string) executorService {
	ec.rwLock.RLock()
	defer ec.rwLock.RUnlock()
	ec.logger.Debugf("services : %v", ec.services)
	if es, ok := ec.services[name]; ok {
		return es
	}
	return nil
}

func (ec *execControllerImpl) Start() error {
	// start bloom filter
	ec.bloomFilter.Start()

	//start jvm
	if ec.jvmStatus.getState() != running && ec.conf.GetBool(common.C_JVM_START) == true {
		if err := ec.jvmManager.Start(); err != nil {
			ec.logger.Error(err)
			return err
		}
		ec.jvmStatus.setState(running)
		ec.logger.Info("start hyperjvm successful")
	}
	return nil
}

func (ec *execControllerImpl) Stop() error {

	for _, es := range ec.services {
		go es.Stop()
	}
	// stop jvm
	if err := ec.jvmManager.Stop(); err != nil {
		ec.logger.Errorf("StopExecutorServiceByName hyperjvm error %v", err)
		return err
	}

	// close bloom filter
	ec.bloomFilter.Close()
	return nil
}

func (ec *execControllerImpl) StartAllExecutorSrvs() error {
	//TODO:start all executor service by config
	return ec.StartExecutorServiceByName("global")
}
