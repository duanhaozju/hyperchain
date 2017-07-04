package auth

import (
	"fmt"
	"os"
	"github.com/urfave/cli"
	"hyperchain/hypercli/common"
)

const tokenpath = "./.token"

// login implements the login logic in hypercli
func login(c *cli.Context) error {
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))
	username := common.GetNonEmptyValueByName(c, "username")
	password := common.GetNonEmptyValueByName(c, "password")
	token, err := client.Login(username, password)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	//CurrentUser = username
	//fmt.Println(CurrentUser)

	err = common.SaveToFile(tokenpath, token)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return nil
}

// logout implements the logout logic in hypercli
func logout(c *cli.Context) {
	err := os.Remove(tokenpath)
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(1)
}
