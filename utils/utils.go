package main

import (
	"flag"
	"fmt"
	"github.com/golang/protobuf/proto"
	"hyperchain/accounts"
	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/crypto"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	encryption = crypto.NewEcdsaEncrypto("ecdsa")
	kec256Hash = crypto.NewKeccak256Hash("keccak256")
)

const (
	keystore           = "./keystore"
	transferUpperLimit = 100
	transferLowerLimit = 0
	defaultGas         = 10000
	defaultGasPrice    = 10000
)

var option = flag.String("o", "", "Two option:(1)account: create account  (2)transaction: create normal transaction or contract transaction")

// account releated
var password = flag.String("p", "", "use to specify account password")

// transaction related
var from = flag.String("from", "", "use to specify transaction sender address")
var to = flag.String("to", "", "use to specify transaction receiver address")
var amount = flag.String("v", "", "use to specify transaction transfer amount")
var payload = flag.String("l", "", "use to specify transaction payload")
var t = flag.Int("t", 0, "use to specify transaction type to generate, o represent normal transaction, 1 represent contract transaction")

func NewAccount(password string) (string, bool) {
	am := accounts.NewAccountManager(keystore, encryption)
	account, err := am.NewAccount(password)
	if err != nil {
		fmt.Println("Create Account Failed! Detail error message: ", err.Error())
		return "", false
	} else {
		fmt.Println("================================ Create Account ========================================")
		fmt.Println("Create new account! Your Account address is:", account.Address.Hex())
		fmt.Println("Your Account password is:", password)
		return account.Address.Hex(), true
	}
}

func NewTransaction() {
	if *password == "" {
		fmt.Println("Please enter your password")
		flag.PrintDefaults()
		return
	}
	var _from string
	var _to string
	var _timestamp int64
	var _amount int64
	var success bool
	if *from != "" {
		_from = *from
	} else {
		_from, success = NewAccount(*password)
		if success == false {
			fmt.Println("Create account failed!")
			return
		}
	}
	_timestamp = time.Now().UnixNano()

	if *amount == "" {
		_amount = generateTransferAmount()
	} else {
		_amount, _ = strconv.ParseInt(*amount, 10, 64)
	}
	if *to == "" {
		_to = generateAddress()
	} else {
		_to = *to
	}
	am := accounts.NewAccountManager(keystore, encryption)
	if *t == 0 {
		txValue := types.NewTransactionValue(int64(defaultGasPrice), int64(defaultGas), _amount, nil)
		value, _ := proto.Marshal(txValue)
		tx := types.NewTransaction(common.HexToAddress(_from).Bytes(), common.HexToAddress(_to).Bytes(), value, _timestamp)
		signature, err := am.SignWithPassphrase(common.BytesToAddress(tx.From), tx.SighHash(kec256Hash).Bytes(), *password)
		if err != nil {
			fmt.Println("Create Transaction failed!, detail error message: ", err)
			return
		}
		fmt.Println("Create Normal Transaction Success!")
		fmt.Println("Arugments:")
		fmt.Println("\tfrom:", _from)
		fmt.Println("\tto:", _to)
		fmt.Println("\ttimestamp:", _timestamp)
		fmt.Println("\tvalue:", _amount)
		fmt.Println("\tsignature:", common.Bytes2Hex(signature))
		fmt.Println("JSONRPC COMMAND:")
		command := fmt.Sprintf("curl localhost:8081 --data '{\"jsonrpc\":\"2.0\",\"method\":\"tx_sendTransaction\",\"params\":[{\"from\":\"%s\",\"to\":\"%s\",\"timestamp\":%d,\"value\":%d,\"signature\":\"%s\"}],\"id\":1}'", _from, _to,_timestamp, _amount, common.Bytes2Hex(signature))
		fmt.Println("\t", command)
	} else {
		_payload := common.StringToHex(*payload)
		txValue := types.NewTransactionValue(int64(defaultGasPrice), int64(defaultGas), 0, common.FromHex(*payload))
		value, _ := proto.Marshal(txValue)
		var tx *types.Transaction
		if *to == "" {
			tx = types.NewTransaction(common.HexToAddress(_from).Bytes(), nil, value, _timestamp)
		} else {
			tx = types.NewTransaction(common.HexToAddress(_from).Bytes(), common.HexToAddress(*to).Bytes(), value, _timestamp)
		}
		signature, err := am.SignWithPassphrase(common.BytesToAddress(tx.From), tx.SighHash(kec256Hash).Bytes(), *password)
		if err != nil {
			fmt.Println("Create Transaction failed!, detail error message: ", err)
			return
		}
		fmt.Println("Create Contract Transaction Success!")
		fmt.Println("ARGUMENTS:")
		fmt.Println("\tfrom:", _from)
		fmt.Println("\tto:", *to)
		fmt.Println("\ttimestamp:", _timestamp)
		fmt.Println("\tpayload:", _payload)
		fmt.Println("\tsignature:", common.Bytes2Hex(signature))
		fmt.Println("JSONRPC COMMAND:")
		if *to == "" {
			command := fmt.Sprintf("curl localhost:8081 --data '{\"jsonrpc\":\"2.0\",\"method\":\"contract_deployContract\",\"params\":[{\"from\":\"%s\",\"timestamp\":%d,\"payload\":\"%s\",\"signature\":\"%s\"}],\"id\":1}'", _from, _timestamp, _payload, common.Bytes2Hex(signature))
			fmt.Println("\t", command)
		} else {
			command := fmt.Sprintf("curl localhost:8081 --data '{\"jsonrpc\":\"2.0\",\"method\":\"contract_invokeContract\",\"params\":[{\"from\":\"%s\", \"to\":\"%s\",\"timestamp\":%d,\"payload\":\"%s\",\"signature\":\"%s\"}],\"id\":1}'", _from, *to, _timestamp, _payload, common.Bytes2Hex(signature))
			fmt.Println("\t", command)
		}
	}
}

func generateAddress() string {
	var letters = []byte("abcdef0123456789")
	b := make([]byte, 40)
	for i := 0; i < 40; i += 1 {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return "0x" + string(b)
}
func generateTransferAmount() int64 {
	return int64(rand.Intn(transferUpperLimit-transferLowerLimit) + transferLowerLimit)
}

func main() {
	flag.Parse()
	if *option == "" {
		flag.PrintDefaults()
	} else if strings.ToLower(*option) == "account" {
		if *password == "" {
			flag.PrintDefaults()
		} else {
			NewAccount(*password)
		}
	} else if strings.ToLower(*option) == "transaction" {
		NewTransaction()
	}
}
