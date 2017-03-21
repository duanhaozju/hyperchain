//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package server

import (
	"github.com/op/go-logging"
	"github.com/urfave/cli"
	"hyperchain/api"
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
			Aliases: []string{"s"},
			Usage:   "start hyperchain server",
			Action:  start,
		},
		{
			Name:    "stop",
			Aliases: []string{"sp"},
			Usage:   "stop hyperchain server",
			Action:  stop,
		},
		{
			Name:    "restart",
			Aliases: []string{"r"},
			Usage:   "restart hyperchain server",
			Action:  restart,
		},
	}
}

func start(c *cli.Context) error {
	logger.Info("server start...")
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))
	cmd := &hpc.Command{
		MethodName: "admin_stopServer",
		Args:       []string{},
	}
	rs := client.InvokeCmd(cmd)
	logger.Critical(rs)
	return nil
}

func stop(c *cli.Context) error {
	//todo: impl stop method
	return nil
}

func restart(c *cli.Context) error {
	//todo: impl restart method
	return nil
}
