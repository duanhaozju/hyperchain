//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

//Package namespace provide mechanism to manage namespaces and request process
package namespace

import (
	"errors"
	"github.com/op/go-logging"
	"hyperchain/common"
	"io/ioutil"
	"sync"
)

var logger *logging.Logger

var (
	ErrInvalidNs   = errors.New("namespace/nsmgr: invalid namespace")
	ErrCannotNewNs = errors.New("namespace/nsmgr: can not new namespace")
	ErrNsClosed    = errors.New("namespace/nsmgr: namespace closed")
)

const (
	DEFAULT_NAMESPACE  = "global"
	NS_CONFIG_DIR_ROOT = "global.nsConfigRootPath"
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
}

//nsManagerImpl implementation of NsRegistry.
type nsManagerImpl struct {
	rwLock     *sync.RWMutex
	namespaces map[string]Namespace
	jvmManager *JvmManager
	conf       *common.Config
}

//NewNsManager new a namespace manager
func newNsManager(conf *common.Config) *nsManagerImpl {
	nr := &nsManagerImpl{
		namespaces: make(map[string]Namespace),
		conf:       conf,
		jvmManager: NewJvmManager(),
	}
	nr.rwLock = new(sync.RWMutex)
	err := nr.init()
	if err != nil {
		panic(err)
	}
	return nr
}

//GetNamespaceManager get namespace registry instance.
func GetNamespaceManager(conf *common.Config) NamespaceManager {
	logger = common.GetLogger(common.DEFAULT_LOG, "nsmgr")
	once.Do(func() {
		nr = newNsManager(conf)
	})
	return nr
}

//init the namespace registry by configuration.
func (nr *nsManagerImpl) init() error {
	//init all namespace instance by configuration
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
		err := nr.StartNamespace(name)
		if err != nil {
			logger.Errorf("namespace %s start failed, %v", name, err)
			return err
		}
	}
	if err := nr.jvmManager.ledgerProxy.Server(); err != nil {
		return err
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

//Register register a new namespace, by the new namespace config dir.
func (nr *nsManagerImpl) Register(name string) error {
	logger.Noticef("Register namespace: %s", name)
	configRootDir := nr.conf.GetString(NS_CONFIG_DIR_ROOT)
	if configRootDir == "" {
		return errors.New("Namespace config root dir is not valid")
	}
	nsConfigDir := configRootDir + "/" + name + "/config"
	nsConfig := nr.constructConfigFromDir(nsConfigDir)
	nsConfig.Set(common.NAMESPACE, name)
	ns, err := GetNamespace(name, nsConfig)
	if err != nil {
		logger.Errorf("Construct namespace %s error, %v", name, err)
		return ErrCannotNewNs
	}
	nr.rwLock.Lock()
	nr.namespaces[name] = ns
	nr.rwLock.Unlock()
	return nil
}

//DeRegister de-register namespace from system by name.
func (nr *nsManagerImpl) DeRegister(name string) error {
	logger.Criticalf("Try to deregister the namespace:%s ", name)
	if ns, ok := nr.namespaces[name]; ok {
		if ns.Status().state == running {
			logger.Noticef("namespace: %s is running, stop it first", name)
			ns.Stop()
			nr.rwLock.Lock()
			delete(nr.namespaces, name)
			nr.rwLock.Unlock()
		}

	} else {
		return ErrInvalidNs
	}
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
			return err
		} else {
			nr.jvmManager.ledgerProxy.Register(name, ns.GetExecutor().FetchStateDb())
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
