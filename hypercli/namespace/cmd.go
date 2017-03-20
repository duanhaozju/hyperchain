//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package namespace

import "github.com/urfave/cli"

//NewNamespaceCMD new namespace related
func NewNamespaceCMD() []cli.Command {
	return []cli.Command{
		{
			Name:    "start",
			Aliases: []string{"st"},
			Usage:   "start the namespace manager",
			Action:  start,
		},
		{
			Name:    "stop",
			Aliases: []string{"sp"},
			Usage:   "stop the namespace manager",
			Action:  stop,
		},
		{
			Name:    "startNamespace",
			Aliases: []string{"stn"},
			Usage:   "start namespace",
			Action:  startNamespace,
		},
		{
			Name:    "stopNamespace",
			Aliases: []string{"spn"},
			Usage:   "start namespace",
			Action:  stopNamespace,
		},
		{
			Name:    "restartNamespace",
			Aliases: []string{"rsn"},
			Usage:   "restart namespace",
			Action:  restartNamespace,
		},
		{
			Name:    "registerNamespace",
			Aliases: []string{"rrn"},
			Usage:   "register namespace",
			Action:  registerNamespace,
		},
		{
			Name:    "deregisterNamespace",
			Aliases: []string{"drn"},
			Usage:   "deregister namespace",
			Action:  deregisterNamespace,
		},
	}
}

func start(c *cli.Context) error {
	//TODO: impl start namespace manager
	return nil
}

func stop(c *cli.Context) error {
	//TODO: impl stop namespace manager
	return nil
}

func startNamespace(c *cli.Context) error {
	//TODO: impl start namespace
	return nil
}

func stopNamespace(c *cli.Context) error {
	//TODO: impl stop namespace
	return nil
}

func restartNamespace(c *cli.Context) error {
	//TODO: impl restart namespace
	return nil
}

func registerNamespace(c *cli.Context) error {
	//TODO: impl register namespace
	return nil
}

func deregisterNamespace(c *cli.Context) error {
	//TODO: impl deregister namespace
	return nil
}
