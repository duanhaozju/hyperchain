package account

import (
	"github.com/urfave/cli"
	"hyperchain/cmd/dbcli/database"
	"fmt"
	"hyperchain/cmd/dbcli/constant"
	"hyperchain/cmd/dbcli/version"
	"io/ioutil"
	"path"
	"os"
	"io"
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
				cli.StringFlag{
					Name: "number",
					Value: "-1",
					Usage: "specify the block number",
				},
				cli.StringFlag{
					Name: "ns",
					Value: "global",
					Usage: "specify the namespace",
				},
				cli.StringFlag{
					Name: "globalconf",
					Value: "../../configuration/namespaces/global/config/namespace.toml",
					Usage: "specify the namespace global config",
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
				cli.StringFlag{
					Name: "number",
					Value: "-1",
					Usage: "specify the block number",
				},
				cli.StringFlag{
					Name: "ns",
					Value: "global",
					Usage: "specify the namespace",
				},
				cli.StringFlag{
					Name: "globalconf",
					Value: "../../configuration/namespaces/global/config/namespace.toml",
					Usage: "specify the namespace global config",
				},
			},
		},
	}
}

// getAccountByAddress -- get account by address
func getAccountByAddress(c *cli.Context) {
	if c.String(constant.PATH) != "" && c.String(constant.DATABASE) != "" && c.String(constant.ADDRESS) != "" {
		path := c.String(constant.PATH)
		var err error
		if c.Int64(constant.NUMBER) >= 0 {
			path, err = createDBCopy(c.String(constant.PATH))
			defer os.RemoveAll(path)
			if err != nil {
				fmt.Println(constant.ErrCreateDBCopy.Error(), err.Error())
				return
			}
		}
		db := c.String(constant.DATABASE)
		address := c.String(constant.ADDRESS)
		dataBase, err := database.DBFactory(db, path)
		if err != nil {
			fmt.Println(constant.ErrDBInit.Error(), err.Error())
			return
		}
		hyperChain := version.NewVersion(dataBase)
		parameter := &constant.Parameter{
			Verbose: c.Bool(constant.VERBOSE),
		}
		if c.Int64(constant.NUMBER) >= 0 {
			hyperChain.RevertDB(c.String(constant.NS), c.String(constant.GLOBALCONF), path, c.Uint64(constant.NUMBER), parameter)
		}
		dataBase, err = database.DBFactory(db, path)
		if err != nil {
			fmt.Println(constant.ErrDBInit.Error(), err.Error())
			return
		}
		hyperChain = version.NewVersion(dataBase)
		hyperChain.GetAccountByAddress(address, c.String(constant.OUTPUT), parameter)
		defer hyperChain.GetDB().Close()
	} else {
		fmt.Println(constant.ErrInvalidParams.Error())
	}
}

// getAllAccount -- get all account
func getAllAccount(c *cli.Context) {
	if c.String(constant.PATH) != "" && c.String(constant.DATABASE) != "" {
		path := c.String(constant.PATH)
		var err error
		if c.Int64(constant.NUMBER) >= 0 {
			path, err = createDBCopy(c.String(constant.PATH))
			defer os.RemoveAll(path)
			if err != nil {
				fmt.Println(constant.ErrCreateDBCopy.Error(), err.Error())
				return
			}
		}
		db := c.String(constant.DATABASE)
		dataBase, err := database.DBFactory(db, path)
		if err != nil {
			fmt.Println(constant.ErrDBInit.Error(), err.Error())
			return
		}
		hyperChain := version.NewVersion(dataBase)
		parameter := &constant.Parameter{
			Verbose: c.Bool(constant.VERBOSE),
		}
		if c.Int64(constant.NUMBER) >= 0 {
			hyperChain.RevertDB(c.String(constant.NS), c.String(constant.GLOBALCONF), path, c.Uint64(constant.NUMBER), parameter)
		}
		dataBase, err = database.DBFactory(db, path)
		if err != nil {
			fmt.Println(constant.ErrDBInit.Error(), err.Error())
			return
		}
		hyperChain = version.NewVersion(dataBase)
		hyperChain.GetAllAccount(c.String(constant.OUTPUT), parameter)
		defer hyperChain.GetDB().Close()
	} else {
		fmt.Println(constant.ErrInvalidParams.Error())
	}
}

func createDBCopy(dbPath string) (string, error) {
	tmpDir, err := ioutil.TempDir("", "copy")
	if err != nil {
		return "", err
	}
	fileInfos, err := ioutil.ReadDir(dbPath)
	if err != nil {
		return "", err
	}
	for i:=0 ; i < len(fileInfos); i++ {
		if fileInfos[i].IsDir() {
			continue
		}
		file := path.Join(dbPath, fileInfos[i].Name())
		src, err := os.Open(file)
		if err != nil {
			return "", err
		}
		dst, err := os.OpenFile(path.Join(tmpDir, fileInfos[i].Name()), os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return "", err
		}
		_, err = io.Copy(dst, src)
		if err != nil {
			return "", err
		}
	}
	return tmpDir, nil
}