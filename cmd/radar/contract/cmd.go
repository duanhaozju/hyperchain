package contract

import (
	"encoding/json"
	"fmt"
	"github.com/hyperchain/hyperchain/cmd/radar/core/api"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/urfave/cli"
	"io/ioutil"
)

func NewContractCMD() []cli.Command {
	return []cli.Command{
		{
			Name:   "getStorage",
			Usage:  "get the storage of contract",
			Action: getStorage,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "contractFile, c",
					Value: "",
					Usage: "specify the path of contract file",
				},
				cli.StringFlag{
					Name:  "database, db",
					Value: "",
					Usage: "specify the path of database",
				},
				cli.StringFlag{
					Name:  "key",
					Value: "",
					Usage: "specify the key of map in contract",
				},
				cli.StringFlag{
					Name:  "address",
					Value: "",
					Usage: "specify the contract address",
				},
			},
		},
	}
}

func getStorage(c *cli.Context) {
	if c.String("contractFile") != "" && c.String("database") != "" && c.String("address") != "" {
		contractFile := c.String("contractFile")
		database := c.String("database")
		contractAddress := c.String("address")

		db, err := leveldb.OpenFile(database, nil)
		defer db.Close()
		if err != nil {
			fmt.Println("radar/cmd:", err.Error())
		}

		var originKeyOfMaps map[string]map[string][][]string
		originKeyOfMaps = make(map[string]map[string][][]string)
		if c.String("key") != "" {
			content, err := ioutil.ReadFile(c.String("key"))
			if err != nil {
				fmt.Println("radar/cmd:", err.Error())
				return
			}
			var temp Keys
			err = json.Unmarshal(content, &temp)
			if err != nil {
				fmt.Println("radar/cmd:", err.Error())
				return
			}
			originKeyOfMaps[temp.ContractName] = temp.Key
		}
		res, err := api.GetResult(contractFile, db, contractAddress, originKeyOfMaps)
		if err != nil {
			fmt.Println("radar/cmd:", err.Error())
		} else {
			for k, v := range res {
				fmt.Println("contract:", k)
				for i := 0; i < len(v); i++ {
					fmt.Println("\t", v[i])
				}
			}
		}
	} else {
		fmt.Println("radar/cmd: invalid params.")
	}
}

type Keys struct {
	ContractName string
	Key          map[string][][]string
}
