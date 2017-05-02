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
	"math/rand"
	"time"
	"os"
	"regexp"
	"strings"
	"path/filepath"
)

const frequency = 10

//NewContractCMD new contract related commands.
func NewContractCMD() []cli.Command {
	return []cli.Command{
		{
			Name:    "deploy",
			Aliases: []string{"d"},
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
					Name:  "path, p",
					Value: "",
					Usage: "specify the contract file path",
				},
				cli.StringFlag{
					Name:  "namespace, n",
					Value: "global",
					Usage: "specify the namespace to deploy to, default is global",
				},
				cli.StringFlag{
					Name:  "from, f",
					Value: "17d806c92fa941b4b7a8ffffc58fa2f297a3bffc",
					Usage: "specify the deploy account",
				},
				cli.StringFlag{
					Name:  "signature, s",
					Value: "0x19c0655d05b9c24f5567846528b81a25c48458a05f69f05cf8d6c46894b9f12a02af471031ba11f155e41adf42fca639b67fb7148ddec90e7628ec8af60c872c00",
					Usage: "specify the signature",
				},
			},
		},
		{
			Name:    "invoke",
			Aliases: []string{"i"},
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
					Name:  "path, p",
					Value: "",
					Usage: "specify the contract file path",
				},
				cli.StringFlag{
					Name:  "namespace, n",
					Value: "global",
					Usage: "specify the namespace to deploy to, default is global",
				},
				cli.StringFlag{
					Name:  "from, f",
					Value: "17d806c92fa941b4b7a8ffffc58fa2f297a3bffc",
					Usage: "specify the deploy account",
				},
				cli.StringFlag{
					Name:  "signature, s",
					Value: "0x19c0655d05b9c24f5567846528b81a25c48458a05f69f05cf8d6c46894b9f12a02af471031ba11f155e41adf42fca639b67fb7148ddec90e7628ec8af60c872c00",
					Usage: "specify the signature",
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
	//fmt.Println(deployCmd)
	result, err := client.Call(deployCmd)
	if err != nil {
		fmt.Println("Error in call deploy cmd request")
		fmt.Print(err)
		os.Exit(1)
	}

	txHash := getTransactionHash(result.Result)
	method := "tx_getTransactionReceipt"
	gtrCmd := getTransactionReceiptCmd(method, txHash, c)
	//fmt.Println(gtrCmd)
	err = getTransactionReceipt(client, gtrCmd)
	if err != nil {
		fmt.Println("Error in call get transaction receipt")
		fmt.Println(err)
		os.Exit(1)
	}

	return nil
}

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
	fmt.Println(invokeCmd)
	client.Call(invokeCmd)

	return nil
}

func destroy(c *cli.Context) error {
	//TODO: implement destroy cmd
	return nil
}

func getCmd(method string, deploy_params []string, c *cli.Context) string {
	namespace := c.String("namespace")

	params := "[{"
	for i, param := range deploy_params{
		if i > 0 {
			params = params + ","
		}
		switch param {
		case "payload":
			var payload string
			if c.Bool("jvm") {
				if method == "contract_invokeContract" {
					payload = ""
				} else {
					if c.String("path") != "" {
						payload = getPayloadFromPath(c.String("path"))
					} else {
						for {
							var path string
							fmt.Println("Please specify a non-empty contarct path:")
							fmt.Scanln(&path)
							if path != "" {
								payload = getPayloadFromPath(path)
								break
							}
						}
					}
				}
			} else {
				for {
					fmt.Println("Please specify a non-empty contarct payload:")
					fmt.Scanln(&payload)
					if payload != "" {
						break
					}
				}
			}
			params = params + fmt.Sprintf("\"%s\":\"%s\"", param, payload)

		case "nonce":
			nonce := rand.Int63()
			params = params + fmt.Sprintf("\"%s\":%d", param, nonce)

		case "timestamp":
			timestamp := time.Now().UnixNano()
			params = params + fmt.Sprintf("\"%s\":%d", param, timestamp)

		//TODO generate from, signature automatically
		case "from":
			params = params + fmt.Sprintf("\"%s\":\"%s\"", param, c.String("from"))

		case "signature":
			params = params + fmt.Sprintf("\"%s\":\"%s\"", param, c.String("signature"))

		//below params must be input by user
		case "to":
			var to string
			if c.String("to") != "" {
				to = c.String("to")
			} else {
				for {
					fmt.Println("Please specify a non-empty contractAddress:")
					fmt.Scanln(&to)
					if to != "" {
						break
					}
				}
			}
			params = params + fmt.Sprintf("\"%s\":\"%s\"", param, to)

		case "method":
			var method string
			if c.String("method") != "" {
				method = c.String("method")
			} else {
				for {
					fmt.Println("Please specify a non-empty invoke method:")
					fmt.Scanln(&method)
					if method != "" {
						break
					}
				}
			}
			params = params + fmt.Sprintf("\"%s\":\"%s\"", param, method)

		case "args":
			var arg string
			if c.String("args") != "" {
				arg = c.String("args")
			} else {
				for {
					fmt.Println("Please specify a non-empty invoke args:")
					fmt.Scanln(&arg)
					if arg != "" {
						break
					}
				}
			}
			params = params + fmt.Sprintf("\"%s\":%s", param, arg)

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

func getPayloadFromPath (path string) string {
	target := "contract.tar.gz"
	compress(path, target)
	buf, err := ioutil.ReadFile(target)
	if err != nil {
		fmt.Printf("Error in read compressed file: %s\n", target)
		fmt.Println(err.Error())
		os.Exit(1)
	}
	payload := hex.EncodeToString(buf)
	delCompressedFile(target)

	return payload
}

func compress(source, target string) {
	//source path must be absolute path, so first convert source path to an absolute path
	abs, err := filepath.Abs(source)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	dir, file := filepath.Split(abs)
	// compress files to a tar.gz with only one-level path
	command := exec.Command("tar", "-czf", target, "-C", dir, file)
	if err := command.Run(); err != nil {
		fmt.Printf("Error in read compress specefied file: %s\n", source)
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func delCompressedFile(file string) {
	command := exec.Command("rm", "-rf", file)
	if err := command.Run(); err != nil {
		fmt.Printf("Error in remove compressed file: %s\n", file)
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func getTransactionHash (result interface{}) string {
	if jrp, ok := result.(string); ok {
		pat := `"result"\:".+"`
		reg, err := regexp.Compile(pat)
		if (err != nil ) {
			fmt.Println("Error in compile regular expression")
			fmt.Println(err)
			os.Exit(1)
		}
		strArr := reg.FindAllString(jrp, -1)
		str := strArr[0][10:]
		str = strings.TrimSuffix(str, "\"")
		return str
	} else {
		fmt.Println("Error in assert interface to string !")
		os.Exit(1)
		return ""
	}
}

func getTransactionReceiptCmd(method string, txHash string, c *cli.Context) string {
	namespace := c.String("namespace")

	return fmt.Sprintf(
		"{\"jsonrpc\":\"2.0\",\"namespace\":\"%s\",\"method\":\"%s\",\"params\":[\"%s\"],\"id\":1}",
		namespace, method, txHash)
}

func getTransactionReceipt(client *common.CmdClient, cmd string) error {
	for i:= 1; i<= frequency; i ++ {
		response, err := client.Call(cmd)
		if err != nil {
			return err
		} else {
			if result, ok := response.Result.(string); !ok {
				return fmt.Errorf("Error in assert interface to string !")
			} else {
				if strings.Contains(result, "SUCCESS") {
					fmt.Print(result)
					return nil
				}
			}
		}
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("Cant't get transaction receipt after %v attempts", frequency)
}