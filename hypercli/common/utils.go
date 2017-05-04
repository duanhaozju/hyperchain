package common

import (
	"fmt"
	"github.com/urfave/cli"
	"hyperchain/core/types"
	"github.com/golang/protobuf/proto"
	"hyperchain/common"
	"hyperchain/accounts"
	"hyperchain/crypto"
)

const (
	password        = "123"
	defaultGas      = 10000
	defaultGasPrice = 10000
)

var (
	kec256Hash       = crypto.NewKeccak256Hash("keccak256")
)

func GetNonEmptyValueByName(c *cli.Context, name string) string {
	var value string
	if c.String(name) != "" {
		value = c.String(name)
	} else {
		for {
			fmt.Printf("Please specify a non-empty %s:\n", name)
			fmt.Scanln(&value)
			if value != "" {
				break
			}
		}
	}
	return value
}

// GenSignature generates the transaction signature by many params ...
func GenSignature(from string, to string, timestamp int64, amount int64, payload string, nonce int64, opcode int32) ([]byte, error){
	conf := common.NewConfig("./cli.yaml")
	conf.Set(common.KEY_STORE_DIR, "./keystore")
	conf.Set(common.KEY_NODE_DIR, "./keynodes")
	conf.Set(common.C_NODE_ID, 1)

	am := accounts.NewAccountManager(conf)

	payload = common.StringToHex(payload)
	txValue := types.NewTransactionValue(int64(defaultGasPrice), int64(defaultGas), amount, common.FromHex(payload), opcode)
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