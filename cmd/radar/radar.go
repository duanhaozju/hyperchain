package main

import (
	"github.com/hyperchain/hyperchain/cmd/radar/contract"
	"github.com/urfave/cli"
	"os"
	"time"
)

var app *cli.App

func initApp() {
	app = cli.NewApp()
	app.Name = "radar"
	app.Version = "1.0"
	app.Compiled = time.Now()
	app.Usage = "analyse data using contract source code and data stored in leveldb."
	app.Description = "Run 'radar COMMAND --help' for more information on a command"
	app.Commands = []cli.Command{
		{
			Name:        "contract",
			Usage:       "contract specific commands",
			Subcommands: contract.NewContractCMD(),
		},
	}

}

func main() {
	initApp()
	app.Run(os.Args)
}
