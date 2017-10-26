//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package server

import (
	"fmt"

	"hyperchain/api/admin"
	"hyperchain/hypercli/common"

	"github.com/urfave/cli"
)

// NewJvmCMD new jvm related command.
func NewJvmCMD() []cli.Command {
	return []cli.Command{
		{
			Name:   "start",
			Usage:  "start jvm server",
			Action: startJvm,
		},
		{
			Name:   "stop",
			Usage:  "stop jvm server",
			Action: stopJvm,
		},
		{
			Name:   "restart",
			Usage:  "restart jvm server",
			Action: restartJvm,
		},
	}
}

// startJvm helps start jvm server.
func startJvm(c *cli.Context) error {
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))

	cmd := &admin.Command{
		MethodName: "admin_startJvmServer",
		Args:       c.Args(),
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}

// stopJvm helps stop jvm server.
func stopJvm(c *cli.Context) error {
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))

	cmd := &admin.Command{
		MethodName: "admin_stopJvmServer",
		Args:       c.Args(),
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}

// restartJvm helps restart jvm server.
func restartJvm(c *cli.Context) error {
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))

	cmd := &admin.Command{
		MethodName: "admin_restartJvmServer",
		Args:       c.Args(),
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}
