package jcee

import (
	"errors"
	"github.com/hyperchain/hyperchain/core/vm"
	"sync"
	//"github.com/hyperchain/hyperchain/common"
)

var (
	DuplicateReigistErr = errors.New("duplicate register error")
)

type StateManager struct {
	stateDbs map[string]vm.Database
	lock     sync.RWMutex
}

func NewStateManager() *StateManager {
	stateDbs := make(map[string]vm.Database)
	return &StateManager{
		stateDbs: stateDbs,
	}
}

func (mgr *StateManager) GetStateDb(namespace string) (bool, vm.Database) {
	mgr.lock.RLock()
	defer mgr.lock.RUnlock()
	if ret, exist := mgr.stateDbs[namespace]; exist == true {
		return true, ret
	} else {
		return false, nil
	}
}

func (mgr *StateManager) Register(namespace string, stateDb vm.Database) error {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()
	if _, exist := mgr.stateDbs[namespace]; exist == true {
		return DuplicateReigistErr
	} else {
		mgr.stateDbs[namespace] = stateDb
	}
	return nil
}

func (mgr *StateManager) UnReigister(namespace string) error {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()
	delete(mgr.stateDbs, namespace)
	return nil
}
