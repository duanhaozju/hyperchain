//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jcee

import (
	"errors"
	"golang.org/x/net/context"
	"hyperchain/common"
	"hyperchain/core/vm"
	pb "hyperchain/core/vm/jcee/protos"
)

var (
	NamespaceNotExistErr = errors.New("namespace not exist")
)

//LedgerProxy used to manipulate data
type LedgerProxy struct {
	stateMgr *StateManager
}

func NewLedgerProxy() *LedgerProxy {
	return &LedgerProxy{
		stateMgr: NewStateManager(),
	}
}

func (lp *LedgerProxy) Register(namespace string, db vm.Database) error {
	return lp.stateMgr.Register(namespace, db)
}

func (lp *LedgerProxy) UnRegister(namespace string) error {
	return lp.stateMgr.UnReigister(namespace)
}

func (lp *LedgerProxy) Get(ctx context.Context, key *pb.Key) (*pb.Value, error) {
	exist, state := lp.stateMgr.GetStateDb(key.Context.Namespace)
	if exist == false {
		return nil, NamespaceNotExistErr
	}
	_, value := state.GetState(common.Address{}, common.BytesToHash(key.K))
	v := &pb.Value{
		V: value,
	}
	return v, nil
}

func (lp *LedgerProxy) Put(ctx context.Context, kv *pb.KeyValue) (*pb.Response, error) {
	exist, state := lp.stateMgr.GetStateDb(kv.Context.Namespace)
	if exist == false {
		return nil, NamespaceNotExistErr
	}
	// TODO for extension leave a opcode field
	state.SetState(common.Address{}, common.BytesToHash(kv.K), kv.V, 0)
	return nil, nil
}
