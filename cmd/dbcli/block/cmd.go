package block

import (
	"fmt"
	"github.com/hyperchain/hyperchain/cmd/dbcli/constant"
	"github.com/hyperchain/hyperchain/cmd/dbcli/database"
	"github.com/hyperchain/hyperchain/cmd/dbcli/utils"
	"github.com/hyperchain/hyperchain/cmd/dbcli/version"
	"github.com/urfave/cli"
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
					Name:  "verbose",
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
					Name:  "verbose",
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
		{
			Name:   "getAllBlockSequential",
			Usage:  "get all block sequential",
			Action: getAllBlockSequential,
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
					Name:  "verbose",
					Value: "false",
					Usage: "specify the transaction content",
				},
			},
		},
		{
			Name:   "getBlockRange",
			Usage:  "get block within the range",
			Action: getBlockRange,
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
					Name:  "verbose",
					Value: "false",
					Usage: "specify the transaction content",
				},
				cli.StringFlag{
					Name:  "min",
					Value: "0",
					Usage: "specify the min block number",
				},
				cli.StringFlag{
					Name:  "max",
					Value: "0",
					Usage: "specify the max block number",
				},
			},
		},
	}
}

// getBlockByNumber -- get the block By block number
func getBlockByNumber(c *cli.Context) {
	if c.String(constant.PATH) != "" && c.String(constant.DATABASE) != "" && c.Uint64(constant.NUMBER) >= 0 {
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
		result, err := hyperChain.GetBlockByNumber(c.Uint64(constant.NUMBER), parameter)
		if err != nil {
			fmt.Println(constant.ErrQuery.Error(), err.Error())
		} else {
			if c.String(constant.OUTPUT) != "" {
				file := utils.CreateOrAppend(c.String(constant.OUTPUT), result)
				defer utils.Close(file)
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
		defer hyperChain.GetDB().Close()
		parameter := &constant.Parameter{
			Verbose: c.Bool(constant.VERBOSE),
		}
		result, err := hyperChain.GetBlockByHash(c.String(constant.HASH), parameter)
		if err != nil {
			fmt.Println(constant.ErrQuery.Error(), err.Error())
		} else {
			if c.String(constant.OUTPUT) != "" {
				file := utils.CreateOrAppend(c.String(constant.OUTPUT), result)
				defer utils.Close(file)
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
		defer hyperChain.GetDB().Close()
		result, err := hyperChain.GetBlockHashByNum(c.Uint64(constant.NUMBER))
		if err != nil {
			fmt.Println(constant.ErrQuery.Error(), err.Error())
		} else {
			if c.String(constant.OUTPUT) != "" {
				file := utils.CreateOrAppend(c.String(constant.OUTPUT), result)
				defer utils.Close(file)
			} else {
				fmt.Println(utils.Decorate(result))
			}
		}
	} else {
		fmt.Println(constant.ErrInvalidParams.Error())
	}
}

// getBlockRange -- get block within the range
func getBlockRange(c *cli.Context) {
	if c.String(constant.PATH) != "" && c.String(constant.DATABASE) != "" && c.Uint64(constant.MIN) >= 0 && c.Uint64(constant.MAX) >= 0 {
		path := c.String(constant.PATH)
		db := c.String(constant.DATABASE)
		min := c.Uint64(constant.MIN)
		max := c.Uint64(constant.MAX)
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
		hyperChain.GetBlockRange(c.String(constant.OUTPUT), min, max, parameter)
	} else {
		fmt.Println(constant.ErrInvalidParams.Error())
	}
}

// getAllBlockSequential -- get all block sequential
func getAllBlockSequential(c *cli.Context) {
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
		hyperChain.GetAllBlockSequential(c.String(constant.OUTPUT), parameter)
	} else {
		fmt.Println(constant.ErrInvalidParams.Error())
	}
}
