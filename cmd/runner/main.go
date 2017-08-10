package main

import (
	"fmt"
	"github.com/urfave/cli"
	"hyperchain/cmd/common"
	cm "hyperchain/common"
	"os"
)

var gitcommit = ""

var (
	app = common.NewApp(gitcommit, "vm helper")

	SenderFlag = cli.StringFlag{
		Name:  "sender",
		Usage: "The transaction origin",
	}
	ReceiverFlag = cli.StringFlag{
		Name:  "receiver",
		Usage: "The transaction origin",
	}
	BlockFlag = cli.Uint64Flag{
		Name:  "block",
		Usage: "based block number",
	}
	DisableMemoryFlag = cli.BoolFlag{
		Name:  "nomemory",
		Usage: "disable memory output",
	}
	DisableStackFlag = cli.BoolFlag{
		Name:  "nostack",
		Usage: "disable stack output",
	}
	DisableStorageFlag = cli.BoolFlag{
		Name:  "nostorage",
		Usage: "disable storage output",
	}
	DisableGasMeteringFlag = cli.BoolFlag{
		Name:  "nogasmetering",
		Usage: "disable gas metering",
	}
	DebugFlag = cli.BoolFlag{
		Name:  "debug",
		Usage: "output all debug trace info",
	}
	CodeFlag = cli.StringFlag{
		Name:  "code",
		Usage: "EVM code",
	}
	StatFlag = cli.BoolFlag{
		Name:  "stat",
		Usage: "output all stat info",
	}
	CodeFileFlag = cli.StringFlag{
		Name:  "codefile",
		Usage: "file containing EVM code",
	}
	InputFlag = cli.StringFlag{
		Name:  "input",
		Usage: "invoke input",
	}
	DisableExtendDBFlag = cli.BoolFlag{
		Name:  "nodb",
		Usage: "disable db, use empty memory db as default",
	}
	DbFileFlag = cli.StringFlag{
		Name:  "dbfile",
		Usage: "database path",
	}
	MemProfileFlag = cli.StringFlag{
		Name:  "memprofile",
		Usage: "creates a memory profile at the given path",
	}
	CPUProfileFlag = cli.StringFlag{
		Name:  "cpuprofile",
		Usage: "creates a CPU profile at the given path",
	}
)

func init() {
	app.Flags = []cli.Flag{
		SenderFlag,
		ReceiverFlag,
		BlockFlag,
		DisableMemoryFlag,
		DisableStackFlag,
		DisableStorageFlag,
		DebugFlag,
		StatFlag,
		CodeFlag,
		CodeFileFlag,
		InputFlag,
		DbFileFlag,
		DisableExtendDBFlag,
	}
	app.Commands = []cli.Command{
		runCommand,
	}
	initLog()
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initLog() {
	conf := cm.NewRawConfig()
	conf.Set(cm.LOG_DUMP_FILE, false)
	conf.Set(cm.LOG_BASE_LOG_LEVEL, "NOTICE")
	conf.Set(cm.LOG_FILE_FORMAT, "[%{module}][%{level:.5s}] %{time:15:04:05.000} %{shortfile} %{message}")
	conf.Set(cm.LOG_CONSOLE_FORMAT, "%{color}[%{module}][%{level:.5s}] %{time:15:04:05.000} %{shortfile} %{message} %{color:reset}")
	conf.Set(cm.LOG_DUMP_FILE_DIR, "")
	conf.Set(cm.NAMESPACE, "global")
	cm.InitHyperLoggerManager(conf)
	cm.InitHyperLogger(conf, "global")
}
