//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package log

import "github.com/urfave/cli"

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
	//TODO: impl getLevel
	return nil
}

func setLevel(c *cli.Context) error {
	//TODO: impl setLevel
	return nil
}
