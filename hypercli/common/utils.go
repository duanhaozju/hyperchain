//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package common

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/crypto"
	"hyperchain/rpc"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

const (
	tokenpath       = "./.token"
	password        = "123"
	defaultGas      = 10000
	defaultGasPrice = 10000
	frequency       = 10
	hexAddr         = "000f1a7a08ccc48e5d30f80850cf1cf283aa3abd"
)

var (
	kec256Hash = crypto.NewKeccak256Hash("keccak256")
)

// UserInfo records the user's username and token.
type UserInfo struct {
	Username string
	Token    string
}

// GetNonEmptyValueByName first finds the value from cli flags, if not found,
// lets user input from stdin util user inputs an non-empty value.
func GetNonEmptyValueByName(c *cli.Context, name string) string {
	var value string
	if c.String(name) != "" {
		value = c.String(name)
	} else {
		// for username, first find username from token file, if found, return
		// it, else lets user input.
		if name == "username" {
			if user := GetCurrentUser(); user != "" {
				return user
			}
		}

		for {
			if name == "to" {
				name = "contract address"
			}
			fmt.Printf("%s:\n", name)
			fmt.Scanln(&value)
			if value != "" {
				break
			}
		}
	}
	return value
}

// GetJSONResponse returns a JSONResponse from http response result.
func GetJSONResponse(result string) (jsonrpc.JSONResponse, error) {
	var response jsonrpc.JSONResponse
	err := json.Unmarshal([]byte(result), &response)
	if err != nil {
		return response, err
	}
	return response, nil
}

// checkToken checks if given json-format response contains an invalid token error.
func checkToken(result string) error {
	tokenErr := &common.InvalidTokenError{}
	response, err := GetJSONResponse(result)
	if err == nil && response.Code == tokenErr.Code() {
		return errors.New(response.Message)
	}
	return nil
}

// GenSignature generates the transaction signature by many params ...
func GenSignature(from string, to string, timestamp int64, amount int64, payload string, nonce int64, opcode int32, vmtype types.TransactionValue_VmType) ([]byte, error) {
	payload = common.StringToHex(payload)
	txValue := types.NewTransactionValue(int64(defaultGasPrice), int64(defaultGas), amount, common.FromHex(payload), opcode, nil, vmtype)
	value, _ := proto.Marshal(txValue)
	var tx *types.Transaction
	if to == "" {
		tx = types.NewTransaction(common.HexToAddress(from).Bytes(), nil, value, timestamp, nonce)
	} else {
		tx = types.NewTransaction(common.HexToAddress(from).Bytes(), common.HexToAddress(to).Bytes(), value, timestamp, nonce)
	}

	hash := tx.SignHash(kec256Hash).Bytes()
	file := path.Join("./keyconfigs/keystore", hexAddr)
	privKey, err := getKey(file, password)
	if err != nil {
		fmt.Println("Get private key failed!, detail error message: ", err)
		return nil, err
	}

	encryp := crypto.NewEcdsaEncrypto("ecdsa")
	signature, err := encryp.Sign(hash, privKey)
	if err != nil {
		fmt.Println("Sign Transaction failed!, detail error message: ", err)
		return nil, err
	}

	return signature, nil
}

// getTransactionReceiptCmd returns the jsonrpc command of getTransactionReceipt.
func getTransactionReceiptCmd(txHash string, namespace string) string {
	method := "tx_getTransactionReceipt"

	return fmt.Sprintf(
		"{\"jsonrpc\":\"2.0\",\"namespace\":\"%s\",\"method\":\"%s\",\"params\":[\"%s\"],\"id\":1}",
		namespace, method, txHash)
}

// getTransactionReceipt try to get the transaction receipt for 10 times, with 1s interval.
func GetTransactionReceipt(txHash string, namespace string, client *CmdClient) error {
	cmd := getTransactionReceiptCmd(txHash, namespace)
	method := "tx_getTransactionReceipt"

	for i := 1; i <= frequency; i++ {
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

	return fmt.Errorf("cant't get transaction receipt after %v attempts", frequency)
}

// Compress compresses source directory to an archive with '.tar.gz' format.
func Compress(source, target string) {
	// source path must be absolute path, so first convert source path to an absolute path
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

// DelCompressedFile deletes useless archive.
func DelCompressedFile(file string) {
	command := exec.Command("rm", "-rf", file)
	if err := command.Run(); err != nil {
		fmt.Printf("Error in remove compressed file: %s\n", file)
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

// ReadFile reads Userinfo from given file.
func ReadFile(path string, object interface{}) error {
	file, err := os.Open(path)
	defer file.Close()
	if err == nil {
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(object)
	}
	return err
}

// SaveToFile saves username and token to the given file.
func SaveToFile(file, username, token string) error {
	var f *os.File
	if _, err := os.Stat(file); os.IsExist(err) {
		os.Remove(file)

	}
	f, err := os.Create(file)
	if err != nil {
		return err
	}

	defer f.Close()
	userinfo := &UserInfo{Username: username, Token: token}
	encoder := gob.NewEncoder(f)
	return encoder.Encode(userinfo)
}

// GetCurrentUser reads into token file, if existed a username, return it, else
// return a blank string.
func GetCurrentUser() string {
	var username string
	userinfo := new(UserInfo)
	err := ReadFile(tokenpath, userinfo)
	if err == nil {
		username = userinfo.Username
	}
	return username
}
