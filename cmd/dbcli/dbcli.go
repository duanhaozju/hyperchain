package main

import (
	"github.com/urfave/cli"
	"hyperchain/cmd/dbcli/block"
	"hyperchain/cmd/dbcli/chain"
	"hyperchain/cmd/dbcli/receipt"
	"hyperchain/cmd/dbcli/transaction"
	"os"
	"time"
)

var app *cli.App

func initApp() {
	app = cli.NewApp()
	app.Name = "dbcli"
	app.Version = "1.0"
	app.Compiled = time.Now()
	app.Usage = "DB command line client"
	app.Description = "Run 'dbcli COMMAND --help' for more information on a command"
	app.Commands = []cli.Command{
		{
			Name:        "block",
			Usage:       "block specific commands",
			Subcommands: block.NewBlockCMD(),
		},
		{
			Name:        "transaction",
			Usage:       "transaction specific commands",
			Subcommands: transaction.NewTransactionCMD(),
		},
		{
			Name:        "chain",
			Usage:       "chain specific commands",
			Subcommands: chain.NewChainCMD(),
		},
		{
			Name:        "receipt",
			Usage:       "receipt specific commands",
			Subcommands: receipt.NewReceiptCMD(),
		},
	}

}

func main() {
	initApp()
	app.Run(os.Args)
}
