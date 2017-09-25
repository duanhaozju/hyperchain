//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

//Package namespace provide mechanism to manage namespaces and request process
package namespace

import (
	"errors"
	"fmt"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/core/db_utils"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

var logger *logging.Logger

const (
	DEFAULT_NAMESPACE  = "global"
	NS_CONFIG_DIR_ROOT = "namespace.config_root_dir"
)

var once sync.Once
var nr NamespaceManager

//NamespaceManager namespace manager.
type NamespaceManager interface {
	//Start start namespace manager service.
	Start() error
	//Stop stop namespace manager.
	Stop() error
	//List list all namespace names in system.
	List() []string
	//Register register a new namespace.
	Register(name string) error
	//DeRegister de-register namespace from system by name.
	DeRegister(name string) error
	//StartJvm start jvm manager
	StartJvm() error
	//StopJvm stop jvm manager
	StopJvm() error
	//RestartJvm restart jvm manager
	RestartJvm() error
	//GetNamespaceByName get namespace instance by name.
	GetNamespaceByName(name string) Namespace
	//ProcessRequest process received request
	ProcessRequest(namespace string, request interface{}) interface{}
	//StartNamespace start namespace by name.
	StartNamespace(name string) error
	//StopNamespace stop namespace by name.
	StopNamespace(name string) error
	//RestartNamespace restart namespace by name.
	RestartNamespace(name string) error
	//GlobalConfig global configuration of the system.
	GlobalConfig() *common.Config
	//GetStopFlag returns the flag of stop hyperchain server
	GetStopFlag() chan bool
	//GetRestartFlag returns the flag of restart hyperchain server
	GetRestartFlag() chan bool
}

//nsManagerImpl implementation of NsRegistry.
type nsManagerImpl struct {
	rwLock      *sync.RWMutex
	namespaces  map[string]Namespace
	jvmManager  *JvmManager
	bloomfilter *db_utils.BloomFilterCache // transaciton bloom filter
	// help to do transaction duplication checking
	conf *common.Config
	stopHp      chan bool
	restartHp   chan bool
}

//NewNsManager new a namespace manager
func newNsManager(conf *common.Config, stopHp chan bool, restartHp chan bool) *nsManagerImpl {
	nr := &nsManagerImpl{
		namespaces:  make(map[string]Namespace),
		conf:        conf,
		jvmManager:  NewJvmManager(conf),
		bloomfilter: db_utils.NewBloomCache(conf),
		stopHp:      stopHp,
		restartHp:   restartHp,
	}
	nr.rwLock = new(sync.RWMutex)
	err := nr.init()
	if err != nil {
		panic(err)
	}
	return nr
}

//GetNamespaceManager get namespace registry instance.
func GetNamespaceManager(conf *common.Config, stopHp chan bool, restartHp chan bool) NamespaceManager {
	logger = common.GetLogger(common.DEFAULT_LOG, "nsmgr")
	once.Do(func() {
		nr = newNsManager(conf, stopHp, restartHp)
	})
	return nr
}

//init the namespace registry by configuration.
func (nr *nsManagerImpl) init() error {
	configRootDir := nr.conf.GetString(NS_CONFIG_DIR_ROOT)
	if configRootDir == "" {
		return errors.New("Namespace config root dir is not valid")
	}
	dirs, err := ioutil.ReadDir(configRootDir)
	if err != nil {
		return err
	}
	for _, d := range dirs {
		if d.IsDir() {
			name := d.Name()
			start := nr.conf.GetBool(common.START_NAMESPACE + name)
			if !start {
				continue
			}
			nr.Register(name)
		} else {
			logger.Errorf("Invalid folder %v", d)
		}
	}
	return nil
}

//Start start namespace registry service.
//which will also start all namespace in this Namespace Registry
func (nr *nsManagerImpl) Start() error {
	nr.rwLock.RLock()
	defer nr.rwLock.RUnlock()
	for name := range nr.namespaces {
		go func(name string) {
			err := nr.StartNamespace(name)
			if err != nil {
				logger.Errorf("namespace %s start failed, %v", name, err)
			}
		}(name)
	}
	if nr.conf.GetBool(common.C_JVM_START) == true {
		if err := nr.jvmManager.Start(); err != nil {
			logger.Error(err)
			return err
		}
	}
	return nil
}

//Stop stop namespace registry.
func (nr *nsManagerImpl) Stop() error {
	logger.Noticef("Try to stop NamespaceManager ...")
	nr.rwLock.RLock()
	defer nr.rwLock.RUnlock()
	for name := range nr.namespaces {
		err := nr.StopNamespace(name)
		if err != nil {
			logger.Errorf("namespace %s stop failed, %v", name, err)
			return err
		}
	}
	// err := nr.hyperjvm.Stop()
	// if err != nil {
	//	logger.Errorf("Stop hyperjvm error %v", err)
	//	return err
	//}
	if err := nr.jvmManager.Stop(); err != nil {
		logger.Errorf("Stop hyperjvm error %v", err)
	}
	nr.bloomfilter.Close()
	logger.Noticef("NamespaceManager stopped!")
	return nil
}

//List list all namespace names in system.
func (nr *nsManagerImpl) List() (names []string) {
	nr.rwLock.RLock()
	defer nr.rwLock.RUnlock()
	for name := range nr.namespaces {
		names = append(names, name)
	}
	return names
}

func (nr *nsManagerImpl) checkNamespaceName(name string) bool {
	if name == "global" {
		return true
	}
	if len(name) == 13 && strings.HasPrefix(name, "ns_") {
		return true
	}
	return false
}

//Register register a new namespace, by the new namespace config dir.
func (nr *nsManagerImpl) Register(name string) error {
	logger.Noticef("Register namespace: %s", name)
	if _, ok := nr.namespaces[name]; ok {
		logger.Warningf("namespace [%s] has been registered", name)
		return ErrRegistered
	}
	configRootDir := nr.conf.GetString(NS_CONFIG_DIR_ROOT)
	if configRootDir == "" {
		return errors.New("Namespace config root dir is not valid")
	}
	nsRootPath := configRootDir + "/" + name
	if _, err := os.Stat(nsRootPath); os.IsNotExist(err) {
		logger.Errorf("namespace [%s] root path doesn't exist!", name)
		return ErrNonExistConfig
	}
	nsConfigDir := nsRootPath + "/config"
	nsConfig, err := nr.constructConfigFromDir(name, nsConfigDir)
	if err != nil {
		return err
	}
	nsConfig.Set(common.NAMESPACE, name)
	nsConfig.Set(common.C_JVM_START, nr.GlobalConfig().GetBool(common.C_JVM_START))
	delFlag := make(chan bool)
	ns, err := GetNamespace(name, nsConfig, delFlag)
	if err != nil {
		logger.Errorf("Construct namespace %s error, %v", name, err)
		return ErrCannotNewNs
	}
	nr.addNamespace(ns)
	if err := nr.bloomfilter.Register(name); err != nil {
		logger.Error("register bloom filter failed", err.Error())
		return err
	}
	if err = updateNamespaceStartConfig(name, nr.conf); err != nil {
		logger.Criticalf("Update namespace start for [%s] config failed", name)
	}
	go nr.ListenDelNode(name, delFlag)
	return nil
}

func updateNamespaceStartConfig(name string, conf *common.Config) error {
	if !conf.ContainsKey(fmt.Sprintf("namespace.start.%s", name)) {
		return common.SeekAndAppend("[namespace.start]", conf.GetString(common.GLOBAL_CONFIG_PATH),
			fmt.Sprintf("    %s = true", name))
	}
	return nil
}

func (nr *nsManagerImpl) removeNamespace(name string)  {
	nr.rwLock.Lock()
	delete(nr.namespaces, name)
	nr.rwLock.Unlock()
}

func (nr *nsManagerImpl) addNamespace(ns Namespace)  {
	nr.rwLock.Lock()
	nr.namespaces[ns.Name()] = ns
	nr.rwLock.Unlock()
}

//DeRegister de-register namespace from system by name.
func (nr *nsManagerImpl) DeRegister(name string) error {
	logger.Criticalf("Try to deregister the namespace:%s ", name)
	if ns, ok := nr.namespaces[name]; ok {
		if ns.Status().state == running {
			logger.Noticef("namespace: %s is running, stop it first", name)
			ns.Stop()
		}
		nr.removeNamespace(name)
	} else {
		logger.Warningf("namespace %s not exist, please register first.", name)
		return ErrNoSuchNamespace
	}
	nr.bloomfilter.UnRegister(name)
	logger.Criticalf("namespace: %s stopped", name)
	//TODO: need to delete the data?
	return nil
}

//GetNamespaceByName get namespace instance by name.
func (nr *nsManagerImpl) GetNamespaceByName(name string) Namespace {
	nr.rwLock.RLock()
	defer nr.rwLock.RUnlock()
	if ns, ok := nr.namespaces[name]; ok {
		return ns
	}
	return nil
}

//ProcessRequest process received request
func (nr *nsManagerImpl) ProcessRequest(namespace string, request interface{}) interface{} {
	ns := nr.GetNamespaceByName(namespace)
	if ns == nil {
		logger.Noticef("no namespace found for name: %s", namespace)
		return nil
	}
	return ns.ProcessRequest(request)
}

//StartNamespace start namespace by name
func (nr *nsManagerImpl) StartNamespace(name string) error {
	nr.rwLock.RLock()
	defer nr.rwLock.RUnlock()
	if ns, ok := nr.namespaces[name]; ok {
		if err := ns.Start(); err != nil {
			ns.Stop() //start failed, try to stop some started components
			return err
		} else {
			nr.jvmManager.ledgerProxy.RegisterDB(name, ns.GetExecutor().FetchStateDb())
			return nil
		}

	}
	logger.Errorf("No namespace instance for %s found", name)
	return ErrInvalidNs
}

//StopNamespace stop namespace by name
func (nr *nsManagerImpl) StopNamespace(name string) error {
	nr.rwLock.RLock()
	defer nr.rwLock.RUnlock()
	if ns, ok := nr.namespaces[name]; ok {
		if err := ns.Stop(); err != nil {
			return err
		} else {
			nr.jvmManager.ledgerProxy.UnRegister(name)
			return nil
		}
	}
	logger.Errorf("No namespace instance for %s found")
	return ErrInvalidNs
}

//RestartNamespace restart namespace by name
func (nr *nsManagerImpl) RestartNamespace(name string) error {
	nr.rwLock.RLock()
	defer nr.rwLock.RUnlock()
	if ns, ok := nr.namespaces[name]; ok {
		return ns.Restart()
	}
	logger.Errorf("No namespace instance for %s found")
	return ErrInvalidNs
}

//GlobalConfig global configuration of the system.
func (nr *nsManagerImpl) GlobalConfig() *common.Config {
	return nr.conf
}

//StartJvm starts the jvm manager
func (nr *nsManagerImpl) StartJvm() error {
	if err := nr.jvmManager.Start(); err != nil {
		return err
	}
	return nil
}

//StopJvm stops the jvm manager
func (nr *nsManagerImpl) StopJvm() error {
	if err := nr.jvmManager.Stop(); err != nil {
		return err
	}
	return nil
}

//RestartJvm restarts the jvm manager
func (nr *nsManagerImpl) RestartJvm() error {
	if err := nr.jvmManager.Stop(); err != nil {
		return err
	}
	if err := nr.jvmManager.Start(); err != nil {
		return err
	}
	return nil
}

//ListenDelNode listen delete node event
func (nr *nsManagerImpl) ListenDelNode(name string, delFlag chan bool) {
	for {
		select {
		case <-delFlag:
			nr.StopNamespace(name)
			nr.DeRegister(name)
		}
	}
}

//GetStopFlag returns the flag of stop hyperchain server
func (nr *nsManagerImpl) GetStopFlag() chan bool {
	return nr.stopHp
}

//GetRestartFlag returns the flag of restart hyperchain server
func (nr *nsManagerImpl) GetRestartFlag() chan bool {
	return nr.restartHp
}
