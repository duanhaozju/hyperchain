//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package main

import (
	//"bufio"
	"fmt"
	"github.com/urfave/cli"
	"hyperchain/hypercli/contract"
	"hyperchain/hypercli/log"
	"hyperchain/hypercli/namespace"
	"hyperchain/hypercli/server"
	"os"
	//"strings"
	"time"
	//"strings"
	"bufio"
	"strings"
)

func main() {
	app := cli.NewApp()
	app.Name = "HyperCLI"
	app.Version = "1.3.0"
	app.Compiled = time.Now()
	app.Description = "Hyperchain command line client"

	app.Commands = []cli.Command{
		{
			Name:  "connect",
			Usage: "connect to hyperchain server",
			Action: func(c *cli.Context) {
				fmt.Println(os.Args)
				for {
					initCommands(app)
					reader := bufio.NewReader(os.Stdin)
					fmt.Println("Please enter some input: ")
					input, err := reader.ReadString('\n')

					input = strings.Trim(input, "\n")
					args := strings.Split("hypercli " + input, " ")
					fmt.Println(len(args))
					if err == nil {
						fmt.Printf("The input was: %s######", input)
					}
					os.Args = args
					app.Run(os.Args)
				}
			},
		},
	}
	app.Run(os.Args)
}

func initCommands(app *cli.App) {
	app.Commands = []cli.Command{
		{
			Name:        "server",
			Aliases:     []string{""},
			Usage:       "server specific commands",
			Subcommands: server.NewServerCMD(),
		},
		{
			Name:        "namespace",
			Aliases:     []string{""},
			Usage:       "namespace specific commands",
			Subcommands: namespace.NewNamespaceCMD(),
		},
		{
			Name:        "contract",
			Aliases:     []string{""},
			Usage:       "contract specific commands",
			Subcommands: contract.NewContractCMD(),
		},
		{
			Name:        "log",
			Aliases:     []string{""},
			Usage:       "log specific commands",
			Subcommands: log.NewLogCMD(),
		},
	}
}
