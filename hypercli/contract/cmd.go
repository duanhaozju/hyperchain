//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package contract

import (
	"github.com/urfave/cli"
	"fmt"
	"hyperchain/hypercli/common"
	"os/exec"
	"io/ioutil"
	"encoding/hex"
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
					Usage: "specify the payload of deploy contract",
				},
				cli.StringFlag{
					Name:  "jvm, j",
					Value: "",
					Usage: "specify the type how the contract is generated, j represented java",
				},
				cli.StringFlag{
					Name:  "path, p",
					Value: "",
					Usage: "specify the contract file path",
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
		deployCmd = getCmd(method, deployParams, c)
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
		invokeCmd = getCmd(method, invokeParams, c)
	}
	fmt.Println(invokeCmd)
	client.Call(invokeCmd)

	return nil
}

func destroy(c *cli.Context) error {
	//TODO: implement destroy cmd
	return nil
}

func getCmd(method string, deploy_params []string, c *cli.Context) string {
	var namespace string
	fmt.Print("namespace: ")
	fmt.Scanln(&namespace)

	values := make([]string, len(deploy_params))
	args := "[{"
	for i, param := range deploy_params{
		if i > 0 {
			args = args + ","
		}

		if param == "payload" && c.String("jvm") == "true" && c.String("path") != "" {
			args = args + fmt.Sprintf("\"%s\":\"%s\"", param, getPayloadFromPath(c.String("path")))
			continue
		}
		fmt.Printf("%s: ", param)
		fmt.Scanln(&values[i])

		if param == "nonce" || param == "timestamp" {
			args = args + fmt.Sprintf("\"%s\":%s", param, values[i])
		}else {
			args = args + fmt.Sprintf("\"%s\":\"%s\"", param, values[i])
		}
	}
	if c.String("jvm") == "true" {
		args = args + ","
		args = args + fmt.Sprint("\"type\":\"jvm\"")
	}
	args = args + "}]"

	return fmt.Sprintf(
		"{\"jsonrpc\":\"2.0\",\"namespace\":\"%s\",\"method\":\"%s\",\"params\":%s,\"id\":1}",
		namespace, method, args)

}

func getPayloadFromPath (path string) string {
	fmt.Println("start get payload from path...")
	target := "contract.tar.gz"
	compress(path, target)
	buf, err := ioutil.ReadFile(target)
	if err != nil {
		fmt.Printf("Error in read compressed file: %s", target)
		fmt.Println(err.Error())
		return ""
	}
	payload := hex.EncodeToString(buf)
	fmt.Println(payload)
	delCompressedFile(target)

	return payload
}

func compress(source, target string) {
	command := exec.Command("tar", "-czf", target, source)
	if err := command.Run(); err != nil {
		fmt.Printf("Error in read compress specefied file: %s", source)
		fmt.Println(err.Error())
	}
}

func delCompressedFile(file string) {
	command := exec.Command("rm", "-rf", file)
	if err := command.Run(); err != nil {
		fmt.Printf("Error in remove compressed file: %s", file)
		fmt.Println(err.Error())
	}
}