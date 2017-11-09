package service

import (
	"fmt"
	"sync"
)

type ServiceRegistry interface {
	// Init init the service registry.
	Init() error

	// Register register new service.
	Register(s Service) error

	// UnRegister service by service id.
	UnRegister(namespace, sid string) error

	// Close close the service registry.
	Close()

	// ContainsNamespace whether contains the namespace service.
	ContainsNamespace(name string) bool

	// Namespace fetch namespace services.
	Namespace(name string) *NamespaceServices

	// AdminService fetch admin service.
	AdminService(aid string) Service

	// AddAdminService add admin service, if adminSrv exists means adminSrv is restart and return true.
	AddAdminService(adminSrv Service) bool
}

func NewServiceRegistry() ServiceRegistry {
	return &serviceRegistryImpl{
		namespaces: make(map[string]*NamespaceServices),
		admins:     make(map[string]Service),
	}
}

type NamespaceServices struct {
	services map[string]Service //<service name, service>
	adlock   sync.RWMutex
	adminSrv Service
	lock     sync.RWMutex
}

func newNamespace() *NamespaceServices {
	return &NamespaceServices{
		services: make(map[string]Service),
	}
}

func (nc *NamespaceServices) AddAdminSrv(service Service) {
	nc.adlock.Lock()
	nc.adminSrv = service
	nc.adlock.Unlock()
}

func (nc *NamespaceServices) AdminService() Service {
	nc.adlock.RLock()
	defer nc.adlock.RUnlock()
	return nc.adminSrv
}

func (nc *NamespaceServices) AddService(service Service) {
	nc.lock.Lock()
	nc.services[service.Id()] = service //TODO: add duplicate detect
	nc.lock.Unlock()
}

//Remove delete service by service id
func (nc *NamespaceServices) Remove(sid string) {
	nc.lock.Lock()
	delete(nc.services, sid) //TODO: add existence detect
	nc.lock.Unlock()
}

func (nc *NamespaceServices) Service(sid string) Service {
	nc.lock.RLock()
	defer nc.lock.RUnlock()
	return nc.services[sid] //TODO: add existence detect
}

//Contains NamespaceServices whether contains service with sid.
func (nc *NamespaceServices) Contains(sid string) bool {
	nc.lock.RLock()
	defer nc.lock.RUnlock()
	_, ok := nc.services[sid]
	return ok
}

func (nc *NamespaceServices) Close() {
	for _, s := range nc.services {
		s.Close()
	}
}

type serviceRegistryImpl struct {
	lock       sync.RWMutex
	namespaces map[string]*NamespaceServices // <namespace, component>

	adlock sync.RWMutex
	admins map[string]Service
}

func (sri *serviceRegistryImpl) AdminService(aid string) Service {
	sri.adlock.RLock()
	defer sri.adlock.RUnlock()
	return sri.admins[aid]
}

func (sri *serviceRegistryImpl) AddAdminService(adminSrv Service) bool {
	adminSrvRestart := false
	sri.adlock.Lock()
	defer sri.adlock.Unlock()
	if _, ok := sri.admins[adminSrv.Id()]; ok {
		adminSrvRestart = true
	}
	sri.admins[adminSrv.Id()] = adminSrv
	return adminSrvRestart
}

// Init init the service registry.
func (sri *serviceRegistryImpl) Init() error {
	// TODO: add more init details
	return nil
}

func (sri *serviceRegistryImpl) AddNamespace(namespace string) {
	sri.lock.Lock()
	defer sri.lock.Unlock()
	sri.namespaces[namespace] = newNamespace()
}

// Register register new service.
func (sri *serviceRegistryImpl) Register(s Service) error {
	sri.lock.Lock()
	defer sri.lock.Unlock()
	if _, ok := sri.namespaces[s.Namespace()]; !ok {
		sri.namespaces[s.Namespace()] = newNamespace()
	}
	sri.namespaces[s.Namespace()].AddService(s)
	return nil
}

// UnRegister service by service id.
func (sri *serviceRegistryImpl) UnRegister(namespace, sid string) error {
	sri.lock.Lock()
	defer sri.lock.Unlock()
	//delete(sri.namespaces, sid)
	if !sri.ContainsNamespace(namespace) {
		return fmt.Errorf("deregister error: namespace[%s] is not found! ", namespace)
	} else {
		c := sri.namespaces[namespace]
		if !c.Contains(sid) {
			return fmt.Errorf("UnRegister error: service id[%s] is not found", sid)
		}
		c.Service(sid).Close()
		c.Remove(sid)
	}

	return nil
}

// Close close the service registry.
func (sri *serviceRegistryImpl) Close() {
	for _, c := range sri.namespaces {
		c.Close()
	}
}

func (sri *serviceRegistryImpl) ContainsNamespace(namespace string) bool {
	sri.lock.RLock()
	defer sri.lock.RUnlock()
	_, ok := sri.namespaces[namespace]
	return ok
}

func (sri *serviceRegistryImpl) Namespace(name string) *NamespaceServices {
	sri.lock.RLock()
	defer sri.lock.RUnlock()
	return sri.namespaces[name]
}
