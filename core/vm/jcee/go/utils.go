package jvm

import (
	"github.com/gogo/protobuf/proto"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/types"
	pb "github.com/hyperchain/hyperchain/core/vm/jcee/protos"
	"os"
	"path"
	"strings"
)

const (
	ContractHome   = "hyperjvm/contracts"
	BinHome        = "hyperjvm/bin"
	ContractPrefix = "contract"
	CompressFileN  = "contract.tar.gz"
)

func getContractDir() (string, error) {
	cur, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return path.Join(cur, ContractHome), nil
}

func getBinDir() (string, error) {
	cur, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return path.Join(cur, BinHome), nil
}

func cd(target string, relative bool) (string, error) {
	cur, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if relative {
		err := os.Chdir(path.Join(cur, target))
		if err != nil {
			return "", err
		}
	} else {
		err := os.Chdir(target)
		if err != nil {
			return "", err
		}
	}
	return cur, nil
}

// parse parse input data and encapsulate as a invocation request.
func Parse(ctx *Context, in []byte) *pb.Request {
	if ctx.IsCreation() {
		var args types.InvokeArgs
		if err := proto.Unmarshal(in, &args); err != nil {
			return nil
		}
		var iArgs [][]byte
		iArgs = append(iArgs, []byte(ctx.GetCodePath()))
		iArgs = append(iArgs, args.Args...)
		return &pb.Request{
			Context: &pb.RequestContext{
				Cid:         common.HexToString(ctx.Address().Hex()),
				Namespace:   ctx.GetEnv().Namespace(),
				Txid:        ctx.GetEnv().TransactionHash().Hex(),
				BlockNumber: ctx.GetEnv().BlockNumber().Uint64(),
				Invoker:     common.HexToString(ctx.GetCaller().Address().Hex()),
			},
			Method: "deploy",
			Args:   iArgs,
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
			Context: &pb.RequestContext{
				Cid:         common.HexToString(ctx.Address().Hex()),
				Namespace:   ctx.GetEnv().Namespace(),
				Txid:        ctx.GetEnv().TransactionHash().Hex(),
				BlockNumber: ctx.GetEnv().BlockNumber().Uint64(),
			},
			Method: "update",
			Args:   iArgs,
		}
	} else {
		var args types.InvokeArgs
		if err := proto.Unmarshal(in, &args); err != nil {
			return nil
		}
		return &pb.Request{
			Context: &pb.RequestContext{
				Cid:         common.HexToString(ctx.Address().Hex()),
				Namespace:   ctx.GetEnv().Namespace(),
				Txid:        ctx.GetEnv().TransactionHash().Hex(),
				BlockNumber: ctx.GetEnv().BlockNumber().Uint64(),
			},
			Method: args.MethodName,
			Args:   args.Args,
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
