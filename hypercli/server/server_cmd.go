//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package server

import (
	"fmt"
	"github.com/op/go-logging"
	"github.com/urfave/cli"
	admin "hyperchain/api/admin"
	"hyperchain/hypercli/common"
)

var logger *logging.Logger

func init() {
	logger = logging.MustGetLogger("hypercli/server")
}

//NewServerCMD new server related commands.
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

func start(c *cli.Context) error {
	fmt.Println("Not Support Yet!")
	return nil
}

func stop(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_stopServer",
		Args:       c.Args(),
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}

func restart(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_restartServer",
		Args:       c.Args(),
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}
