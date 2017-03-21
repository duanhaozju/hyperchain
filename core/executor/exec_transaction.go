package executor

import (
	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/core/vm"
	"math/big"
	"bytes"
)

type Code []byte

func ExecTransaction(tx *types.Transaction, env vm.Environment) (receipt *types.Receipt, ret []byte, addr common.Address, err error) {
	var (
		from     = common.BytesToAddress(tx.From)
		to       = common.BytesToAddress(tx.To)
		tv       = tx.GetTransactionValue()
		data     = tv.RetrievePayload()
		gas      = big.NewInt(100000000)
		gasPrice = tv.RetrieveGasPrice()
		amount   = tv.RetrieveAmount()
		op       = tv.GetOp()
	)

	if valid := checkPermission(env, from, to, op); !valid {
		return nil, nil, common.Address{}, InvalidInvokePermissionErr("not enough permission to invocation")
	}

	if tx.To == nil {
		ret, addr, err = Exec(env, &from, nil, data, gas, gasPrice, amount, op)
	} else {
		ret, _, err = Exec(env, &from, &to, data, gas, gasPrice, amount, op)
	}

	receipt = makeReceipt(env, addr, tx.GetHash(), gas, ret, err)
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
			log.Error("VM create err:", err)
		}
	} else {
		ret, err = vmenv.Call(sender, *to, data, gas, gasPrice, value, int32(op))
		if err != nil {
			log.Error("VM call err:", err)
		}
	}
	return ret, addr, err
}

func checkPermission(env vm.Environment, from, to common.Address, op types.TransactionValue_Opcode) bool {
	if op == types.TransactionValue_UPDATE || op == types.TransactionValue_FREEZE || op == types.TransactionValue_UNFREEZE {
		log.Debugf("caller address %s", from.Hex())
		log.Debugf("callee address %s", from.Hex())
		if bytes.Compare(to.Bytes(), nil) == 0 {
			return false
		}
		if bytes.Compare(from.Bytes(), env.Db().GetCreator(to).Bytes()) != 0 {
			log.Errorf("only contract owner %s can do `update`, `freeze` or `unfreeze` operations. %s doesn't has enough permission",
				env.Db().GetCreator(to).Hex(), from.Hex())
			return false
		}
	}
	return true
}


func makeReceipt(env vm.Environment, addr common.Address, txHash common.Hash, gas *big.Int, ret []byte, err error) *types.Receipt {
	receipt := types.NewReceipt(nil, gas)
	receipt.ContractAddress = addr.Bytes()
	receipt.TxHash = txHash.Bytes()
	receipt.GasUsed = 100000
	receipt.Ret = ret
	receipt.SetLogs(env.Db().GetLogs(common.BytesToHash(receipt.TxHash)))

	if err != nil {
		if !IsValueTransferErr(err) && !IsExecContractErr(err) &&!IsInvalidInvokePermissionErr(err) {
			receipt.Status = types.Receipt_FAILED
			receipt.Message = []byte(err.Error())
		}
	} else {
		receipt.Status = types.Receipt_SUCCESS
		receipt.Message = nil
	}
	return receipt
}

