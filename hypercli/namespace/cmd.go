//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package namespace

import "github.com/urfave/cli"

//NewNamespaceCMD new namespace related
func NewNamespaceCMD() []cli.Command {
	return []cli.Command{
		{
			Name:    "start",
			Aliases: []string{"-s"},
			Usage:   "start the namespace manager",
			Action:  start,
		},
		{
			Name:    "stop",
			Usage:   "stop the namespace manager",
			Action:  stop,
		},
		{
			Name:    "startNamespace",
			Usage:   "start namespace",
			Action:  startNamespace,
		},
		{
			Name:    "stopNamespace",
			Usage:   "start namespace",
			Action:  stopNamespace,
		},
		{
			Name:    "restartNamespace",
			Usage:   "restart namespace",
			Action:  restartNamespace,
		},
		{
			Name:    "registerNamespace",
			Usage:   "register namespace",
			Action:  registerNamespace,
		},
		{
			Name:    "deregisterNamespace",
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
