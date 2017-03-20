//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package server

import (
	"github.com/urfave/cli"
	"fmt"
)

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
	//todo: impl start method
	fmt.Println("server start...")
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
