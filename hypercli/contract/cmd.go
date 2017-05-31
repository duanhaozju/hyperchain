//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package contract

import (
	"github.com/urfave/cli"
	"fmt"
	"hyperchain/hypercli/common"
	"io/ioutil"
	"encoding/hex"
	"math/rand"
	"time"
	"os"
	"hyperchain/core/types"
	"hyperchain/api/jsonrpc/core"
	"strconv"
)

//NewContractCMD new contract related commands.
func NewContractCMD() []cli.Command {
	return []cli.Command{
		{
			Name:    "deploy",
			Usage:   "Deploy a contract",
			Action:  deploy,
			Flags:   []cli.Flag{
				cli.StringFlag{
					Name:  "deploycmd, c",
					Value: "",
					Usage: "specify the payload of deploy contract",
				},
				cli.BoolFlag{
					Name:  "jvm, j",
					Usage: "specify how the contract is generated, false is solidity, true is jvm",
				},
				cli.StringFlag{
					Name:  "directory, d",
					Value: "",
					Usage: "specify the contract file directory",
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
			Usage:   "Invoke a contract method",
			Action:  invoke,
			Flags:   []cli.Flag{
				cli.StringFlag{
					Name:  "invokecmd, i",
					Value: "",
					Usage: "specify the payload of invoke contract",
				},
				cli.BoolFlag{
					Name:  "jvm, j",
					Usage: "specify how the contract is generated, false is solidity, true is jvm",
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
				cli.StringFlag{
					Name:  "method, m",
					Value: "",
					Usage: "specify the method of invoke contract",
				},
				cli.StringFlag{
					Name:  "args, a",
					Value: "",
					Usage: "specify the args of invoke contract",
				},
			},
		},
		{
			Name:    "maintain",
			Usage:   "Invoke a contract method",
			Action:  maintain,
			Flags:   []cli.Flag{
				cli.StringFlag{
					Name:  "maintaincmd, m",
					Value: "",
					Usage: "specify the payload of maintain contract",
				},
				cli.BoolFlag{
					Name:  "jvm, j",
					Usage: "specify how the contract is generated, false is solidity, true is jvm",
				},
				cli.StringFlag{
					Name:  "directory, d",
					Value: "",
					Usage: "specify the contract file directory",
				},
				cli.StringFlag{
					Name:  "namespace, n",
					Value: "global",
					Usage: "specify the namespace to maintain, default is global",
				},
				cli.StringFlag{
					Name:  "from, f",
					Value: "000f1a7a08ccc48e5d30f80850cf1cf283aa3abd",
					Usage: "specify the maintain account",
				},
				cli.StringFlag{
					Name:  "payload, p",
					Value: "",
					Usage: "specify the maintain contract payload",
				},

				//args with no default value which must be specified by user
				cli.StringFlag{
					Name:  "to, t",
					Value: "",
					Usage: "specify the destination account",
				},
				cli.StringFlag{
					Name:  "opcode, o",
					Value: "",
					Usage: "specify the opcode",
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
		fmt.Println(err)
		os.Exit(1)
	}
	//fmt.Println(result.Result)

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
		invokeParams := []string{"from", "to", "nonce", "payload", "timestamp", "signature", "method", "args"}
		method := "contract_invokeContract"
		invokeCmd = getCmd(method, invokeParams, c)
	}
	//fmt.Println(invokeCmd)
	result, err := client.Call(invokeCmd)
	if err != nil {
		fmt.Println("Error in call invoke cmd request")
		fmt.Println(err)
		os.Exit(1)
	}
	//fmt.Println(result.Result)

	txHash := getTransactionHash(result)
	err = common.GetTransactionReceipt(txHash, c, client)
	if err != nil {
		fmt.Println("Error in call get transaction receipt")
		fmt.Println(err)
		os.Exit(1)
	}

	return nil
}

// maintain implements maintain methods with specified opcode and return the transaction receipt
func maintain(c *cli.Context) error {
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))
	var maintainCmd string
	if c.String("maintaincmd") != "" {
		maintainCmd = c.String("maintaincmd")
	} else {
		invokeParams := []string{"from", "to", "nonce", "payload", "timestamp", "signature", "opcode"}
		method := "contract_maintainContract"
		maintainCmd = getCmd(method, invokeParams, c)
	}
	//fmt.Println(maintainCmd)
	result, err := client.Call(maintainCmd)
	if err != nil {
		fmt.Println("Error in call maintain cmd request")
		fmt.Println(err)
		os.Exit(1)
	}
	//fmt.Print(result.Result)

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
	//TODO: implement destroy cmd
	fmt.Println("Not support yet!")
	return nil
}

func getCmd(method string, deploy_params []string, c *cli.Context) string {
	namespace := c.String("namespace")
	var from, to, payload, invokemethod, arg string
	var nonce, timestamp, amount int64
	var opcode int
	var vmtype types.TransactionValue_VmType
	var err error

	if method == "contract_maintainContract" {
		opcode, err = strconv.Atoi(common.GetNonEmptyValueByName(c, "opcode"))
		if err != nil {
			fmt.Println("Error in convert input opcode to int.")
			fmt.Println(err)
			os.Exit(1)
		}
	}

	params := "[{"
	for i, param := range deploy_params{
		if i > 0 {
			params = params + ","
		}
		switch param {
		case "payload":
			var payload string
			if c.Bool("jvm") {
				if method == "contract_deployContract" || opcode == 1 {
					if c.String("directory") != "" {
						payload = getPayloadFromPath(c.String("directory"))
					} else {
						dir := common.GetNonEmptyValueByName(c, "directory")
						payload = getPayloadFromPath(dir)
					}
				} else {
					payload = ""
				}
			} else {
				payload = common.GetNonEmptyValueByName(c, "payload")
			}
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

		case "opcode":
			params = params + fmt.Sprintf("\"%s\":%d", param, int32(opcode))

		// signature is generated automatically
		case "signature":
			amount = 0
			if c.Bool("jvm") {
				vmtype = 1
			} else {
				vmtype = 0
			}
			sig, err := common.GenSignature(from, to, timestamp, amount, payload, nonce, int32(opcode), vmtype)
			if err != nil {
				fmt.Println("Error in generate signature.")
				fmt.Println(err)
				os.Exit(1)
			}
			signature := hex.EncodeToString(sig)
			params = params + fmt.Sprintf("\"%s\":\"%s\"", param, signature)

		case "method":
			if c.Bool("jvm") {
				invokemethod = common.GetNonEmptyValueByName(c, "method")
				params = params + fmt.Sprintf("\"%s\":\"%s\"", param, invokemethod)
			}

		case "args":
			if c.Bool("jvm") {
				arg = common.GetNonEmptyValueByName(c, "args")
				params = params + fmt.Sprintf("\"%s\":%s", param, arg)
			}

		default:
			fmt.Printf("Invalid param name: %s\n", param)
			os.Exit(1)
		}
	}
	if c.Bool("jvm") {
		params = params + "," + fmt.Sprint("\"type\":\"jvm\"")
	}
	params = params + "}]"

	return fmt.Sprintf(
		"{\"jsonrpc\":\"2.0\",\"namespace\":\"%s\",\"method\":\"%s\",\"params\":%s,\"id\":1}",
		namespace, method, params)
}

func getPayloadFromPath(dir string) string {
	target := "contract.tar.gz"
	common.Compress(dir, target)
	buf, err := ioutil.ReadFile(target)
	if err != nil {
		fmt.Printf("Error in read compressed file: %s\n", target)
		fmt.Println(err.Error())
		os.Exit(1)
	}
	payload := hex.EncodeToString(buf)
	common.DelCompressedFile(target)

	return payload
}

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