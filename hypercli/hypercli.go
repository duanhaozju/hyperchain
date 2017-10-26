//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package main

import (
	"github.com/urfave/cli"
	"hyperchain/hypercli/contract"
	"hyperchain/hypercli/log"
	"hyperchain/hypercli/namespace"
	"hyperchain/hypercli/node"
	"hyperchain/hypercli/server"

	"hyperchain/common"
	"hyperchain/hypercli/auth"
	"os"
	"time"
)

var app *cli.App

func initApp() {
	initLogger()
	app = cli.NewApp()
	app.Name = "hypercli"
	app.Version = "1.4.0"
	app.Compiled = time.Now()
	app.Usage = "Hyperchain command line client"
	app.Description = "Run 'hypercli COMMAND --help' for more information on a command"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "host, H",
			Value: "127.0.0.1",
			Usage: "setting the host ip",
		},
		cli.StringFlag{
			Name:  "port, P",
			Value: "8081",
			Usage: "setting the host port",
		},
	}

	app.Commands = []cli.Command{
		auth.NewLoginCMD(),
		auth.NewLogoutCMD(),
		{
			Name:        "auth",
			Usage:       "auth specific commands",
			Subcommands: auth.NewAuthCMD(),
		},
		{
			Name:        "namespace",
			Usage:       "namespace specific commands",
			Subcommands: namespace.NewNamespaceCMD(),
		},
		{
			Name:        "contract",
			Usage:       "contract specific commands",
			Subcommands: contract.NewContractCMD(),
		},
		{
			Name:        "log",
			Usage:       "log specific commands",
			Subcommands: log.NewLogCMD(),
		},
		{
			Name:        "node",
			Usage:       "add/delete node specific commands",
			Subcommands: node.NewNodeCMD(),
		},
		{
			Name:        "jvm",
			Usage:       "jvm specific commands",
			Subcommands: server.NewJvmCMD(),
		},
		{
			Name:        "server",
			Usage:       "server specific commands",
			Subcommands: server.NewServerCMD(),
		},
	}

}

func initLogger() {
	conf := common.NewRawConfig()
	common.InitHyperLogger(common.DEFAULT_NAMESPACE, conf)
}

func main() {
	initApp()
	app.Run(os.Args)
}
