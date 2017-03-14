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

func init() {
	logger = logging.MustGetLogger("namespace")
}

var once sync.Once
var nr NamespaceManager

//NamespaceManager namespace manager.
type NamespaceManager interface {
	//Start start namespace manager service.
	Start()
	//Stop stop namespace manager.
	Stop()
	//List list all namespace names in system.
	List() []string
	//Register register a new namespace.
	Register(ns Namespace)
	//DeRegister de-register namespace from system by name.
	DeRegister(name string)
	//GetNamespaceByName get namespace instance by name.
	GetNamespaceByName(name string) Namespace
	//ProcessRequest process received request
	ProcessRequest(request interface{}) interface{}
	//StartNamespace start namespace by name
	StartNamespace(name string)
	//StopNamespace stop namespace by name
	StopNamespace(name string)
	//RestartNamespace restart namespace by name
	RestartNamespace(name string)
}

//nsManagerImpl implementation of NsRegistry.
type nsManagerImpl struct {
	namespaces map[string]Namespace
	conf       *common.Config
}

//NewNsRegistry new a namespace registry
func newNsRegistry(conf *common.Config) *nsManagerImpl {
	nri := &nsManagerImpl{
		namespaces: make(map[string]Namespace),
		conf:       conf,
	}
	err := nri.init()
	if err != nil {
		panic(err)
	}
	return nri
}

//GetNamespaceManager get namespace registry instance.
func GetNamespaceManager(conf *common.Config) NamespaceManager {
	once.Do(func() {
		nr = newNsRegistry(conf)
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
			nsConfigDir := configRootDir +"/" + d.Name() + "/config"
			nsConfig := ConstructConfigFromDir(nsConfigDir)
			nsConfig.Set(NS_CONFIG_DIR, nsConfigDir)
			ns, err := GetNamespace(name, nsConfig)
			if err != nil {
				logger.Errorf("Construct namespace %s error, %v", name, err)
			}
			nr.namespaces[name] = ns
		} else {
			logger.Errorf("Invalid folder %v", d)
		}
	}
	//TODO:....
	return nil
}

//Start start namespace registry service.
//which will also start all namespace in this Namespace Registry
func (nr *nsManagerImpl) Start() {
	for name, ns := range nr.namespaces {
		err := ns.Start()
		if err != nil {
			logger.Errorf("namespace %s start failed, %v", name, err)
		}
	}
}

//Stop stop namespace registry.
func (nr *nsManagerImpl) Stop() {
	for name, ns := range nr.namespaces {
		err := ns.Stop()
		if err != nil {
			logger.Errorf("namespace %s stop failed, %v", name, err)
		}
	}
}

//List list all namespace names in system.
func (nr *nsManagerImpl) List() (names []string) {
	for name := range nr.namespaces {
		names = append(names, name)
	}
	return names
}

//Register register a new namespace.
func (nr *nsManagerImpl) Register(ns Namespace) {
	//TODO
}

//DeRegister de-register namespace from system by name.
func (nr *nsManagerImpl) DeRegister(name string) {
	//TODO:
}

//GetNamespaceByName get namespace instance by name.
func (nr *nsManagerImpl) GetNamespaceByName(name string) Namespace {
	return nr.namespaces[name]
}

//ProcessRequest process received request
func (nr *nsManagerImpl) ProcessRequest(request interface{}) interface{} {
	req := request.(*common.RPCRequest)
	if _, ok := nr.namespaces[req.Namespace]; !ok {
		return &common.RPCResponse{
			Namespace: req.Namespace,
			Error: &common.NamespaceNotFound{req.Namespace},
		}
	}
	return nr.namespaces[req.Namespace].ProcessRequest(req)
}

//StartNamespace start namespace by name
func (nr *nsManagerImpl) StartNamespace(name string) {
	//TODO: start a namespace by name
	//1. check if the namespace instance already registered in the system
	//2. check the namespace state
	//3. register the namespace or init a new namespace
}

//StopNamespace stop namespace by name
func (nr *nsManagerImpl) StopNamespace(name string) {
	//TODO: stop namespace by name
}

//RestartNamespace restart namespace by name
func (nr *nsManagerImpl) RestartNamespace(name string) {
	//TODO: restart a namespace
}
