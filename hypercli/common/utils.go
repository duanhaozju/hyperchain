//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package common

import (
	"fmt"
	"github.com/urfave/cli"
	"hyperchain/core/types"
	"github.com/golang/protobuf/proto"
	"hyperchain/common"
	"hyperchain/accounts"
	"hyperchain/crypto"
	"encoding/json"
	"hyperchain/rpc"
	"strings"
	"errors"
	"time"
	"path/filepath"
	"os"
	"os/exec"
)

const (
	password        = "123"
	defaultGas      = 10000
	defaultGasPrice = 10000
	frequency       = 10
)

var (
	kec256Hash       = crypto.NewKeccak256Hash("keccak256")
)

// GetNonEmptyValueByName first finds the value from cli flags, if not found,
// lets user input from stdin util user inputs an non-empty value
func GetNonEmptyValueByName(c *cli.Context, name string) string {
	var value string
	if c.String(name) != "" {
		value = c.String(name)
	} else {
		for {
			if name == "to" {
				name = "contract address"
			}
			fmt.Printf("Please specify a non-empty %s:\n", name)
			fmt.Scanln(&value)
			if value != "" {
				break
			}
		}
	}
	return value
}

// GetJSONResponse returns a JSONResponse from http response result
func GetJSONResponse(result *jsonrpc.CommandResult) (jsonrpc.JSONResponse, error) {
	var response jsonrpc.JSONResponse
	err := json.Unmarshal([]byte(result.Result.(string)), &response)
	if err != nil {
		return response, err
	}
	return response, nil
}

// GenSignature generates the transaction signature by many params ...
func GenSignature(from string, to string, timestamp int64, amount int64, payload string, nonce int64, opcode int32, vmtype types.TransactionValue_VmType) ([]byte, error){
	conf := common.NewConfig("./keyconfigs/cli.yaml")
	conf.Set(common.C_NODE_ID, 1)

	am := accounts.NewAccountManager(conf)

	payload = common.StringToHex(payload)
	txValue := types.NewTransactionValue(int64(defaultGasPrice), int64(defaultGas), amount, common.FromHex(payload), opcode, vmtype)
	value, _ := proto.Marshal(txValue)
	var tx *types.Transaction
	if to == "" {
		tx = types.NewTransaction(common.HexToAddress(from).Bytes(), nil, value, timestamp, nonce)
	} else {
		tx = types.NewTransaction(common.HexToAddress(from).Bytes(), common.HexToAddress(to).Bytes(), value, timestamp, nonce)
	}

	signature, err := am.SignWithPassphrase(common.BytesToAddress(tx.From), tx.SighHash(kec256Hash).Bytes(), password)

	if err != nil {
		fmt.Println("Sign Transaction failed!, detail error message: ", err)
		return nil, err
	}
	return signature, nil
}

// getTransactionReceiptCmd returns the jsonrpc command of getTransactionReceipt
func getTransactionReceiptCmd(txHash string, c *cli.Context) string {
	method := "tx_getTransactionReceipt"
	namespace := c.String("namespace")

	return fmt.Sprintf(
		"{\"jsonrpc\":\"2.0\",\"namespace\":\"%s\",\"method\":\"%s\",\"params\":[\"%s\"],\"id\":1}",
		namespace, method, txHash)
}

// getTransactionReceipt try to get the transaction receipt 10 times, with 1s interval
func GetTransactionReceipt(txHash string, c *cli.Context, client *CmdClient) error {
	cmd := getTransactionReceiptCmd(txHash, c)

	for i:= 1; i<= frequency; i ++ {
		response, err := client.Call(cmd)
		if err != nil {
			return err
		} else {
			if result, ok := response.Result.(string); !ok {
				return errors.New("Error in assert interface to string !")
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

func Compress(source, target string) {
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

func DelCompressedFile(file string) {
	command := exec.Command("rm", "-rf", file)
	if err := command.Run(); err != nil {
		fmt.Printf("Error in remove compressed file: %s\n", file)
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
