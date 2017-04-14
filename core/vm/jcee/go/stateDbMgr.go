package jcee

import "hyperchain/core/vm"

type StateDbManager struct {
	stateDbs    map[string]vm.Database
}

func NewStateDbManager() *StateDbManager {
	stateDbs := make(map[string]vm.Database)
}
