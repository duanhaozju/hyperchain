package executor

import (
	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/core/vm/evm"
	"math/big"
	"bytes"
)

type Code []byte

func (executor *Executor) ExecTransaction(db evm.Database, tx *types.Transaction, idx int, blockNumber uint64) (receipt *types.Receipt, ret []byte, addr common.Address, err error) {
	var (
		from     = common.BytesToAddress(tx.From)
		to       = common.BytesToAddress(tx.To)
		tv       = tx.GetTransactionValue()
		data     = tv.RetrievePayload()
		gas      = big.NewInt(100000000)
		gasPrice = tv.RetrieveGasPrice()
		amount   = tv.RetrieveAmount()
		op       = tv.GetOp()
		vmType   = tv.GetVmType()
	)
	env := executor.initEnvironment(db, blockNumber)
	env.Db().StartRecord(tx.GetHash(), common.Hash{}, idx)
	if valid := executor.checkPermission(env, from, to, op); !valid {
		return nil, nil, common.Address{}, InvalidInvokePermissionErr("not enough permission to invocation")
	}

	if tx.To == nil {
		ret, addr, err = executor.Exec(env, &from, nil, data, gas, gasPrice, amount, op)
	} else {
		ret, _, err = executor.Exec(env, &from, &to, data, gas, gasPrice, amount, op)
	}

	receipt = makeReceipt(env, vmType, addr, tx.GetHash(), gas, ret, err)
	return receipt, ret, addr, err
}

func (executor *Executor) Exec(vmenv evm.Environment, from, to *common.Address, data []byte, gas,
gasPrice, value *big.Int, op types.TransactionValue_Opcode) (ret []byte, addr common.Address, err error) {
	var sender evm.Account

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
			executor.logger.Error("VM create err:", err)
		}
	} else {
		ret, err = vmenv.Call(sender, *to, data, gas, gasPrice, value, int32(op))
		if err != nil {
			executor.logger.Error("VM call err:", err)
		}
	}
	return ret, addr, err
}

func (executor *Executor) checkPermission(env evm.Environment, from, to common.Address, op types.TransactionValue_Opcode) bool {
	if op == types.TransactionValue_UPDATE || op == types.TransactionValue_FREEZE || op == types.TransactionValue_UNFREEZE {
		executor.logger.Debugf("caller address %s", from.Hex())
		executor.logger.Debugf("callee address %s", from.Hex())
		if bytes.Compare(to.Bytes(), nil) == 0 {
			return false
		}
		if bytes.Compare(from.Bytes(), env.Db().GetCreator(to).Bytes()) != 0 {
			executor.logger.Errorf("only contract owner %s can do `update`, `freeze` or `unfreeze` operations. %s doesn't has enough permission",
				env.Db().GetCreator(to).Hex(), from.Hex())
			return false
		}
	}
	return true
}


func makeReceipt(env evm.Environment, vmType types.TransactionValue_VmType, addr common.Address, txHash common.Hash, gas *big.Int, ret []byte, err error) *types.Receipt {
	receipt := types.NewReceipt(nil, gas, int32(vmType))
	receipt.ContractAddress = addr.Bytes()
	receipt.TxHash = txHash.Bytes()
	receipt.GasUsed = 100000
	receipt.Ret = ret
	SetLogs(receipt, int32(vmType), env.Db().GetLogs(common.BytesToHash(receipt.TxHash)))

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

