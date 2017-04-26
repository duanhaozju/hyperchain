package jcee

import (
	"hyperchain/core/vm"
	pb "hyperchain/core/vm/jcee/protos"
	"hyperchain/core/types"
	"github.com/golang/protobuf/proto"
	"hyperchain/common"
)

// parse parse input data and encapsulate as a invocation request.
func (cei *contractExecutorImpl) parse(ctx vm.VmContext, in []byte) *pb.Request {
	if ctx.IsCreation() {
		return &pb.Request{
			Context:  &pb.RequestContext{
				Cid:         common.HexToString(ctx.Address().Hex()),
				Namespace:   ctx.GetEnv().Namespace(),
				Txid:        ctx.GetEnv().TransactionHash().Hex(),
			},
			Method:   "deploy",
			Args:     [][]byte{[]byte(ctx.GetCodePath())},
		}

	} else {
		var args types.InvokeArgs
		if err := proto.Unmarshal(in, &args); err != nil {
			return nil
		}
		return &pb.Request{
			Context:  &pb.RequestContext{
				Cid:         common.HexToString(ctx.Address().Hex()),
				Namespace:   ctx.GetEnv().Namespace(),
				Txid:        ctx.GetEnv().TransactionHash().Hex(),
			},
			Method:   args.MethodName,
			Args:     args.Args,
		}

	}
}

