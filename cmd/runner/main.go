package main

import (
	"fmt"
	"github.com/hyperchain/hyperchain/cmd/common"
	cm "github.com/hyperchain/hyperchain/common"
	"github.com/urfave/cli"
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
		Usage: "The transaction recipient",
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
	DebugFlag = cli.BoolFlag{
		Name:  "debug",
		Usage: "output all debug trace info",
	}
	CodeFlag = cli.StringFlag{
		Name:  "code",
		Usage: "EVM code",
	}
	CodeFileFlag = cli.StringFlag{
		Name:  "codefile",
		Usage: "file containing EVM code",
	}
	InputFlag = cli.StringFlag{
		Name:  "input",
		Usage: "invoke input",
	}
	StatFlag = cli.BoolFlag{
		Name:  "stat",
		Usage: "output all stat info",
	}
	DisableExtendDBFlag = cli.BoolFlag{
		Name:  "nodb",
		Usage: "disable db, use empty memory db as default",
	}
	DbFileFlag = cli.StringFlag{
		Name:  "dbfile",
		Usage: "database path",
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
	cm.InitHyperLogger(cm.DEFAULT_NAMESPACE, conf)
	cm.GetLogger(cm.DEFAULT_NAMESPACE, "state")
	cm.SetLogLevel(cm.DEFAULT_NAMESPACE, "state", "NOTICE")
	cm.GetLogger(cm.DEFAULT_NAMESPACE, "buckettree")
	cm.SetLogLevel(cm.DEFAULT_NAMESPACE, "buckettree", "NOTICE")
}
