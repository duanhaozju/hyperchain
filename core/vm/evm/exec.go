// Copyright 2016-2017 Hyperchain Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package evm

import (
	"bytes"
	"github.com/hyperchain/hyperchain/common"
	er "github.com/hyperchain/hyperchain/core/errors"
	"github.com/hyperchain/hyperchain/core/types"
	"github.com/hyperchain/hyperchain/core/vm"
	"github.com/op/go-logging"
	"math/big"
)

const (
	RC1_2_TXGASUPPERLIMIT = 100000000
)

type Message struct {
	From     common.Address
	To       common.Address
	Gas      *big.Int
	GasPrice *big.Int
	Amount   *big.Int
	Payload  []byte
	Op       types.TransactionValue_Opcode
}

func ExecTransaction(db vm.Database, tx *types.Transaction, idx int, blockNumber uint64, logger *logging.Logger, namespace string) (receipt *types.Receipt, ret []byte, addr common.Address, err error) {
	message := setDefaults(tx)
	env := initEnvironment(db, blockNumber, logger, namespace, tx.GetHash())
	env.Db().StartRecord(tx.GetHash(), common.Hash{}, idx)
	if valid := checkPermission(env, message.From, message.To, message.Op); !valid {
		return nil, nil, common.Address{}, er.InvalidInvokePermissionErr("not enough permission to invocation")
	}

	preGas := big.NewInt(0).Set(message.Gas)

	if tx.To == nil {
		ret, addr, err = Exec(env, &message.From, nil, message.Payload, message.Gas,
			message.GasPrice, message.Amount, message.Op)
	} else {
		ret, _, err = Exec(env, &message.From, &message.To, message.Payload, message.Gas,
			message.GasPrice, message.Amount, message.Op)
	}

	receipt = makeReceipt(env, addr, tx.GetHash(), message.Gas, big.NewInt(0).Sub(preGas, message.Gas), ret, err)
	return receipt, ret, addr, err
}

func Exec(vmenv vm.Environment, from, to *common.Address, data []byte, gas,
	gasPrice, value *big.Int, op types.TransactionValue_Opcode) (ret []byte, addr common.Address, err error) {
	var sender vm.Account

	if !(vmenv.Db().Exist(*from)) {
		sender = vmenv.Db().CreateAccount(*from)
	} else {
		sender = vmenv.Db().GetAccount(*from)
	}

	contractCreation := (nil == to)

	if contractCreation {
		ret, addr, err = vmenv.Create(sender, data, gas, gasPrice, value)
		if err != nil {
			ret = nil
			vmenv.Logger().Error("VM create err:", err)
		}
	} else {
		ret, err = vmenv.Call(sender, *to, data, gas, gasPrice, value, int32(op))
		if err != nil {
			ret = nil
			vmenv.Logger().Error("VM call err:", err)
		}
	}
	return ret, addr, err
}

// checkPermission make sure the caller is the contract creator if the opcode is special.
func checkPermission(env vm.Environment, from, to common.Address, op types.TransactionValue_Opcode) bool {
	if op == types.TransactionValue_UPDATE || op == types.TransactionValue_FREEZE || op == types.TransactionValue_UNFREEZE {
		env.Logger().Debugf("caller address %s", from.Hex())
		env.Logger().Debugf("callee address %s", from.Hex())
		if bytes.Compare(to.Bytes(), nil) == 0 {
			return false
		}
		if bytes.Compare(from.Bytes(), env.Db().GetCreator(to).Bytes()) != 0 {
			env.Logger().Errorf("only contract owner %s can do `update`, `freeze` or `unfreeze` operations. %s doesn't has enough permission",
				env.Db().GetCreator(to).Hex(), from.Hex())
			return false
		}
	}
	return true
}

// makeReceipt encapsulates execution result to a receipt.
func makeReceipt(env vm.Environment, addr common.Address, txHash common.Hash, gasRemained, gas *big.Int, ret []byte, err error) *types.Receipt {
	receipt := types.NewReceipt(gasRemained, 0)
	receipt.ContractAddress = addr.Bytes()
	receipt.TxHash = txHash.Bytes()
	receipt.GasUsed = gas.Int64()
	receipt.Ret = ret
	receipt.SetLogs(env.Db().GetLogs(common.BytesToHash(receipt.TxHash)))
	bloom, _ := types.CreateBloom([]*types.Receipt{receipt})
	receipt.Bloom = bloom
	if err != nil {
		if !er.IsValueTransferErr(err) && !er.IsExecContractErr(err) && !er.IsInvalidInvokePermissionErr(err) {
			receipt.Status = types.Receipt_FAILED
			receipt.Message = []byte(err.Error())
		}
	} else {
		receipt.Status = types.Receipt_SUCCESS
		receipt.Message = nil
	}
	return receipt
}

// initEnvironment creates a transaction environment with given env info.
func initEnvironment(state vm.Database, seqNo uint64, logger *logging.Logger, namespace string, txHash common.Hash) vm.Environment {
	// TODO pass block package time to env
	vmenv := NewEnv(state, int64(seqNo), 0, logger, namespace, txHash)
	return vmenv
}

// setDefaults for backward compatibility.
func setDefaults(tx *types.Transaction) Message {
	tv := tx.GetTransactionValue()
	switch string(tx.Version) {
	case "1.0":
		fallthrough
	case "1.1":
		fallthrough
	case "1.2":
		return Message{
			From:     common.BytesToAddress(tx.From),
			To:       common.BytesToAddress(tx.To),
			Gas:      big.NewInt(RC1_2_TXGASUPPERLIMIT),
			GasPrice: tv.RetrieveGasPrice(),
			Amount:   tv.RetrieveAmount(),
			Payload:  tv.RetrievePayload(),
			Op:       tv.GetOp(),
		}
	default:
		// Current version
		return Message{
			From:     common.BytesToAddress(tx.From),
			To:       common.BytesToAddress(tx.To),
			Gas:      tv.RetrieveGas(),
			GasPrice: tv.RetrieveGasPrice(),
			Amount:   tv.RetrieveAmount(),
			Payload:  tv.RetrievePayload(),
			Op:       tv.GetOp(),
		}
	}
}
