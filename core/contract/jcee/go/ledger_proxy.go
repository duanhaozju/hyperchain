//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jcee

import (
	context "golang.org/x/net/context"
	"fmt"
	pb "hyperchain/core/contract/jcee/protos"
)

//LedgerProxy used to manipulate data
type LedgerProxy struct {
}

func (lp *LedgerProxy) Get(ctx context.Context, key *pb.Key) (*pb.Value, error) {
	fmt.Println(key)
	v := &pb.Value{}
	v.Id = key.Id
	v.V = []byte("xiaoyi")
	return v, nil
}

func (lp *LedgerProxy) Put(ctx context.Context, kv *pb.KeyValue) (*pb.Response, error) {

	//Todo fetch data from ledger
	return nil, nil
}

func NewLedgerProxy() *LedgerProxy {
	return &LedgerProxy{}
}
