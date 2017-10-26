//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package server

import (
	"fmt"
	"github.com/urfave/cli"
	admin "hyperchain/api/admin"
	"hyperchain/hypercli/common"
)

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

func startJvm(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_startJvmServer",
		Args:       c.Args(),
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}

func stopJvm(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_stopJvmServer",
		Args:       c.Args(),
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}

func restartJvm(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_restartJvmServer",
		Args:       c.Args(),
	}
	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}
