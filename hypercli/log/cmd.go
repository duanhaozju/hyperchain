//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package log

import (
	"fmt"
	"github.com/urfave/cli"
	"hyperchain/hypercli/common"
	admin "hyperchain/api/jsonrpc/core"
)

//NewLogCMD new log related commands.
func NewLogCMD() []cli.Command {
	return []cli.Command{
		{
			Name:    "getLevel",
			Aliases: []string{"-g"},
			Usage:   "getLevel get a logger's level",
			Action:  getLevel,
		},
		{
			Name:    "setLevel",
			Aliases: []string{"-s"},
			Usage:   "setLevel set a logger's level",
			Action:  setLevel,
		},
	}
}

func getLevel(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_setLevel",
		Args:       c.Args(),
	}
	if len(cmd.Args) != 3 {
		fmt.Println(common.ErrInvalidArgsNum)
		return common.ErrInvalidArgsNum
	}

	rs := client.InvokeCmd(cmd)
	fmt.Println(rs)
	return nil
}

func setLevel(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_getLevel",
		Args:       c.Args(),
	}
	fmt.Println(cmd.Args)
	if len(cmd.Args) != 2 {
		fmt.Println(common.ErrInvalidArgsNum)
		return common.ErrInvalidArgsNum
	}
	rs := client.InvokeCmd(cmd)
	fmt.Println(rs)
	return nil
}
