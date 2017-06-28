//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jcee

import (
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"hyperchain/common"
	"hyperchain/core/vm"
	pb "hyperchain/core/vm/jcee/protos"
	"net"
	"strings"
	"hyperchain/hyperdb/db"
	"hyperchain/core/types"
)

var (
	NamespaceNotExistErr = errors.New("namespace not exist")
	InvalidRequestErr    = errors.New("invalid request permission")
)

const  (
	BatchSize = 100
)

//LedgerProxy used to manipulate data
type LedgerProxy struct {
	stateMgr  *StateManager
	conf      *common.Config
	server    *grpc.Server
	iterStack map[string]*Iterator
}

type Iterator struct {
	dbIter     db.Iterator
	stateIdx   int
}

func NewLedgerProxy(conf *common.Config) *LedgerProxy {
	return &LedgerProxy{
		stateMgr:   NewStateManager(),
		conf:       conf,
		iterStack:  make(map[string]*Iterator),
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
		Id: key.Context.Txid,
		V: value,
	}
	return v, nil
}

func (lp *LedgerProxy) Put(ctx context.Context, kv *pb.KeyValue) (*pb.Response, error) {
	exist, state := lp.stateMgr.GetStateDb(kv.Context.Namespace)
	if exist == false {
		return &pb.Response{Ok: false}, NamespaceNotExistErr
	}
	if valid := lp.requestCheck(kv.Context); !valid {
		return &pb.Response{Ok: false}, InvalidRequestErr
	}
	state.SetState(common.HexToAddress(kv.Context.Cid), common.BytesToHash(kv.K), kv.V, 0)
	return &pb.Response{Ok: true}, nil
}

func (lp *LedgerProxy) BatchRead(ctx context.Context, batch *pb.BatchKey) (*pb.BathValue, error) {
	exist, state := lp.stateMgr.GetStateDb(batch.Context.Namespace)
	if exist == false {
		return nil, NamespaceNotExistErr
	}
	if valid := lp.requestCheck(batch.Context); !valid {
		return nil, InvalidRequestErr
	}
	response := &pb.BathValue{}
	for _, key := range batch.K {
		exist, value := state.GetState(common.HexToAddress(batch.Context.Cid), common.BytesToHash(key))
		if exist == true {
			response.V = append(response.V, value)
		} else {
			response.V = append(response.V, nil)
		}
	}
	response.HasMore = false
	response.Id = batch.Context.Txid
	return response, nil
}
func (lp *LedgerProxy) BatchWrite(ctx context.Context, batch *pb.BatchKV) (*pb.Response, error) {
	exist, state := lp.stateMgr.GetStateDb(batch.Context.Namespace)
	if exist == false {
		return &pb.Response{Ok: false}, NamespaceNotExistErr
	}
	if valid := lp.requestCheck(batch.Context); !valid {
		return &pb.Response{Ok: false}, InvalidRequestErr
	}
	for _, kv := range batch.Kv {
		state.SetState(common.HexToAddress(batch.Context.Cid), common.BytesToHash(kv.K), kv.V, 0)
	}
	return &pb.Response{Ok: true}, nil
}

func (lp *LedgerProxy) RangeQuery(r *pb.Range, stream pb.Ledger_RangeQueryServer) error  {
	exist, state := lp.stateMgr.GetStateDb(r.Context.Namespace)
	if exist == false {
		return NamespaceNotExistErr
	}
	if valid := lp.requestCheck(r.Context); !valid {
		return InvalidRequestErr
	}
	start := common.BytesToHash(r.Start)
	limit := common.BytesToHash(r.End)
	iterRange := vm.IterRange{
		Start:   &start,
		Limit:   &limit,
	}
	iter, err := state.NewIterator(common.BytesToAddress(common.FromHex(r.Context.Cid)), &iterRange)
	if err != nil {
		return err
	}
	cnt := 0
	batchValue := pb.BathValue{
		Id:  r.Context.Txid,
	}
	for iter.Next() {
		batchValue.V = append(batchValue.V, iter.Value())
		cnt += 1
		if cnt == BatchSize {
			batchValue.HasMore = true
			if err := stream.Send(&batchValue); err != nil {
				return err
			}
			cnt = 0
			batchValue = pb.BathValue{
				Id:  r.Context.Txid,
			}
		}
	}
	batchValue.HasMore = false
	if err := stream.Send(&batchValue); err != nil {
		return err
	}
	return nil
}

func (lp *LedgerProxy) Delete(ctx context.Context, in *pb.Key) (*pb.Response, error) {
	exist, state := lp.stateMgr.GetStateDb(in.Context.Namespace)
	if exist == false {
		return &pb.Response{Ok: false}, NamespaceNotExistErr
	}
	if valid := lp.requestCheck(in.Context); !valid {
		return &pb.Response{Ok: false}, InvalidRequestErr
	}
	state.SetState(common.HexToAddress(in.Context.Cid), common.BytesToHash(in.K), nil, 0)
	return &pb.Response{Ok: true}, nil
}

func (lp *LedgerProxy) Post(ctx context.Context, event *pb.Event) (*pb.Response, error) {

	exist, state := lp.stateMgr.GetStateDb(event.Context.Namespace)
	if exist == false {
		return &pb.Response{Ok: false}, NamespaceNotExistErr
	}
	if valid := lp.requestCheck(event.Context); !valid {
		return &pb.Response{Ok: false}, InvalidRequestErr
	}

	var topics []common.Hash
	for _, topic := range event.Topics {
		topics = append(topics, common.BytesToHash(topic))
	}

	log := types.NewLog(common.HexToAddress(event.Context.Cid), topics, event.Body, event.Context.BlockNumber, types.LogVmType_JVM)
	state.AddLog(log)

	return &pb.Response{
		Ok:true,
	}, nil
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
