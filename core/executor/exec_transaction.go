package executor

import (
	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/core/vm"
	er "hyperchain/core/errors"
	"hyperchain/core/vm/evm"
)

func (executor *Executor) ExecTransaction(db vm.Database, tx *types.Transaction, idx int, blockNumber uint64) (*types.Receipt, []byte, common.Address, error) {
	tv := tx.GetTransactionValue()
	switch tv.GetVmType() {
	case types.TransactionValue_EVM:
		return evm.ExecTransaction(db, tx, idx, blockNumber, executor.logger)
	case types.TransactionValue_JVM:
		return nil, nil, common.Address{}, nil
	default:
		executor.logger.Warningf("try to execute a transaction with undefined vm type %s", tv.GetVmType().String())
		return nil, nil, common.Address{}, er.ExecContractErr(1, "undefined vm type")
	}
}
