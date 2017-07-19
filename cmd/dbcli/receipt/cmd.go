package receipt

import (
	"fmt"
	"github.com/urfave/cli"
	"hyperchain/cmd/dbcli/constant"
	"hyperchain/cmd/dbcli/database"
	"hyperchain/cmd/dbcli/utils"
	"hyperchain/cmd/dbcli/version"
)

func NewReceiptCMD() []cli.Command {
	return []cli.Command{
		{
			Name:   "getReceipt",
			Usage:  "get receipt from database",
			Action: getReceipt,
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
					Name:  "hash",
					Value: "",
					Usage: "specify the transaction hash",
				},
				cli.StringFlag{
					Name:  "output, o",
					Value: "",
					Usage: "specify the output file",
				},
			},
		},
	}
}

// getReceipt -- get the receipt by transaction hash
func getReceipt(c *cli.Context) {
	if c.String(constant.PATH) != "" && c.String(constant.DATABASE) != "" && c.String(constant.HASH) != "" {
		path := c.String(constant.PATH)
		db := c.String(constant.DATABASE)
		database, err := database.DBFactory(db, path)
		if err != nil {
			fmt.Println(constant.ErrDBInit.Error(), err.Error())
			return
		}
		hyperChain := version.NewVersion(database)
		defer hyperChain.GetDB().Close()
		result, err := hyperChain.GetReceipt(c.String(constant.HASH))
		if err != nil {
			fmt.Println(constant.ErrQuery.Error(), err.Error())
		} else {
			if c.String(constant.OUTPUT) != "" {
				utils.CreateOrAppend(c.String(constant.OUTPUT), result)
			} else {
				fmt.Println(utils.Decorate(result))
			}
		}
	} else {
		fmt.Println(constant.ErrInvalidParams.Error())
	}
}
