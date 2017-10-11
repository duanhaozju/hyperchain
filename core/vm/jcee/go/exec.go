package jvm

import (
	"bytes"
	"github.com/op/go-logging"
	"hyperchain/common"
	er "hyperchain/core/errors"
	"hyperchain/core/types"
	"hyperchain/core/vm"
	"math/big"
)

func ExecTransaction(db vm.Database, tx *types.Transaction, idx int, blockNumber uint64, logger *logging.Logger, namespace string, jvmCli ContractExecutor) (*types.Receipt, []byte, common.Address, error) {
	var (
		from   = common.BytesToAddress(tx.From)
		to     = common.BytesToAddress(tx.To)
		tv     = tx.GetTransactionValue()
		data   = tv.RetrievePayload()
		amount = tv.RetrieveAmount()
		op     = tv.GetOp()
	)
	env := initEnvironment(db, blockNumber, logger, namespace, tx.GetHash(), jvmCli)
	if env == nil {
		return nil, nil, common.Address{}, er.ExecContractErr(1, "init environment failed")
	}
	env.Db().StartRecord(tx.GetHash(), common.Hash{}, idx)
	if valid := checkPermission(env, from, to, op); !valid {
		return nil, nil, common.Address{}, er.InvalidInvokePermissionErr("not enough permission to invocation")
	}
	if tx.To == nil {
		// deploy
		ret, addr, err := Exec(env, &from, nil, amount, data, op)
		receipt := makeReceipt(env, addr, tx.GetHash(), ret, err)
		return receipt, ret, addr, err
	} else {
		// invoke
		ret, addr, err := Exec(env, &from, &to, amount, data, op)
		receipt := makeReceipt(env, common.Address{}, tx.GetHash(), ret, err)
		return receipt, ret, addr, err
	}
}

func Exec(vmenv vm.Environment, from, to *common.Address, value *big.Int, data []byte, op types.TransactionValue_Opcode) (ret []byte, addr common.Address, err error) {
	var (
		sender vm.Account
	)

	if !(vmenv.Db().Exist(*from)) {
		sender = vmenv.Db().CreateAccount(*from)
	} else {
		sender = vmenv.Db().GetAccount(*from)
	}

	contractCreation := (nil == to)

	if contractCreation {
		ret, addr, err = vmenv.Create(sender, data, nil, nil, value)
		if err != nil {
			ret = nil
			vmenv.Logger().Errorf("VM create err: %v", err)
		}
	} else {
		ret, err = vmenv.Call(sender, *to, data, nil, nil, value, int32(op))
		if err != nil {
			vmenv.Logger().Errorf("VM call err: %v", err)
		}
	}
	return ret, addr, err
}

func initEnvironment(state vm.Database, seqNo uint64, logger *logging.Logger, namespace string, txHash common.Hash, jvmCli ContractExecutor) vm.Environment {
	vmenv := NewEnv(state, int64(seqNo), 0, logger, namespace, txHash, jvmCli)
	if vmenv == nil {
		return nil
	}
	return vmenv
}

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

func makeReceipt(env vm.Environment, addr common.Address, txHash common.Hash, ret []byte, err error) *types.Receipt {
	// jvm receipt vmType = 1
	receipt := types.NewReceipt(big.NewInt(0), 1)
	receipt.ContractAddress = addr.Bytes()
	receipt.TxHash = txHash.Bytes()
	receipt.GasUsed = 100000
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
