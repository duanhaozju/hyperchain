package main

import (
	"github.com/fatih/color"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/urfave/cli"
	cm "hyperchain/cmd/common"
	"hyperchain/common"
	"hyperchain/core/state"
	"hyperchain/core/vm/evm"
	"hyperchain/core/vm/evm/compiler"
	"hyperchain/core/vm/evm/runtime"
	"hyperchain/hyperdb/db"
	"hyperchain/hyperdb/hleveldb"
	"hyperchain/hyperdb/mdb"
	"io/ioutil"
	"math/big"
	"os"
	rt "runtime"
	"time"
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
		invoke   bool = false
		rawCode  bool = false
		db       db.Database
		sender   string
		receiver string
		block    *big.Int
		state    *state.StateDB
		code     []byte
		input    []byte
		log      *cm.CWriter = &cm.CWriter{os.Stdout}
	)
	// initialize database
	if ctx.GlobalBool(DisableExtendDBFlag.Name) {
		db, _ = mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	} else {
		lvldb, err := leveldb.OpenFile(ctx.GlobalString(DbFileFlag.Name), nil)
		if err != nil {
			return err
		}
		db = hleveldb.NewRawLDBDatabase(lvldb, common.DEFAULT_NAMESPACE)
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
	state = state.NewRaw(db, block.Uint64(), common.DEFAULT_NAMESPACE, runtime.InitConf())

	runtimeConfig := &runtime.Config{
		Origin:         common.HexToAddress(sender),
		Receiver:       common.HexToAddress(receiver),
		BlockNumber:    block,
		State:          state,
		Debug:          ctx.GlobalBool(DebugFlag.Name),
		DisableMemory:  ctx.GlobalBool(DisableMemoryFlag.Name),
		DisableStack:   ctx.GlobalBool(DisableStackFlag.Name),
		DisableStorage: ctx.GlobalBool(DisableStorageFlag.Name),
	}

	var (
		ret        []byte
		runtimeErr error
		structLogs []evm.StructLog
		start      time.Time
		elapsed    time.Duration
	)

	if invoke {
		start = time.Now()
		ret, runtimeErr = runtime.Call(common.HexToAddress(receiver), input, runtimeConfig)
		elapsed = time.Since(start)
	} else {
		if rawCode {
			code, _, _ = runtime.Create(db, code, runtimeConfig)
		}
		start = time.Now()
		ret, _, structLogs, runtimeErr = runtime.Execute(db, code, input, runtimeConfig)
		elapsed = time.Since(start)
	}

	logs := state.Logs()

	log.WriteF(color.FgHiGreen, "[Result]:        %v\n", common.Bytes2Hex(ret))
	log.WriteF(color.FgHiRed, "[Error]:         %v\n", runtimeErr)
	log.WriteF(color.FgHiBlue, "[Logs]:          %v\n", logs)
	if ctx.GlobalBool(DebugFlag.Name) {
		evm.StdErrFormat(structLogs)
	}
	if ctx.GlobalBool(StatFlag.Name) {
		var mem rt.MemStats
		rt.ReadMemStats(&mem)
		log.WriteF(color.FgHiMagenta, `[Elapsed]:       %v
[heap object]:   %v
[allocation]:    %v
[GC calls]:      %v

`, elapsed, mem.HeapObjects, mem.TotalAlloc, mem.NumGC)
	}
	return nil
}
