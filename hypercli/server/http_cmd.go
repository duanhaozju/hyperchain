//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package server

import (
	"github.com/urfave/cli"
	admin "hyperchain/api/jsonrpc/core"
	"hyperchain/hypercli/common"
	"fmt"
)

func NewHttpCMD() []cli.Command {
	return []cli.Command{
		{
			Name:   "start",
			Usage:  "start http server",
			Action: startHttp,
		},
		{
			Name:   "stop",
			Usage:  "stop http server",
			Action: stopHttp,
		},
		{
			Name:   "restart",
			Usage:  "restart http server",
			Action: restartHttp,
		},
	}
}

func startHttp(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_startHttpServer",
		Args:       c.Args(),
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}

func stopHttp(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_stopHttpServer",
		Args:       c.Args(),
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}

func restartHttp(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_restartHttpServer",
		Args:       c.Args(),
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}
