//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package contract

import (
	"github.com/urfave/cli"
)

//NewContractCMD new contract related commands.
func NewContractCMD() []cli.Command {
	return []cli.Command{
		{
			Name:    "invoke",
			Aliases: []string{"-i"},
			Usage:   "Invoke a contract method",
			Action:  invoke,
		},
		{
			Name:    "deploy",
			Aliases: []string{"-d"},
			Usage:   "Deploy a contract",
			Action:  deploy,
		},
		{
			Name:    "destroy",
			Usage:   "Destroy a contract",
			Action:  destroy,
		},
	}
}

func invoke(c *cli.Context) error {
	//TODO: implement invoke cmd
	return nil
}

func deploy(c *cli.Context) error {
	//TODO: implement deploy cmd
	return nil
}

func destroy(c *cli.Context) error {
	//TODO: implement destroy cmd
	return nil
}
