package jvm

import "sync"

/*
	Deprecated
*/

import (
	"github.com/hyperchain/hyperchain/common"
)

var ClientMgr *ClientManager
var clientMgrOnce *sync.Once

func init() {
	clientMgrOnce = &sync.Once{}
}

type ClientManager struct {
	clients map[string]ContractExecutor
	lock    sync.RWMutex
	conf    *common.Config
}

func NewClientManager(conf *common.Config) *ClientManager {
	clientMgrOnce.Do(func() {
		ClientMgr = &ClientManager{
			clients: make(map[string]ContractExecutor),
			conf:    conf,
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
		client := NewContractExecutor(mgr.conf, namespace)
		if err := client.Start(); err != nil {
			return err
		} else {
			mgr.clients[namespace] = client
			return nil
		}
	}
}

func (mgr *ClientManager) UnRegister(namespace string) error {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()
	delete(mgr.clients, namespace)
	return nil
}

func (mgr *ClientManager) Get(namespace string) ContractExecutor {
	mgr.lock.RLock()
	defer mgr.lock.RUnlock()
	return mgr.clients[namespace]
}
