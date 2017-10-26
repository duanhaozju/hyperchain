//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package server

import (
	"fmt"

	"hyperchain/api/admin"
	"hyperchain/hypercli/common"

	"github.com/urfave/cli"
)

// NewServerCMD new server related commands.
func NewServerCMD() []cli.Command {
	return []cli.Command{
		{
			Name:   "start",
			Usage:  "start hyperchain server",
			Action: start,
		},
		{
			Name:   "stop",
			Usage:  "stop hyperchain server",
			Action: stop,
		},
		{
			Name:   "restart",
			Usage:  "restart hyperchain server",
			Action: restart,
		},
	}
}

// Not support yet!
func start(c *cli.Context) error {
	fmt.Println("Not Support Yet!")
	return nil
}

// stop helps stop hyperchain service.
func stop(c *cli.Context) error {
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))
	cmd := &admin.Command{
		MethodName: "admin_stopServer",
		Args:       c.Args(),
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}

// restart helps restart hyperchain service.
func restart(c *cli.Context) error {
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))
	cmd := &admin.Command{
		MethodName: "admin_restartServer",
		Args:       c.Args(),
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}
