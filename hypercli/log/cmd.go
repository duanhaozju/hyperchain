//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package log

import (
	"fmt"

	"github.com/hyperchain/hyperchain/api/admin"
	"github.com/hyperchain/hyperchain/hypercli/common"

	"github.com/urfave/cli"
)

// NewLogCMD new log related commands.
func NewLogCMD() []cli.Command {
	return []cli.Command{
		{
			Name:   "getLevel",
			Usage:  "getLevel get a logger's level",
			Action: getLevel,
		},
		{
			Name:   "setLevel",
			Usage:  "setLevel set a logger's level",
			Action: setLevel,
		},
	}
}

// setLevel sets the log level of specified namespace:module with 3 params:
// the first param specifies the namespace
// the second param specifies the module(such as p2p, consensus...)
// the third param specifies the new log level user wants to set.
func setLevel(c *cli.Context) error {
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))
	cmd := &admin.Command{
		MethodName: "admin_setLevel",
		Args:       c.Args(),
	}
	if len(cmd.Args) != 3 {
		fmt.Println(common.ErrInvalidArgsNum)
		return common.ErrInvalidArgsNum
	}

	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}

// getLevel gets the log level of specified namespace:module with 2 params:
// the first param specifies the namespace
// the second param specifies the module(such as p2p, consensus...)
func getLevel(c *cli.Context) error {
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))
	cmd := &admin.Command{
		MethodName: "admin_getLevel",
		Args:       c.Args(),
	}
	if len(cmd.Args) != 2 {
		fmt.Println(common.ErrInvalidArgsNum)
		return common.ErrInvalidArgsNum
	}

	result := client.InvokeCmd(cmd)
	fmt.Print(result)
	return nil
}
