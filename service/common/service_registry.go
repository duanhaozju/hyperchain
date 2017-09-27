package common

import (
	"fmt"
	pb "hyperchain/service/common/protos"
	"sync"
)

type ServiceRegistry interface {
	Init() error                            // Init init the service registry.
	Register(s Service) error               // Register register new service.
	UnRegister(namespace, sid string) error // UnRegister service by service id.
	Close()                                 // Close close the service registry.
}

type NamespaceComponent struct {
	services map[string]Service //<service name, service>
	lock     sync.RWMutex

	cq MessageQueue // consenter message queue
	eq MessageQueue // executor message queue
	nq MessageQueue // network message queue
	aq MessageQueue // apiserver message queue
}

func newNamespaceComponent() *NamespaceComponent {
	size := 1000
	return &NamespaceComponent{
		services: make(map[string]Service),
		cq:       newMQImpl(size),
		eq:       newMQImpl(size),
		nq:       newMQImpl(size),
		aq:       newMQImpl(size),
	}
}

func (nc *NamespaceComponent) dispatch() {

}

func (nc *NamespaceComponent) dispatchCQ() {
	for {
		msg, err := nc.cq.Get()
		if err != nil {
			//TODO handle error
		}
		if cm, ok := msg.(pb.ConsenterMessage); ok {
			switch cm.Type {
			case pb.ConsenterMessage_InformPrimaryEvent:
				//inform primary

				m := &pb.Message{
				//Type:
				}

				nc.services[NETWORK].Send(true, m)
			case pb.ConsenterMessage_VCResetEvent:
				//vc reset
			default:
				//undefined message
			}
		} else {
			//TODO handle error
		}

	}
}

func (nc *NamespaceComponent) dispatchEQ() {

}

func (nc *NamespaceComponent) dispatchAQ() {

}

func (nc *NamespaceComponent) AddService(service Service) {
	nc.lock.Lock()
	nc.services[service.Id()] = service //TODO: add duplicate detect
	nc.lock.Unlock()
}

//Remove delete service by service id
func (nc *NamespaceComponent) Remove(sid string) {
	nc.lock.Lock()
	delete(nc.services, sid) //TODO: add existence detect
	nc.lock.Unlock()
}

func (nc *NamespaceComponent) Service(sid string) Service {
	nc.lock.RLock()
	defer nc.lock.RUnlock()
	return nc.services[sid] //TODO: add existence detect
}

//Contains NamespaceComponent whether contains service with sid.
func (nc *NamespaceComponent) Contains(sid string) bool {
	nc.lock.RLock()
	defer nc.lock.RUnlock()
	_, ok := nc.services[sid]
	return ok
}

func (nc *NamespaceComponent) Close() {
	for _, s := range nc.services {
		s.Close()
	}
}

type serviceRegistryImpl struct {
	lock       sync.RWMutex
	components map[string]*NamespaceComponent // <namespace, component>
}

// Init init the service registry.
func (sri *serviceRegistryImpl) Init() error {
	// TODO: add more init details
	return nil
}

func (sri *serviceRegistryImpl) AddNamespaceComponent(namespace string) {
	sri.lock.Lock()
	defer sri.lock.Unlock()
	sri.components[namespace] = newNamespaceComponent()
}

// Register register new service.
func (sri *serviceRegistryImpl) Register(s Service) error {
	sri.lock.Lock()
	if _, ok := sri.components[s.Namespace()]; !ok {
		sri.AddNamespaceComponent(s.Namespace())
	}
	sri.components[s.Namespace()].AddService(s)
	sri.lock.Unlock()
	return nil
}

// UnRegister service by service id.
func (sri *serviceRegistryImpl) UnRegister(namespace, sid string) error {
	sri.lock.Lock()
	//delete(sri.components, sid)
	if !sri.ContainsNamespace(namespace) {
		return fmt.Errorf("UnRegister error: namespace[%s] is not found!", namespace)
	} else {
		c := sri.components[namespace]
		if !c.Contains(sid) {
			return fmt.Errorf("UnRegister error: service id[%s] is not found", sid)
		}
		c.Service(sid).Close()
		c.Remove(sid)
	}
	sri.lock.Unlock()
	return nil
}

// Close close the service registry.
func (sri *serviceRegistryImpl) Close() {
	for _, c := range sri.components {
		c.Close()
	}
}

func (sri *serviceRegistryImpl) ContainsNamespace(namespace string) bool {
	sri.lock.RLock()
	defer sri.lock.RUnlock()
	_, ok := sri.components[namespace]
	return ok
}
