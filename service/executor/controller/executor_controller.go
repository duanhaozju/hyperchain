package controller

import (
	"errors"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/common/processor"
	"github.com/hyperchain/hyperchain/core/ledger/bloom"
	"github.com/hyperchain/hyperchain/namespace"
	er "github.com/hyperchain/hyperchain/service/executor/errors"
	"github.com/op/go-logging"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

const (
	NS_CONFIG_DIR_ROOT = "namespace.config_root_dir"
)

type ExecutorController interface {
	//Start start executor controller client.
	Start() error

	//Stop stop executor controller client.
	Stop() error

	//StartAllExecutorSrvs start all executor services.
	StartAllExecutorSrvs() error

	//StartExecutorServiceByName
	StartExecutorServiceByName(namespace string) error

	//StopExecutorServiceByName
	StopExecutorServiceByName(namespace string) error

	//GetExecutorServiceByName
	GetExecutorServiceByName(namespace string) executorService

	//ProcessRequest process request by namespace
	ProcessRequest(namespace string, request interface{}) interface{}

	// GetNamespaceProcessor returns the namespace processor by namespace.
	GetNamespaceProcessor(namespace string) processor.NamespaceProcessor

	//start JVM
	StartJVM() error

	//start JVM
	StopJVM() error

	//start JVM
	RestartJVM() error
}

type execControllerImpl struct {
	rwLock   *sync.RWMutex
	services map[string]executorService

	jvmManager *namespace.JvmManager
	jvmStatus  *Status

	bloomFilter bloom.TxBloomFilter //TODO: need bloomFilter here?

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
	ec.logger.Criticalf("Try to start executor client for namespace %s", namespace)

	configRootDir := ec.conf.GetString(NS_CONFIG_DIR_ROOT)
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

	// start executor client
	if ok, name := hasConfig(dirs, namespace); ok {
		if _, ok = ec.services[name]; ok {
			ec.logger.Warningf("Namespace %s for executor already exist.", name)
			return nil
		}
		conf, err := ec.readConfig(name)
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

	ec.logger.Criticalf("start executor for namespace %s successful ", namespace)

	return nil
}

//ledger register is seperated from the start-JVM.
//start part can only focus on the JVMManager start.
func (ec *execControllerImpl) StartJVM() error {
	err := ec.jvmManager.Start()
	return err
}

func (ec *execControllerImpl) StopJVM() error {
	err := ec.jvmManager.Stop()
	return err
}

func (ec *execControllerImpl) RestartJVM() error {
	ec.jvmManager.Stop()

	err := ec.jvmManager.Start()
	return err
}

func (ec *execControllerImpl) readConfig(name string) (*common.Config, error) {
	configRootDir := ec.conf.GetString(NS_CONFIG_DIR_ROOT)
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

func (ec *execControllerImpl) GetNamespaceProcessor(name string) processor.NamespaceProcessor {
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
	configRootDir := ec.conf.GetString(NS_CONFIG_DIR_ROOT)
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
			start := ec.conf.GetBool(common.START_NAMESPACE + name)
			if !start {
				continue
			}

			if _, ok := ec.services[name]; ok {
				ec.logger.Warningf("Namespace %s for executor already exist.", name)
				continue
			}
			conf, err := ec.readConfig(name)
			service := NewExecutorService(name, conf)
			err = service.Start()
			if err != nil {
				ec.logger.Error(err)
				continue
			}
			ec.jvmManager.LedgerProxy().RegisterDB(name, service.executor.FetchStateDb())
			ec.services[name] = service
			if err := ec.bloomFilter.Register(name); err != nil {
				ec.logger.Errorf("register bloom filter for namespace %s failed", name, err.Error())
			}
		} else {
			ec.logger.Errorf("Invalid folder %v", d)
		}
	}
	return nil
}