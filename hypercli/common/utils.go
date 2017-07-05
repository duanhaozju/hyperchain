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
	"hyperchain/api/jsonrpc/core"
	"strings"
	"time"
	"path/filepath"
	"os"
	"os/exec"
	"io/ioutil"
	"bufio"
)

const (
	tokenpath       = "./.token"
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
func GetJSONResponse(result string) (jsonrpc.JSONResponse, error) {
	var response jsonrpc.JSONResponse
	err := json.Unmarshal([]byte(result), &response)
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
func getTransactionReceiptCmd(txHash string, namespace string) string {
	method := "tx_getTransactionReceipt"

	return fmt.Sprintf(
		"{\"jsonrpc\":\"2.0\",\"namespace\":\"%s\",\"method\":\"%s\",\"params\":[\"%s\"],\"id\":1}",
		namespace, method, txHash)
}

// getTransactionReceipt try to get the transaction receipt 10 times, with 1s interval
func GetTransactionReceipt(txHash string, namespace string, client *CmdClient) error {
	cmd := getTransactionReceiptCmd(txHash, namespace)
	method := "tx_getTransactionReceipt"

	for i:= 1; i<= frequency; i ++ {
		response, err := client.Call(cmd, method)
		if err != nil {
			return err
		} else {
			if strings.Contains(response, "SUCCESS") {
				fmt.Print(response)
				return nil
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

func ReadFile(file string) (string, error) {
	token, err := ioutil.ReadFile(file)
	return string(token[:]), err
}

func SaveToFile(file, token string) error {
	var f *os.File
	if _, err := os.Stat(file); os.IsExist(err) {
		os.Remove(file)

	}
	f, err := os.Create(file)
	if err != nil {
		return err
	}

	defer f.Close()
	f.WriteString(token)
	return nil
}

func ReadPermissionsFromFile(file string) ([]string, error) {
	var permissions []string
	abs, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}

	// open a file
	if file, err := os.Open(abs); err == nil {
		// make sure it gets closed
		defer file.Close()

		// create a new scanner and read the file line by line
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			permission := scanner.Text()
			if permission != "" {
				permissions = append(permissions, permission)
			}
		}

		// check for errors
		if err = scanner.Err(); err != nil {
			return nil, err
		}
		return permissions, nil
	} else {
		return nil, err
	}
	return nil, nil
}
