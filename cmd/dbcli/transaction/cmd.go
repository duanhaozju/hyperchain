package transaction

import (
	"fmt"
	"github.com/hyperchain/hyperchain/cmd/dbcli/constant"
	"github.com/hyperchain/hyperchain/cmd/dbcli/database"
	"github.com/hyperchain/hyperchain/cmd/dbcli/utils"
	"github.com/hyperchain/hyperchain/cmd/dbcli/version"
	"github.com/urfave/cli"
)

func NewTransactionCMD() []cli.Command {
	return []cli.Command{
		{
			Name:   "getTransactionByHash",
			Usage:  "get the transaction by hash",
			Action: getTransactionByHash,
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
		{
			Name:   "getAllTransactionSequential",
			Usage:  "get all transactions sequential",
			Action: getAllTransactionSequential,
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
			Name:   "getTransactionMetaByHash",
			Usage:  "get the transaction meta by hash",
			Action: getTransactionMetaByHash,
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
		{
			Name:   "getAllDiscardTransaction",
			Usage:  "get all discard transactions",
			Action: getAllDiscardTransaction,
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
			},
		},
		{
			Name:   "getDiscardTransactionByHash",
			Usage:  "get the discard transaction by hash",
			Action: getDiscardTransactionByHash,
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

// getTransaction -- get the transaction By transaction hash
func getTransactionByHash(c *cli.Context) {
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
		result, err := hyperChain.GetTransactionByHash(c.String(constant.HASH))
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

// getAllTransactionSequential -- get all transactions sequential
func getAllTransactionSequential(c *cli.Context) {
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
		hyperChain.GetAllTransactionSequential(c.String(constant.OUTPUT), parameter)
	} else {
		fmt.Println(constant.ErrInvalidParams.Error())
	}
}

// getTransactionMetaByHash -- get transaction meta by transaction hash
func getTransactionMetaByHash(c *cli.Context) {
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
		result, err := hyperChain.GetTxWithBlock(c.String(constant.HASH))
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

func getDiscardTransactionByHash(c *cli.Context) {
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
		result, err := hyperChain.GetDiscardTransaction(c.String(constant.HASH))
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

func getAllDiscardTransaction(c *cli.Context) {
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
		hyperChain.GetAllDiscardTransaction(c.String(constant.OUTPUT))
	} else {
		fmt.Println(constant.ErrInvalidParams.Error())
	}
}
