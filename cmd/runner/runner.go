package main

import (
	"github.com/urfave/cli"
	"hyperchain/hyperdb/db"
	"hyperchain/hyperdb/mdb"
	"github.com/syndtr/goleveldb/leveldb"
	"hyperchain/hyperdb/hleveldb"
	"hyperchain/core/vm/evm/runtime"
	"hyperchain/core/hyperstate"
	"hyperchain/common"
	"github.com/op/go-logging"
	"io/ioutil"
	"hyperchain/core/vm/evm/compiler"
	"math/big"
)

var runCommand = cli.Command{
	Action:      runCmd,
	Name:        "run",
	Usage:       "run arbitrary evm binary",
	ArgsUsage:   "<code>",
	Description: `The run command runs arbitrary EVM code.`,
}

func runCmd(ctx *cli.Context) error {
	var (
		invoke   bool        = false
		rawCode  bool        = false
		db       db.Database
		sender   string
		receiver string
		block    *big.Int
		state    *hyperstate.StateDB
		code     []byte
		input    []byte
		logger   *logging.Logger = common.GetLogger("global", "runtime")
	)
	// initialize database
	if ctx.GlobalBool(DisableExtendDBFlag.Name) {
		db, _ = mdb.NewMemDatabase()
	} else {
		lvldb, err := leveldb.OpenFile(ctx.GlobalString(DbFileFlag.Name), nil)
		if err != nil {
			return err
		}
		db = hleveldb.NewRawLDBDatabase(lvldb, "global")
	}
	// initialize code
	if ctx.GlobalString(CodeFlag.Name) != "" {
		code = common.Hex2Bytes(ctx.GlobalString(CodeFlag.Name))
	} else if ctx.GlobalString(CodeFileFlag.Name) != "" {
		source, err := ioutil.ReadFile(ctx.GlobalString(CodeFileFlag.Name))
		if err != nil {
			return err
		}
		_, bins, _, err := compiler.CompileSourcefile(string(source))
		if err != nil {
			return err
		}
		for _, bin := range bins {
			if len(code) < len(common.Hex2Bytes(bin)) {
				code = common.Hex2Bytes(bin)
			}
		}
		rawCode = true
	} else {
		invoke = true
	}
	// initialize input
	if ctx.GlobalString(InputFlag.Name) != "" {
		input = common.Hex2Bytes(ctx.GlobalString(InputFlag.Name))
	}
	// initialize sender/receiver
	if ctx.GlobalString(SenderFlag.Name) != "" {
		sender = ctx.GlobalString(SenderFlag.Name)
	}
	if ctx.GlobalString(ReceiverFlag.Name) != "" {
		receiver = ctx.GlobalString(ReceiverFlag.Name)
	}
	// initialize block
	if ctx.GlobalUint64(BlockFlag.Name) != 0 {
		block = big.NewInt(int64(ctx.GlobalUint64(BlockFlag.Name)))
	} else {
		block = big.NewInt(1)
	}
	// initialize state
	state = hyperstate.NewRaw(db, block.Uint64(), "global", runtime.InitConf())

	runtimeConfig := &runtime.Config{
		Origin:      common.HexToAddress(sender),
		Receiver:    common.HexToAddress(receiver),
		BlockNumber: block,
		State:       state,
	}


	var (
		ret        []byte
		runtimeErr error
	)

	if invoke {
		ret, runtimeErr = runtime.Call(common.HexToAddress(receiver), input, runtimeConfig)
	} else {
		if rawCode {
			code, _ ,_ = runtime.Create(db, code, runtimeConfig)
		}
		ret, _, runtimeErr = runtime.Execute(db, code, input, runtimeConfig)
	}
	logger.Notice("ret: ", common.Bytes2Hex(ret), "runtime err", runtimeErr)
	return nil
}
