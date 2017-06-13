//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package contract

import (
	"encoding/hex"
	"fmt"
	"github.com/urfave/cli"
	"hyperchain/api/jsonrpc/core"
	"hyperchain/hypercli/common"
	"math/rand"
	"os"
	"time"
)

//NewContractCMD new contract related commands.
func NewContractCMD() []cli.Command {
	return []cli.Command{
		{
			Name:    "deploy",
			Aliases: []string{"d"},
			Usage:   "Deploy a contract",
			Action:  deploy,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "deploycmd, d",
					Value: "",
					Usage: "specify the payload of deploy contract",
				},
				cli.StringFlag{
					Name:  "namespace, n",
					Value: "global",
					Usage: "specify the namespace to deploy to, default is global",
				},
				cli.StringFlag{
					Name:  "from, f",
					Value: "000f1a7a08ccc48e5d30f80850cf1cf283aa3abd",
					Usage: "specify the deploy account",
				},
				cli.StringFlag{
					Name:  "payload, p",
					Value: "",
					Usage: "specify the deploy contract payload",
				},
			},
		},
		{
			Name:    "invoke",
			Aliases: []string{"i"},
			Usage:   "Invoke a contract method",
			Action:  invoke,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "invokecmd, i",
					Value: "",
					Usage: "specify the payload of invoke contract",
				},
				cli.StringFlag{
					Name:  "namespace, n",
					Value: "global",
					Usage: "specify the namespace to deploy to, default is global",
				},
				cli.StringFlag{
					Name:  "from, f",
					Value: "000f1a7a08ccc48e5d30f80850cf1cf283aa3abd",
					Usage: "specify the deploy account",
				},
				cli.StringFlag{
					Name:  "payload, p",
					Value: "",
					Usage: "specify the invoke contract payload",
				},

				//args with no default value which must be specified by user
				cli.StringFlag{
					Name:  "to, t",
					Value: "",
					Usage: "specify the destination account",
				},
			},
		},
		{
			Name:   "destroy",
			Usage:  "Destroy a contract",
			Action: destroy,
		},
	}
}

// deploy implements deploy contract and return the transaction receipt
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
	//fmt.Println(deployCmd)
	result, err := client.Call(deployCmd)
	if err != nil {
		fmt.Println("Error in call deploy cmd request")
		fmt.Print(err)
		os.Exit(1)
	}

	txHash := getTransactionHash(result)
	err = common.GetTransactionReceipt(txHash, c, client)
	if err != nil {
		fmt.Println("Error in call get transaction receipt")
		fmt.Println(err)
		os.Exit(1)
	}

	return nil
}

// invoke implements invoke contract and return the transaction receipt
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
	//fmt.Println(invokeCmd)
	result, err := client.Call(invokeCmd)
	if err != nil {
		fmt.Println("Error in call invoke cmd request")
		fmt.Print(err)
		os.Exit(1)
	}

	txHash := getTransactionHash(result)
	err = common.GetTransactionReceipt(txHash, c, client)
	if err != nil {
		fmt.Println("Error in call get transaction receipt")
		fmt.Println(err)
		os.Exit(1)
	}

	return nil
}

func destroy() error {
	fmt.Println("Not support yet!")
	return nil
}

// getCmd returns the expected jsonrpc command from specified method and deploy_params
func getCmd(method string, deploy_params []string, c *cli.Context) string {
	namespace := c.String("namespace")
	var from, to, payload string
	var nonce, timestamp, amount int64
	var opcode int

	params := "[{"
	for i, param := range deploy_params {
		if i > 0 {
			params = params + ","
		}
		switch param {
		case "payload":
			payload = common.GetNonEmptyValueByName(c, "payload")
			params = params + fmt.Sprintf("\"%s\":\"%s\"", param, payload)

		case "nonce":
			nonce = rand.Int63()
			params = params + fmt.Sprintf("\"%s\":%d", param, nonce)

		case "timestamp":
			timestamp = time.Now().UnixNano()
			params = params + fmt.Sprintf("\"%s\":%d", param, timestamp)

		case "from":
			from = c.String("from")
			params = params + fmt.Sprintf("\"%s\":\"%s\"", param, from)

		//below params must be input by user
		case "to":
			to = common.GetNonEmptyValueByName(c, "to")
			params = params + fmt.Sprintf("\"%s\":\"%s\"", param, to)

		// signature is generated automatically
		case "signature":
			amount = 0
			opcode = 0
			sig, err := common.GenSignature(from, to, timestamp, amount, payload, nonce, int32(opcode))
			if err != nil {
				fmt.Println("Error in generate signature.")
				fmt.Println(err)
				os.Exit(1)
			}
			signature := hex.EncodeToString(sig)
			params = params + fmt.Sprintf("\"%s\":\"%s\"", param, signature)

		default:
			fmt.Printf("Invalid param name: %s\n", param)
			os.Exit(1)
		}
	}
	params = params + "}]"

	return fmt.Sprintf(
		"{\"jsonrpc\":\"2.0\",\"namespace\":\"%s\",\"method\":\"%s\",\"params\":%s,\"id\":1}",
		namespace, method, params)
}

// getTransactionHash gets the hash of the transaction from the json-format return value
func getTransactionHash(result *jsonrpc.CommandResult) string {
	response, err := common.GetJSONResponse(result)
	if err != nil {
		fmt.Println("Error in call get transaction hash from http response")
		fmt.Println(err)
		os.Exit(1)
		return ""
	}

	if hash, ok := response.Result.(string); !ok {
		fmt.Println("Error in call get transaction hash from http response")
		fmt.Printf("rpc result: %v can't parse to string", response.Result)
		return ""
	} else {
		return hash
	}
}
