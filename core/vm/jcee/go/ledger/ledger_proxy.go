//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jcee

import (
	"errors"
	"golang.org/x/net/context"
	"hyperchain/common"
	"hyperchain/core/vm"
	pb "hyperchain/core/vm/jcee/protos"
	"google.golang.org/grpc"
	"net"
	"fmt"
)

var (
	NamespaceNotExistErr = errors.New("namespace not exist")
)

//LedgerProxy used to manipulate data
type LedgerProxy struct {
	stateMgr *StateManager
	conf     *common.Config
}

func NewLedgerProxy(conf *common.Config) *LedgerProxy {
	return &LedgerProxy{
		stateMgr: NewStateManager(),
		conf:     conf,
	}
}

func (lp *LedgerProxy) Register(namespace string, db vm.Database) error {
	return lp.stateMgr.Register(namespace, db)
}

func (lp *LedgerProxy) UnRegister(namespace string) error {
	return lp.stateMgr.UnReigister(namespace)
}

func (lp *LedgerProxy) Server() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", lp.conf.Get(common.C_LEDGER_PORT)))
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	pb.RegisterLedgerServer(grpcServer, lp)
	grpcServer.Serve(lis)
	return nil
}

func (lp *LedgerProxy) Get(ctx context.Context, key *pb.Key) (*pb.Value, error) {
	fmt.Println("key", key.String())
	exist, state := lp.stateMgr.GetStateDb(key.Context.Namespace)
	if exist == false {
		return nil, NamespaceNotExistErr
	}
	_, value := state.GetState(common.HexToAddress(key.Context.Cid), common.BytesToHash(key.K))
	v := &pb.Value{
		V: value,
	}
	return v, nil
}

func (lp *LedgerProxy) Put(ctx context.Context, kv *pb.KeyValue) (*pb.Response, error) {
	fmt.Println("keyvalue", kv.String())
	exist, state := lp.stateMgr.GetStateDb(kv.Context.Namespace)
	if exist == false {
		return nil, NamespaceNotExistErr
	}
	// TODO for extension leave a opcode field
	state.SetState(common.HexToAddress(kv.Context.Cid), common.BytesToHash(kv.K), kv.V, 0)
	return &pb.Response{}, nil
}
