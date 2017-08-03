package account

import (
	"github.com/urfave/cli"
	"hyperchain/cmd/dbcli/database"
	"fmt"
	"hyperchain/cmd/dbcli/constant"
	"hyperchain/cmd/dbcli/version"
)

func NewAccountCMD() []cli.Command {
	return []cli.Command{
		{
			Name:   "getAccountByAddress",
			Usage:  "get the account by address",
			Action: getAccountByAddress,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "path, p",
					Value: "",
					Usage: "specify the path of db",
				},
				cli.StringFlag{
					Name:  "database, db",
					Value: "leveldb",
					Usage: "specify the database using",
				},
				cli.StringFlag{
					Name:  "address",
					Value: "",
					Usage: "specify the account address",
				},
				cli.StringFlag{
					Name:  "output, o",
					Value: "",
					Usage: "specify the output file",
				},
				cli.StringFlag{
					Name: "verbose",
					Value: "false",
					Usage: "specify the account content",
				},
			},
		},
		{
			Name:   "getAllAccount",
			Usage:  "get all account",
			Action: getAllAccount,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "path, p",
					Value: "",
					Usage: "specify the path of db",
				},
				cli.StringFlag{
					Name:  "database, db",
					Value: "leveldb",
					Usage: "specify the database using",
				},
				cli.StringFlag{
					Name:  "output, o",
					Value: "",
					Usage: "specify the output file",
				},
				cli.StringFlag{
					Name: "verbose",
					Value: "false",
					Usage: "specify the account content",
				},
			},
		},
	}
}

// getAccountByAddress -- get account by address
func getAccountByAddress(c *cli.Context) {
	if c.String(constant.PATH) != "" && c.String(constant.DATABASE) != "" && c.String(constant.ADDRESS) != "" {
		path := c.String(constant.PATH)
		db := c.String(constant.DATABASE)
		address := c.String(constant.ADDRESS)
		database, err := database.DBFactory(db, path)
		if err != nil {
			fmt.Println(constant.ErrDBInit.Error(), err.Error())
			return
		}
		hyperChain := version.NewVersion(database)
		defer hyperChain.GetDB().Close()
		parameter := &constant.Parameter{
			Verbose: c.Bool(constant.VERBOSE),
		}
		hyperChain.GetAccountByAddress(address, c.String(constant.OUTPUT), parameter)
	} else {
		fmt.Println(constant.ErrInvalidParams.Error())
	}
}

// getAllAccount -- get all account
func getAllAccount(c *cli.Context) {
	if c.String(constant.PATH) != "" && c.String(constant.DATABASE) != "" {
		path := c.String(constant.PATH)
		db := c.String(constant.DATABASE)
		database, err := database.DBFactory(db, path)
		if err != nil {
			fmt.Println(constant.ErrDBInit.Error(), err.Error())
			return
		}
		hyperChain := version.NewVersion(database)
		defer hyperChain.GetDB().Close()
		parameter := &constant.Parameter{
			Verbose: c.Bool(constant.VERBOSE),
		}
		hyperChain.GetAllAccount(c.String(constant.OUTPUT), parameter)
	} else {
		fmt.Println(constant.ErrInvalidParams.Error())
	}
}