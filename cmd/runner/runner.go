package main

import (
	"github.com/fatih/color"
	cm "github.com/hyperchain/hyperchain/cmd/common"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/ledger/state"
	"github.com/hyperchain/hyperchain/core/vm/evm"
	"github.com/hyperchain/hyperchain/core/vm/evm/compiler"
	"github.com/hyperchain/hyperchain/core/vm/evm/runtime"
	"github.com/hyperchain/hyperchain/hyperdb/db"
	"github.com/hyperchain/hyperchain/hyperdb/hleveldb"
	"github.com/hyperchain/hyperchain/hyperdb/mdb"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/urfave/cli"
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
		stateDb  *state.StateDB
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
	if ctx.GlobalBool(DisableExtendDBFlag.Name) {
		stateDb, _ = state.New(common.Hash{}, db, db, runtime.DefaultConf(), block.Uint64())
	} else {
		// TODO establishes a transient state correctly
	}

	runtimeConfig := &runtime.Config{
		Origin:         common.HexToAddress(sender),
		Receiver:       common.HexToAddress(receiver),
		BlockNumber:    block,
		State:          stateDb,
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
		addr       common.Address
	)

	if invoke {
		start = time.Now()
		ret, runtimeErr = runtime.Call(common.HexToAddress(receiver), input, runtimeConfig)
		elapsed = time.Since(start)
	} else {
		if rawCode {
			code, addr, _ = runtime.Create(db, code, runtimeConfig)
			runtimeConfig.Receiver = addr
		}
		start = time.Now()
		ret, _, structLogs, runtimeErr = runtime.Execute(db, code, input, runtimeConfig)
		elapsed = time.Since(start)
	}

	logs := stateDb.Logs()

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
