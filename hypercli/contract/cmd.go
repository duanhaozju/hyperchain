//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package contract

import (
	"github.com/urfave/cli"
	"fmt"
	"hyperchain/hypercli/common"
)

//NewContractCMD new contract related commands.
func NewContractCMD() []cli.Command {
	return []cli.Command{
		{
			Name:    "deploy",
			Aliases: []string{"-d"},
			Usage:   "Deploy a contract",
			Action:  deploy,
			Flags:   []cli.Flag{
				cli.StringFlag{
					Name:  "deploycmd, c",
					Value: "",
					Usage: "setting the payload of deploy contract",
				},
			},
		},
		{
			Name:    "invoke",
			Aliases: []string{"-i"},
			Usage:   "Invoke a contract method",
			Action:  invoke,
			Flags:   []cli.Flag{
				cli.StringFlag{
					Name:  "invokecmd, c",
					Value: "",
					Usage: "setting the payload of invoke contract",
				},
			},
		},
		{
			Name:    "destroy",
			Usage:   "Destroy a contract",
			Action:  destroy,
		},
	}
}

func deploy(c *cli.Context) error {
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))
	var deployCmd string
	if c.String("deploycmd") != "" {
		deployCmd = c.String("deploycmd")
	} else {
		deployParams := []string{"from", "nonce", "payload", "timestamp", "signature"}
		method := "contract_deployContract"
		deployCmd = getCmd(method, deployParams)
	}
	fmt.Println(deployCmd)
	client.Call(deployCmd)

	return nil
}

func invoke(c *cli.Context) error {
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))
	var invokeCmd string
	if c.String("invokecmd") != "" {
		invokeCmd = c.String("invokecmd")
	} else {
		invokeParams := []string{"from", "to", "nonce", "payload", "timestamp", "signature"}
		method := "contract_invokeContract"
		invokeCmd = getCmd(method, invokeParams)
	}
	fmt.Println(invokeCmd)
	client.Call(invokeCmd)

	return nil
}

func destroy(c *cli.Context) error {
	//TODO: implement destroy cmd
	return nil
}

func getCmd(method string, deploy_params []string) string {
	var namespace string
	fmt.Print("namespace: ")
	fmt.Scanln(&namespace)

	values := make([]string, len(deploy_params))
	args := "[{"
	for i, param := range deploy_params{
		fmt.Printf("%s: ", param)
		fmt.Scanln(&values[i])
		if i > 0 {
			args = args + ","
		}
		if param == "nonce" || param == "timestamp" {
			args = args + fmt.Sprintf("\"%s\":%s", param, values[i])
		}else {
			args = args + fmt.Sprintf("\"%s\":\"%s\"", param, values[i])
		}
	}
	args = args + "}]"

	return fmt.Sprintf(
		"{\"jsonrpc\":\"2.0\",\"namespace\":\"%s\",\"method\":\"%s\",\"params\":%s,\"id\":1}",
		namespace, method, args)
}
