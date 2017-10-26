package main

import (
	"github.com/hyperchain/hyperchain/cmd/dbcli/account"
	"github.com/hyperchain/hyperchain/cmd/dbcli/block"
	"github.com/hyperchain/hyperchain/cmd/dbcli/chain"
	"github.com/hyperchain/hyperchain/cmd/dbcli/receipt"
	"github.com/hyperchain/hyperchain/cmd/dbcli/transaction"
	"github.com/urfave/cli"
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
		{
			Name:        "account",
			Usage:       "account specific commands",
			Subcommands: account.NewAccountCMD(),
		},
	}

}

func main() {
	initApp()
	app.Run(os.Args)
}
