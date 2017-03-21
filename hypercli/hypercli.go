//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package main

import (
	"github.com/urfave/cli"
	"hyperchain/hypercli/contract"
	"hyperchain/hypercli/log"
	"hyperchain/hypercli/namespace"
	"hyperchain/hypercli/server"

	"os"
	"time"
)

var app *cli.App

func initApp() {
	app = cli.NewApp()
	app.Name = "hypercli"
	app.Version = "1.3.0"
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
		{
			Name:        "server",
			Usage:       "server specific commands",
			Subcommands: server.NewServerCMD(),
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
		//{
		//	Name:  "connect",
		//	Usage: "connect to hyperchain server",
		//	Action: func(c *cli.Context) {
		//		for {
		//			reader := bufio.NewReader(os.Stdin)
		//			fmt.Print("hypercli>")
		//			input, err := reader.ReadString('\n')
		//
		//			input = strings.Trim(input, "\n")
		//			input = strings.Trim(input, " ")
		//			fmt.Println(input)
		//			args := strings.Split("hypercli "+input, " ")
		//			if err == nil {
		//				fmt.Printf("The input was: %s######", input)
		//			}
		//			os.Args = args
		//			app.Run(os.Args)
		//		}
		//	},
		//},
	}

}

func main() {
	initApp()
	app.Run(os.Args)
}
