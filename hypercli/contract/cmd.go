//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package contract

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"

	"hyperchain/core/types"
	"hyperchain/hypercli/common"

	"github.com/golang/protobuf/proto"
	"github.com/urfave/cli"
)

// commonFlags defines some common flags used in contract cmd.
var commonFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "jvm, j",
		Usage: "specify how the contract is generated, false is evm, true is jvm",
	},
	cli.StringFlag{
		Name:  "namespace, n",
		Value: "global",
		Usage: "specify the namespace, default is global",
	},
	cli.StringFlag{
		Name:  "from, f",
		Value: "000f1a7a08ccc48e5d30f80850cf1cf283aa3abd",
		Usage: "specify the account",
	},
	cli.StringFlag{
		Name:  "payload, p",
		Value: "",
		Usage: "specify the contract payload",
	},
	cli.StringFlag{
		Name:  "extra, e",
		Value: "",
		Usage: "specify the extra information",
	},
}

// NewContractCMD new contract related commands.
func NewContractCMD() []cli.Command {
	return []cli.Command{
		{
			Name:   "deploy",
			Usage:  "Deploy a contract",
			Action: deploy,
			Flags: append(commonFlags, []cli.Flag{
				cli.StringFlag{
					Name:  "deploycmd, c",
					Value: "",
					Usage: "specify the payload of deploy contract",
				},
				cli.StringFlag{
					Name:  "directory, d",
					Value: "",
					Usage: "specify the contract file directory",
				},
			}...),
		},
		{
			Name:   "invoke",
			Usage:  "Invoke a contract",
			Action: invoke,
			Flags: append(commonFlags, []cli.Flag{
				cli.StringFlag{
					Name:  "invokecmd, c",
					Value: "",
					Usage: "specify the payload of invoke contract",
				},

				//args with no default value which must be specified by user
				cli.StringFlag{
					Name:  "to, t",
					Value: "",
					Usage: "specify the contract address",
				},
				cli.StringFlag{
					Name:  "args, a",
					Value: "",
					Usage: "specify the args of invoke contract",
				},
			}...),
		},
		{
			Name:   "update",
			Usage:  "Update a contract",
			Action: update,
			Flags: append(commonFlags, []cli.Flag{
				cli.StringFlag{
					Name:  "updatecmd, c",
					Value: "",
					Usage: "specify the payload of update contract",
				},
				cli.StringFlag{
					Name:  "to, t",
					Value: "",
					Usage: "specify the contract address",
				},
				cli.StringFlag{
					Name:  "directory, d",
					Value: "",
					Usage: "specify the contract file directory",
				},
			}...),
		},
		{
			Name:   "frozen",
			Usage:  "Frozen a contract",
			Action: frozen,
			Flags: append(commonFlags, []cli.Flag{
				cli.StringFlag{
					Name:  "frozencmd, c",
					Value: "",
					Usage: "specify the payload of frozen contract",
				},
				cli.StringFlag{
					Name:  "to, t",
					Value: "",
					Usage: "specify the contract address",
				},
			}...),
		},
		{
			Name:   "unfrozen",
			Usage:  "Unfrozen a contract",
			Action: unfrozen,
			Flags: append(commonFlags, []cli.Flag{
				cli.StringFlag{
					Name:  "unfrozencmd, c",
					Value: "",
					Usage: "specify the payload of unfrozen contract",
				},
				cli.StringFlag{
					Name:  "to, t",
					Value: "",
					Usage: "specify the contract address",
				},
			}...),
		},
		{
			Name:   "destroy",
			Usage:  "Destroy a contract",
			Action: destroy,
			Flags: append(commonFlags, []cli.Flag{
				cli.StringFlag{
					Name:  "destroycmd, c",
					Value: "",
					Usage: "specify the payload of destory contract",
				},
				cli.StringFlag{
					Name:  "to, t",
					Value: "",
					Usage: "specify the contract address",
				},
			}...),
		},
	}
}

// deploy implements deploy contract and prints the transaction receipt.
func deploy(c *cli.Context) error {
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))
	var deployCmd string
	method := "contract_deployContract"
	if c.String("deploycmd") != "" {
		deployCmd = c.String("deploycmd")
	} else {
		deployParams := []string{"from", "payload"}
		deployCmd = getCmd(method, deployParams, 0, c)
	}

	result, err := client.Call(deployCmd, method)
	if err != nil {
		fmt.Println("Error in call deploy cmd request")
		fmt.Println(err)
		os.Exit(1)
	}

	txHash := getTransactionHash(result)
	err = common.GetTransactionReceipt(txHash, c.String("namespace"), client)
	if err != nil {
		fmt.Println("Error in call get transaction receipt")
		fmt.Println(err)
		os.Exit(1)
	}

	return nil
}

// invoke implements invoke contract and prints the transaction receipt.
func invoke(c *cli.Context) error {
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))
	var invokeCmd string
	method := "contract_invokeContract"
	if c.String("invokecmd") != "" {
		invokeCmd = c.String("invokecmd")
	} else {
		invokeParams := []string{"from", "to", "payload", "args"}
		invokeCmd = getCmd(method, invokeParams, 0, c)
	}

	result, err := client.Call(invokeCmd, method)
	if err != nil {
		fmt.Println("Error in call invoke cmd request")
		fmt.Println(err)
		os.Exit(1)
	}

	txHash := getTransactionHash(result)
	err = common.GetTransactionReceipt(txHash, c.String("namespace"), client)
	if err != nil {
		fmt.Println("Error in call get transaction receipt")
		fmt.Println(err)
		os.Exit(1)
	}

	return nil
}

// update updates the contract.
func update(c *cli.Context) error {
	if err := maintain(c, 1, "updatecmd"); err != nil {
		fmt.Println("Error in update contract!")
		fmt.Println(err)
		os.Exit(1)
	}
	return nil
}

// frozen frozen the contract.
func frozen(c *cli.Context) error {
	if err := maintain(c, 2, "frozencmd"); err != nil {
		fmt.Println("Error in frozen contract!")
		fmt.Println(err)
		os.Exit(1)
	}
	return nil
}

// unfrozen unfrozen the contract.
func unfrozen(c *cli.Context) error {
	if err := maintain(c, 3, "unfrozencmd"); err != nil {
		fmt.Println("Error in unfrozen contract!")
		fmt.Println(err)
		os.Exit(1)
	}
	return nil
}

// destroy destroys the contract.
func destroy(c *cli.Context) error {
	if err := maintain(c, 4, "destroycmd"); err != nil {
		fmt.Println("Error in destroy contract!")
		fmt.Println(err)
		os.Exit(1)
	}
	return nil
}

// maintain implements maintain methods with specified opcode and prints the transaction receipt.
func maintain(c *cli.Context, opcode int32, maintainMethod string) error {
	client := common.NewRpcClient(c.GlobalString("host"), c.GlobalString("port"))
	var maintainCmd string
	method := "contract_maintainContract"
	if c.String(maintainMethod) != "" {
		maintainCmd = c.String(maintainMethod)
	} else {
		maintainParams := []string{"from", "to", "payload", "opcode"}
		maintainCmd = getCmd(method, maintainParams, opcode, c)
	}

	result, err := client.Call(maintainCmd, method)
	if err != nil {
		fmt.Printf("Error in call %s request\n", maintainCmd)
		return err
	}

	txHash := getTransactionHash(result)
	err = common.GetTransactionReceipt(txHash, c.String("namespace"), client)
	if err != nil {
		fmt.Println("Error in call get transaction receipt")
		return err
	}

	return nil
}

// getCmd returns the jsonrpc string with the given params...
func getCmd(method string, need_params []string, opcode int32, c *cli.Context) string {
	namespace := c.String("namespace")
	var from, to, invokemethod, arg string
	var nonce, timestamp, amount int64
	var vmtype types.TransactionValue_VmType
	var code []byte
	var args [][]byte

	if method == "contract_invokeContract" {
		invokemethod = "invoke"
	}

	params := "[{"
	for i, param := range need_params {
		if i > 0 && param != "payload" && param != "method" && param != "args" {
			params = params + ","
		}
		switch param {
		// For evm contract, user must directly input the complete contract payload
		// For jvm contract, user must input a contract directory which can be used
		// to generate jvm contract payload.
		case "payload":
			if c.Bool("jvm") {
				if method == "contract_deployContract" || opcode == 1 {
					if c.String("directory") != "" {
						code = getPayloadFromPath(c.String("directory"))
					} else {
						dir := common.GetNonEmptyValueByName(c, "directory")
						code = getPayloadFromPath(dir)
					}
				} else {
					code = []byte{}
				}
			} else {
				code = []byte(common.GetNonEmptyValueByName(c, "payload"))
			}

		// from is the sender of the contract, user can input a useful account
		// address, or we will use the default account.
		case "from":
			from = c.String("from")
			params = params + fmt.Sprintf("\"%s\":\"%s\"", param, from)

		// below params must be input by user.

		// to specifies the contract address we will invoke to.
		case "to":
			to = common.GetNonEmptyValueByName(c, "to")
			params = params + fmt.Sprintf("\"%s\":\"%s\"", param, to)

		// opcode is the identification of contract maintain:
		// 1 for update
		// 2 for frozen
		// 3 for unfrozen
		case "opcode":
			params = params + fmt.Sprintf("\"%s\":%d", param, opcode)

		// args is only used in contract invoke which specifies the invoke args.
		case "args":
			if c.Bool("jvm") {
				arg = common.GetNonEmptyValueByName(c, "args")
				tmp := strings.Fields(arg)
				for _, a := range tmp {
					args = append(args, []byte(a))
				}
			}

		default:
			fmt.Printf("Invalid param name: %s\n", param)
			os.Exit(1)
		}
	}
	// jvm contract
	if c.Bool("jvm") {
		params = params + "," + fmt.Sprint("\"type\":\"jvm\"")
	}

	// generate nonce
	nonce = rand.Int63()
	params = params + "," + fmt.Sprintf("\"nonce\":%d", nonce)

	// generate timestamp
	timestamp = time.Now().UnixNano()
	params = params + "," + fmt.Sprintf("\"timestamp\":%d", timestamp)

	// generate payload
	payload := &types.InvokeArgs{
		Code:       code,
		MethodName: invokemethod,
		Args:       args,
	}
	invokeArgs, err := proto.Marshal(payload)
	if err != nil {
		fmt.Println("Marsh error: ", err.Error())
	}
	params = params + "," + fmt.Sprintf("\"payload\":\"%s\"", hex.EncodeToString(invokeArgs))

	// generate signature
	amount = 0
	if c.Bool("jvm") {
		vmtype = 1
	} else {
		vmtype = 0
	}
	sig, err := common.GenSignature(from, to, timestamp, amount, hex.EncodeToString(invokeArgs), nonce, opcode, vmtype)
	if err != nil {
		fmt.Println("Error in generate signature.")
		fmt.Println(err)
		os.Exit(1)
	}
	signature := "00" + hex.EncodeToString(sig)
	params = params + "," + fmt.Sprintf("\"signature\":\"%s\"", signature)

	// extra string
	if c.String("extra") != "" {
		params = params + "," + fmt.Sprintf("\"extra\":\"%s\"", c.String("extra"))
	}

	// end of params
	params = params + "}]"

	cmd := fmt.Sprintf(
		"{\"jsonrpc\":\"2.0\",\"namespace\":\"%s\",\"method\":\"%s\",\"params\":%s,\"id\":1}",
		namespace, method, params)
	return cmd
}

// getPayloadFromPath firsts compresses the given dir, and then reads into
// the archive to []byte.
func getPayloadFromPath(dir string) []byte {
	target := "contract.tar.gz"
	common.Compress(dir, target)
	buf, err := ioutil.ReadFile(target)
	if err != nil {
		fmt.Printf("Error in read compressed file: %s\n", target)
		fmt.Println(err.Error())
		os.Exit(1)
	}
	common.DelCompressedFile(target)

	return buf
}

// getTransactionHash returns the transaction hash using the given http response.
func getTransactionHash(result string) string {
	response, err := common.GetJSONResponse(result)
	if err != nil {
		fmt.Println("Error in call get transaction hash from http response")
		fmt.Println(err)
		os.Exit(1)
		return ""
	}

	if hash, ok := response.Result.(string); !ok {
		fmt.Println("Error in call get transaction hash from http response")
		fmt.Printf("rpc result: %v can't convert to string\n", response.Result)
		return ""
	} else {
		return hash
	}
}
