//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package server

import (
	"github.com/urfave/cli"
	admin "hyperchain/api/jsonrpc/core"
	"hyperchain/hypercli/common"
	"fmt"
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
	if result.Ok != true {
		fmt.Printf("Failed to start jvm manager: %v", result.Result)
	}
	fmt.Print(result.Result)
	return nil
}

func stopJvm(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_stopJvmServer",
		Args:       c.Args(),
	}
	result := client.InvokeCmd(cmd)
	if result.Ok != true {
		fmt.Printf("Failed to stop jvm manager: %v", result.Result)
	}
	fmt.Print(result.Result)
	return nil
}

func restartJvm(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_restartJvmServer",
		Args:       c.Args(),
	}
	result := client.InvokeCmd(cmd)
	if result.Ok != true {
		fmt.Printf("Failed to restart jvm manager: %v", result.Result)
	}
	fmt.Print(result.Result)
	return nil
}
