//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

//Package namespace provide mechanism to manage namespaces and request process
package namespace

import (
	"github.com/op/go-logging"
	"hyperchain/common"
	"sync"
)

var logger *logging.Logger

func init() {
	logger = logging.MustGetLogger("namespace")
}

var once sync.Once
var nr NamespaceRegistry

//NsRegistry namespace registry
type NamespaceRegistry interface {
	//Init init the namespace registry
	Init()
	//Start start namespace registry service.
	Start()
	//Stop stop namespace registry.
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
}

//nsRegistryImpl implementation of NsRegistry
type nsRegistryImpl struct {
	namespaces map[string]Namespace
}

//NewNsRegistry new a namespace registry
func newNsRegistry(conf *common.Config) NamespaceRegistry {
	//TODO: new namespace registry instance
	return nil
}

//GetNamespaceRegistry get namespace registry instance
func GetNamespaceRegistry(conf *common.Config) NamespaceRegistry {
	once.Do(func() {
		nr = newNsRegistry(conf)
		nr.Init()
	})
	return nr
}

//Init init the namespace registry
func (nr *nsRegistryImpl) Init() {
	//TODO: init function of NamespaceRegistry
}

//Start start namespace registry service.
func (nr *nsRegistryImpl) Start() {
	//TODO
}

//Stop stop namespace registry.
func (nr *nsRegistryImpl) Stop() {
	//TODO
}

//List list all namespace names in system.
func (nr *nsRegistryImpl) List() []string {
	//TODO
	return nil
}

//Register register a new namespace.
func (nr *nsRegistryImpl) Register(ns Namespace) {
	//TODO
}

//DeRegister de-register namespace from system by name.
func (nr *nsRegistryImpl) DeRegister(name string) {
	//TODO:
}

//GetNamespaceByName get namespace instance by name.
func (nr *nsRegistryImpl) GetNamespaceByName(name string) Namespace {
	//TODO:
	return nil
}

//ProcessRequest process received request
func (nr *nsRegistryImpl) ProcessRequest(request interface{}) interface{} {
	//TODO:
	return nil
}
