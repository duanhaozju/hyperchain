package common

import (
	"fmt"
	"sync"
)

type ServiceRegistry interface {
	Init() error                            // Init init the service registry.
	Register(s Service) error               // Register register new service.
	UnRegister(namespace, sid string) error // UnRegister service by service id.
	Close()                                 // Close close the service registry.
	ContainsNamespace(name string) bool
	Namespace(name string) *Namespace
}

func NewServiceRegistry() ServiceRegistry {
	return &serviceRegistryImpl{
		namespaces: make(map[string]*Namespace),
	}
}

type Namespace struct {
	services map[string]Service //<service name, service>
	lock     sync.RWMutex

	cq MessageQueue // consenter message queue
	eq MessageQueue // executor message queue
	nq MessageQueue // network message queue
	aq MessageQueue // apiserver message queue
}

func newNamespace() *Namespace {
	size := 1000
	return &Namespace{
		services: make(map[string]Service),
		cq:       newMQImpl(size),
		eq:       newMQImpl(size),
		nq:       newMQImpl(size),
		aq:       newMQImpl(size),
	}
}

func (nc *Namespace) dispatch() {

}

func (nc *Namespace) dispatchCQ() {
	//for {
	//	//msg, err := nc.cq.Get()
	//	//if err != nil {
	//	//	//TODO handle error
	//	//}
	//	//if cm, ok := msg.(pb.ConsenterMessage); ok {
	//	//	switch cm.Type {
	//	//	case pb.ConsenterMessage_InformPrimaryEvent:
	//	//		//inform primary
	//	//
	//	//		m := &pb.Message{
	//	//		}
	//	//
	//	//
	//	//
	//	//		nc.services[NETWORK].Send(true, m)
	//	//	case pb.ConsenterMessage_VCResetEvent:
	//	//		//vc reset
	//	//	default:
	//	//		//undefined message
	//	//	}
	//	//} else {
	//	//	//TODO handle error
	//	//}
	//
	//}
}

func (nc *Namespace) dispatchEQ() {

}

func (nc *Namespace) dispatchAQ() {

}

func (nc *Namespace) AddService(service Service) {
	nc.lock.Lock()
	nc.services[service.Id()] = service //TODO: add duplicate detect
	nc.lock.Unlock()
}

//Remove delete service by service id
func (nc *Namespace) Remove(sid string) {
	nc.lock.Lock()
	delete(nc.services, sid) //TODO: add existence detect
	nc.lock.Unlock()
}

func (nc *Namespace) Service(sid string) Service {
	nc.lock.RLock()
	defer nc.lock.RUnlock()
	return nc.services[sid] //TODO: add existence detect
}

//Contains Namespace whether contains service with sid.
func (nc *Namespace) Contains(sid string) bool {
	nc.lock.RLock()
	defer nc.lock.RUnlock()
	_, ok := nc.services[sid]
	return ok
}

func (nc *Namespace) Close() {
	for _, s := range nc.services {
		s.Close()
	}
}

type serviceRegistryImpl struct {
	lock       sync.RWMutex
	namespaces map[string]*Namespace // <namespace, component>
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
		return fmt.Errorf("UnRegister error: namespace[%s] is not found!", namespace)
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

func (sri *serviceRegistryImpl) Namespace(name string) *Namespace {
	sri.lock.RLock()
	defer sri.lock.RUnlock()
	return sri.namespaces[name]
}
