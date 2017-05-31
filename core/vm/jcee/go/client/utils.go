package jcee

import (
	"hyperchain/core/vm"
	pb "hyperchain/core/vm/jcee/protos"
	"hyperchain/core/types"
	"github.com/golang/protobuf/proto"
	"hyperchain/common"
	"strings"
)

// parse parse input data and encapsulate as a invocation request.
func (cei *contractExecutorImpl) parse(ctx vm.VmContext, in []byte) *pb.Request {
	if ctx.IsCreation() {
		var args types.InvokeArgs
		if err := proto.Unmarshal(in, &args); err != nil {
			return nil
		}
		var iArgs [][]byte
		iArgs = append(iArgs, []byte(ctx.GetCodePath()))
		iArgs = append(iArgs, args.Args...)
		return &pb.Request{
			Context:  &pb.RequestContext{
				Cid:         common.HexToString(ctx.Address().Hex()),
				Namespace:   ctx.GetEnv().Namespace(),
				Txid:        ctx.GetEnv().TransactionHash().Hex(),
			},
			Method:   "deploy",
			Args:     iArgs,
		}
	} else if ctx.IsUpdate() {
		var args types.InvokeArgs
		if err := proto.Unmarshal(in, &args); err != nil {
			return nil
		}
		var iArgs [][]byte
		iArgs = append(iArgs, []byte(ctx.GetCodePath()))
		iArgs = append(iArgs, args.Args...)
		return &pb.Request{
			Context:  &pb.RequestContext{
				Cid:         common.HexToString(ctx.Address().Hex()),
				Namespace:   ctx.GetEnv().Namespace(),
				Txid:        ctx.GetEnv().TransactionHash().Hex(),
			},
			Method:   "update",
			Args:     iArgs,
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

func hexMatch(str1, str2 string) bool {
	if strings.HasPrefix(str1, "0x") {
		str1 = str1[2:]
	}
	if strings.HasPrefix(str2, "0x") {
		str2 = str2[2:]
	}
	return str1 == str2
}

