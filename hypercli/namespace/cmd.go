//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package namespace

import (
	"fmt"
	"github.com/urfave/cli"
	"hyperchain/hypercli/common"
	admin "hyperchain/api/jsonrpc/core"
)

//NewNamespaceCMD new namespace related
func NewNamespaceCMD() []cli.Command {
	return []cli.Command{
		{
			Name:    "startNsMgr",
			Aliases: []string{"s"},
			Usage:   "start the namespace manager",
			Action:  start,
		},
		{
			Name:   "stopNsMgr",
			Usage:  "stop the namespace manager",
			Action: stop,
		},
		{
			Name:   "start",
			Usage:  "start namespace",
			Action: startNamespace,
		},
		{
			Name:   "stop",
			Usage:  "start namespace",
			Action: stopNamespace,
		},
		{
			Name:   "restart",
			Usage:  "restart namespace",
			Action: restartNamespace,
		},
		{
			Name:   "register",
			Usage:  "register namespace",
			Action: registerNamespace,
		},
		{
			Name:   "deregister",
			Usage:  "deregister namespace",
			Action: deregisterNamespace,
		},
		{
			Name:   "list",
			Usage:  "list current registered namespaces",
			Action: listNamespaces,
		},
	}
}

func start(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_startNsMgr",
		Args:       c.Args(),
	}
	client.InvokeCmd(cmd)
	return nil
}

func stop(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_stopNsMgr",
		Args:       c.Args(),
	}
	client.InvokeCmd(cmd)
	return nil
}

func startNamespace(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_startNamespace",
		Args:       c.Args(),
	}
	if len(cmd.Args) != 1 {
		fmt.Println(common.ErrInvalidArgsNum)
		return common.ErrInvalidArgsNum
	}
	client.InvokeCmd(cmd)
	return nil
}

func stopNamespace(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_stopNamespace",
		Args:       c.Args(),
	}
	if len(cmd.Args) != 1 {
		fmt.Println(common.ErrInvalidArgsNum)
		return common.ErrInvalidArgsNum
	}
	client.InvokeCmd(cmd)
	return nil
}

func restartNamespace(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_restartNamespace",
		Args:       c.Args(),
	}
	if len(cmd.Args) != 1 {
		fmt.Println(common.ErrInvalidArgsNum)
		return common.ErrInvalidArgsNum
	}
	client.InvokeCmd(cmd)
	return nil
}

func registerNamespace(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_registerNamespace",
		Args:       c.Args(),
	}
	if len(cmd.Args) != 1 {
		fmt.Println(common.ErrInvalidArgsNum)
		return common.ErrInvalidArgsNum
	}
	client.InvokeCmd(cmd)
	return nil
}

func deregisterNamespace(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_deregisterNamespace",
		Args:       c.Args(),
	}
	if len(cmd.Args) != 1 {
		fmt.Println(common.ErrInvalidArgsNum)
		return common.ErrInvalidArgsNum
	}
	client.InvokeCmd(cmd)
	return nil
}

func listNamespaces(c *cli.Context) error {
	client := common.GetCmdClient(c)
	cmd := &admin.Command{
		MethodName: "admin_listNamespaces",
		Args:       c.Args(),
	}
	client.InvokeCmd(cmd)
	return nil
}
