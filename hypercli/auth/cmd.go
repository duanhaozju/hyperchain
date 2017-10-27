//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package auth

import (
	"github.com/urfave/cli"
)

// NewLoginCMD new login commands.
func NewLoginCMD() cli.Command {
	return cli.Command{
		Name:   "login",
		Usage:  "login to the server with username and password",
		Action: login,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "username, u",
				Value: "",
				Usage: "specify the username to login",
			},
			cli.StringFlag{
				Name:  "password, p",
				Value: "",
				Usage: "specify the password to login",
			},
		},
	}
}

// NewLogoutCMD new logout commands.
func NewLogoutCMD() cli.Command {
	return cli.Command{
		Name:   "logout",
		Usage:  "logout from the server",
		Action: logout,
	}
}

// NewAuthCMD new auth related commands.
func NewAuthCMD() []cli.Command {
	return []cli.Command{
		{
			Name:   "create",
			Usage:  "create user(only permitted by root)",
			Action: createUser,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "group, g",
					Usage: "specify the user group",
				},
			},
		},
		{
			Name:   "alter",
			Usage:  "alter user(only permitted by root)",
			Action: alterUser,
		},
		{
			Name:   "drop",
			Usage:  "delete user(only permitted by root)",
			Action: dropUser,
		},
		{
			Name:   "grant",
			Usage:  "grant permissions to user(only permitted by root)",
			Action: grant,
		},
		{
			Name:   "revoke",
			Usage:  "revoke permissions from user(only permitted by root)",
			Action: revoke,
		},
		{
			Name:   "list",
			Usage:  "list permissions of given user",
			Action: list,
		},
	}
}
