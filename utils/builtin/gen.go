package builtin

import (
	"hyperchain/accounts"
	"flag"
	"time"
	"github.com/golang/protobuf/proto"
	"hyperchain/common"
	"fmt"
	"hyperchain/core/types"
	"math/rand"
)

func NewAccount(password string, silense bool) (string, bool) {
	if password == "" {
		output(silense, "Please enter your password")
		return "", false
	}
	am := accounts.NewAccountManager(keystore, encryption)
	account, err := am.NewAccount(password)
	if err != nil {
		output(silense, "Create Account Failed! Detail error message: ", err.Error())
		return "", false
	} else {
		output(silense, "================================ Create Account ========================================")
		output(silense, "Create new account! Your Account address is:", account.Address.Hex())
		output(silense, "Your Account password is:", password)
		return account.Address.Hex(), true
	}
}
func NewAccounts(count int, password string, silense bool) bool {
	if password == "" {
		password = genesisPassword
	}
	for i := 0; i < count; i += 1 {
		NewAccount(password, silense)
	}
	return true
}
func NewTransaction(password string, from string, to string, timestamp int64, amount int64, payload string, t int, ip string, port int, silense bool, simulate bool) (string, bool) {
	if password == "" {
		output(silense, "Please enter your password")
		if silense == false {
			flag.PrintDefaults()
		}
		return "", false
	}
	var _from string
	var _to string
	var _timestamp int64
	var _amount int64
	var success bool
	if from != "" {
		_from = from
	} else {
		_from, success = NewAccount(password, silense)
		if success == false {
			output(silense, "Create account failed!")
			return "", false
		}
	}
	if timestamp <= 0 {
		_timestamp = time.Now().UnixNano() - timestampRange
	} else {
		_timestamp = timestamp
	}

	if amount < 0 {
		_amount = generateTransferAmount()
	} else {
		_amount = amount
	}
	if to == "" {
		_to = generateAddress()
	} else {
		_to = to
	}
	if t == 0 {
		txValue := types.NewTransactionValue(int64(defaultGasPrice), int64(defaultGas), _amount, nil)
		value, _ := proto.Marshal(txValue)
		tx := types.NewTransaction(common.HexToAddress(_from).Bytes(), common.HexToAddress(_to).Bytes(), value, _timestamp)

		signature, err := am.SignWithPassphrase(common.BytesToAddress(tx.From), tx.SighHash(kec256Hash).Bytes(), password)

		if err != nil {
			output(silense, "Create Transaction failed!, detail error message: ", err)
			return "", false
		}
		output(silense, "Create Normal Transaction Success!")
		output(silense, "Arugments:")
		output(silense, "\tfrom:", _from)
		output(silense, "\tto:", _to)
		output(silense, "\ttimestamp:", _timestamp)
		output(silense, "\tvalue:", _amount)
		output(silense, "\tsignature:", common.Bytes2Hex(signature))
		output(silense, "JSONRPC COMMAND:")
		var command string
		if simulate == true {
			command = fmt.Sprintf("curl %s:%d --data '{\"jsonrpc\":\"2.0\",\"method\":\"tx_sendTransaction\",\"params\":[{\"from\":\"%s\",\"to\":\"%s\",\"timestamp\":%d,\"value\":%d,\"signature\":\"%s\",\"simulate\":true}],\"id\":1}'", ip, port,_from, _to, _timestamp, _amount, common.Bytes2Hex(signature))
		} else {
			command = fmt.Sprintf("curl %s:%d --data '{\"jsonrpc\":\"2.0\",\"method\":\"tx_sendTransaction\",\"params\":[{\"from\":\"%s\",\"to\":\"%s\",\"timestamp\":%d,\"value\":%d,\"signature\":\"%s\"}],\"id\":1}'", ip, port,_from, _to, _timestamp, _amount, common.Bytes2Hex(signature))
		}
		output(silense, "\t", command)
		return command, true
	} else {
		_payload := common.StringToHex(payload)
		txValue := types.NewTransactionValue(int64(defaultGasPrice), int64(defaultGas), 0, common.FromHex(payload))
		value, _ := proto.Marshal(txValue)
		var tx *types.Transaction
		if to == "" {
			tx = types.NewTransaction(common.HexToAddress(_from).Bytes(), nil, value, _timestamp)
		} else {
			tx = types.NewTransaction(common.HexToAddress(_from).Bytes(), common.HexToAddress(to).Bytes(), value, _timestamp)
		}
		signature, err := am.SignWithPassphrase(common.BytesToAddress(tx.From), tx.SighHash(kec256Hash).Bytes(), password)
		if err != nil {
			output(silense, "Create Transaction failed!, detail error message: ", err)
			return "", false
		}
		output(silense, "Create Contract Transaction Success!")
		output(silense, "ARGUMENTS:")
		output(silense, "\tfrom:", _from)
		output(silense, "\tto:", to)
		output(silense, "\ttimestamp:", _timestamp)
		output(silense, "\tpayload:", _payload)
		output(silense, "\tsignature:", common.Bytes2Hex(signature))
		output(silense, "JSONRPC COMMAND:")
		if to == "" {
			var command string
			if simulate {
				command = fmt.Sprintf("curl %s:%d --data '{\"jsonrpc\":\"2.0\",\"method\":\"contract_deployContract\",\"params\":[{\"from\":\"%s\",\"timestamp\":%d,\"payload\":\"%s\",\"signature\":\"%s\",\"simulate\":true}],\"id\":1}'", ip, port,_from, _timestamp, _payload, common.Bytes2Hex(signature))
			} else {
				command = fmt.Sprintf("curl %s:%d --data '{\"jsonrpc\":\"2.0\",\"method\":\"contract_deployContract\",\"params\":[{\"from\":\"%s\",\"timestamp\":%d,\"payload\":\"%s\",\"signature\":\"%s\"}],\"id\":1}'", ip, port,_from, _timestamp, _payload, common.Bytes2Hex(signature))
			}
			output(silense, "\t", command)
			return command, true
		} else {
			var command string
			if simulate {
				command = fmt.Sprintf("curl %s:%d --data '{\"jsonrpc\":\"2.0\",\"method\":\"contract_invokeContract\",\"params\":[{\"from\":\"%s\", \"to\":\"%s\",\"timestamp\":%d,\"payload\":\"%s\",\"signature\":\"%s\",\"simulate\":true}],\"id\":1}'", ip, port,_from, to, _timestamp, _payload, common.Bytes2Hex(signature))
			} else {
				command = fmt.Sprintf("curl %s:%d --data '{\"jsonrpc\":\"2.0\",\"method\":\"contract_invokeContract\",\"params\":[{\"from\":\"%s\", \"to\":\"%s\",\"timestamp\":%d,\"payload\":\"%s\",\"signature\":\"%s\"}],\"id\":1}'", ip, port,_from, to, _timestamp, _payload, common.Bytes2Hex(signature))
			}
			output(silense, "\t", command)
			return command, true
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

func output(silense bool, msg ...interface{}) {
	if !silense {
		fmt.Println(msg...)
	}
}
