//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package server

import (
	"fmt"
	"github.com/op/go-logging"
	"github.com/urfave/cli"
	"hyperchain/api/admin"
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
			Name:    "start",
			Aliases: []string{"-s"},
			Usage:   "start hyperchain server",
			Action:  start,
		},
		{
			Name:    "stop",
			Aliases: []string{"-sp"},
			Usage:   "stop hyperchain server",
			Action:  stop,
		},
		{
			Name:    "restart",
			Aliases: []string{"-r"},
			Usage:   "restart hyperchain server",
			Action:  restart,
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
	rs := client.InvokeCmd(cmd)
	fmt.Println(rs)
	return nil
}

func restart(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_restartServer",
		Args:       c.Args(),
	}
	rs := client.InvokeCmd(cmd)
	fmt.Println(rs)
	return nil
}
