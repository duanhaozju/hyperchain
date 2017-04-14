package jvm

import (
	"hyperchain/core/types"
	"hyperchain/common"
	"github.com/op/go-logging"
	"hyperchain/core/vm"
	"strconv"
	"math/big"
	"bytes"
	er "hyperchain/core/errors"
)

func ExecTransaction(db vm.Database, tx *types.Transaction, idx int, blockNumber uint64, logger *logging.Logger) (*types.Receipt, []byte, common.Address, error) {
	var (
		from     = common.BytesToAddress(tx.From)
		to       = common.BytesToAddress(tx.To)
		tv       = tx.GetTransactionValue()
		// data     = tv.RetrievePayload()
		op       = tv.GetOp()

	)
	env := initEnvironment(db, blockNumber, logger)
	env.Db().StartRecord(tx.GetHash(), common.Hash{}, idx)
	if valid := checkPermission(env, from, to, op); !valid {
		return nil, nil, common.Address{}, er.InvalidInvokePermissionErr("not enough permission to invocation")
	}
	if tx.To == nil {
		// deploy
	} else {
		// invoke
	}
	return nil, nil, common.Address{}, nil
}


func Exec(vmenv vm.Environment, from, to *common.Address, data []byte, op types.TransactionValue_Opcode) (ret []byte, addr common.Address, err error) {
	return nil, common.Address{}, nil
}

func initEnvironment(state vm.Database, seqNo uint64, logger *logging.Logger) vm.Environment {
	env := make(map[string]string)
	env["currentNumber"] = strconv.FormatUint(seqNo, 10)
	env["currentGasLimit"] = "200000000"
	vmenv := NewEnv(state, env, logger)
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

func makeReceipt(env vm.Environment, addr common.Address, txHash common.Hash, gas *big.Int, ret []byte, err error) *types.Receipt {
	// evm receipt vmType = 0
	receipt := types.NewReceipt(nil, gas, 0)
	receipt.ContractAddress = addr.Bytes()
	receipt.TxHash = txHash.Bytes()
	receipt.GasUsed = 100000
	receipt.Ret = ret

	// TODO vm log support
	//logs := env.Db().GetLogs(common.BytesToHash(receipt.TxHash))
	//var tmp Logs
	//for _, l := range logs {
	//	tmp = append(tmp, l.(*Log))
	//}
	//buf, _ := (&tmp).EncodeLogs()
	//receipt.Logs = buf

	if err != nil {
		if !er.IsValueTransferErr(err) && !er.IsExecContractErr(err) &&!er.IsInvalidInvokePermissionErr(err) {
			receipt.Status = types.Receipt_FAILED
			receipt.Message = []byte(err.Error())
		}
	} else {
		receipt.Status = types.Receipt_SUCCESS
		receipt.Message = nil
	}
	return receipt
}
