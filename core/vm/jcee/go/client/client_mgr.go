package jcee

import "sync"

var ClientMgr *ClientManager
var clientMgrOnce *sync.Once

func init() {
	clientMgrOnce = &sync.Once{}
}

type ClientManager struct {
	clients map[string]ContractExecutor
	lock    sync.RWMutex
}

func NewClientManager() *ClientManager {
	clientMgrOnce.Do(func(){
		ClientMgr = &ClientManager{
			clients: make(map[string]ContractExecutor),
		}
	})
	return ClientMgr
}

func (mgr *ClientManager) Register(namespace string) error {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()
	if _, exist := mgr.clients[namespace]; exist == true {
		return nil
	} else {
		client := NewContractExecutor()
		if err := client.Start(); err != nil {
			return err
		} else {
			mgr.clients[namespace] = client
			return nil
		}
	}
}

func (mgr *ClientManager) Get(namespace string) ContractExecutor {
	mgr.lock.RLock()
	defer mgr.lock.RUnlock()
	return mgr.clients[namespace]
}
