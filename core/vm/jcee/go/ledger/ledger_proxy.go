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
	"strings"
)

var (
	NamespaceNotExistErr = errors.New("namespace not exist")
	InvalidRequestErr    = errors.New("invalid request permission")
)

//LedgerProxy used to manipulate data
type LedgerProxy struct {
	stateMgr *StateManager
	conf     *common.Config
	server   *grpc.Server
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
	go grpcServer.Serve(lis)
	lp.server = grpcServer
	return nil
}

func (lp *LedgerProxy) StopServer() {
	lp.server.Stop()
}

func (lp *LedgerProxy) Get(ctx context.Context, key *pb.Key) (*pb.Value, error) {
	exist, state := lp.stateMgr.GetStateDb(key.Context.Namespace)
	if exist == false {
		return nil, NamespaceNotExistErr
	}
	if valid := lp.requestCheck(key.Context); !valid {
		return nil, InvalidRequestErr
	}
	_, value := state.GetState(common.HexToAddress(key.Context.Cid), common.BytesToHash(key.K))
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
	if valid := lp.requestCheck(kv.Context); !valid {
		return nil, InvalidRequestErr
	}
	// TODO for extension leave a opcode field
	state.SetState(common.HexToAddress(kv.Context.Cid), common.BytesToHash(kv.K), kv.V, 0)
	return &pb.Response{}, nil
}

func (lp *LedgerProxy) requestCheck(ctx *pb.LedgerContext) bool {
	exist, state := lp.stateMgr.GetStateDb(ctx.Namespace)
	if exist == false {
		return false
	}
	if strings.Compare(ctx.Txid, state.GetCurrentTxHash().Hex()) == 0 {
		return true
	} else {
		return false
	}
}
