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
		update   = tv.GetUpdate()
		freeze   = tv.GetFreeze()
	)

	if valid := checkPermission(env, from, to, update, freeze); !valid {
		return nil, nil, common.Address{}, InvalidInvokePermissionErr("not enough permission to invocation")
	}

	if tx.To == nil {
		ret, addr, err = Exec(env, &from, nil, data, gas, gasPrice, amount, update)
	} else {
		ret, _, err = Exec(env, &from, &to, data, gas, gasPrice, amount, update)
	}
	receipt = types.NewReceipt(nil, gas)
	receipt.ContractAddress = addr.Bytes()
	receipt.TxHash = tx.GetHash().Bytes()
	receipt.GasUsed = 100000
	receipt.Ret = ret
	receipt.SetLogs(env.Db().GetLogs(common.BytesToHash(receipt.TxHash)))

	if err != nil {
		if !IsValueTransferErr(err) && !IsExecContractErr(err) {
			receipt.Status = types.Receipt_FAILED
			receipt.Message = []byte(err.Error())
		}
	} else {
		receipt.Status = types.Receipt_SUCCESS
		receipt.Message = nil
	}
	return receipt, ret, addr, err
}

func Exec(vmenv vm.Environment, from, to *common.Address, data []byte, gas,
gasPrice, value *big.Int, update bool) (ret []byte, addr common.Address, err error) {
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
		ret, err = vmenv.Call(sender, *to, data, gas, gasPrice, value, update)
		if err != nil {
			log.Error("VM call err:", err)
		}
	}
	return ret, addr, err
}

func checkPermission(env vm.Environment, from, to common.Address, update bool, freeze bool) bool {
	if update == true || freeze == true {
		log.Debugf("caller address %s", from.Hex())
		log.Debugf("callee address %s", from.Hex())
		if bytes.Compare(to.Bytes(), nil) == 0 {
			return false
		}
		if bytes.Compare(from.Bytes(), env.Db().GetCreator(to).Bytes()) != 0 {
			log.Errorf("only contract owner %s has `freeze` or `update` permission. %s doesn't has enough permission",
				env.Db().GetCreator(to).Hex(), from.Hex())
			return false
		}
	}
	return true
}
