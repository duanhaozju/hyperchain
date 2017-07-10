package block

import (
	"fmt"
	"github.com/urfave/cli"
	"hyperchain/cmd/dbcli/constant"
	"hyperchain/cmd/dbcli/database"
	"hyperchain/cmd/dbcli/utils"
	"hyperchain/cmd/dbcli/version"
)

func NewBlockCMD() []cli.Command {
	return []cli.Command{
		{
			Name:   "getBlockByNumber",
			Usage:  "get the block by block number",
			Action: getBlockByNumber,
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
					Name:  "number, num",
					Value: "",
					Usage: "specify the block number",
				},
				cli.StringFlag{
					Name:  "output, o",
					Value: "",
					Usage: "specify the output file",
				},
				cli.StringFlag{
					Name: "verbose",
					Value: "false",
					Usage: "specify the transaction content",
				},
			},
		},
		{
			Name:   "getBlockByHash",
			Usage:  "get the block by block hash",
			Action: getBlockByHash,
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
					Usage: "specify the block hash",
				},
				cli.StringFlag{
					Name:  "output, o",
					Value: "",
					Usage: "specify the output file",
				},
				cli.StringFlag{
					Name: "verbose",
					Value: "false",
					Usage: "specify the transaction content",
				},
			},
		},
		{
			Name:   "getBlockHashByNumber",
			Usage:  "get the block hash by block number",
			Action: getBlockHashByNumber,
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
					Name:  "number, num",
					Value: "",
					Usage: "specify the block number",
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

// getBlockByNumber -- get the block By block number
func getBlockByNumber(c *cli.Context) {
	if c.String(constant.PATH) != "" && c.String(constant.DATABASE) != "" && c.Uint64(constant.NUMBER) > 0 {
		path := c.String(constant.PATH)
		db := c.String(constant.DATABASE)
		database, err := database.DBFactory(db, path)
		if err != nil {
			fmt.Println(constant.ErrDBInit.Error(), err.Error())
			return
		}
		hyperChain := version.NewVersion(database)
		parameter := &constant.Parameter{
			Verbose: c.Bool(constant.VERBOSE),
		}
		result, err := hyperChain.GetBlockByNumber(c.Uint64(constant.NUMBER), parameter)
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

// getBlockByHash -- get the block By block hash
func getBlockByHash(c *cli.Context) {
	if c.String(constant.PATH) != "" && c.String(constant.DATABASE) != "" && c.String(constant.HASH) != "" {
		path := c.String(constant.PATH)
		db := c.String(constant.DATABASE)
		database, err := database.DBFactory(db, path)
		if err != nil {
			fmt.Println(constant.ErrDBInit.Error(), err.Error())
			return
		}
		hyperChain := version.NewVersion(database)
		parameter := &constant.Parameter{
			Verbose: c.Bool(constant.VERBOSE),
		}
		result, err := hyperChain.GetBlockByHash(c.String(constant.HASH), parameter)
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

// GetBlockHashByNum -- retrieve block hash with related block number.
func getBlockHashByNumber(c *cli.Context) {
	if c.String(constant.PATH) != "" && c.String(constant.DATABASE) != "" && c.Uint64(constant.NUMBER) > 0 {
		path := c.String(constant.PATH)
		db := c.String(constant.DATABASE)
		database, err := database.DBFactory(db, path)
		if err != nil {
			fmt.Println(constant.ErrDBInit.Error(), err.Error())
			return
		}
		hyperChain := version.NewVersion(database)
		result, err := hyperChain.GetBlockHashByNum(c.Uint64(constant.NUMBER))
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
