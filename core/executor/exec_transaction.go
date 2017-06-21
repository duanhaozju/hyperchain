package executor

import (
	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/core/vm"
	"hyperchain/core/vm/evm"
	"hyperchain/core/vm/jcee/go"
	er "hyperchain/core/errors"
)

func (executor *Executor) ExecTransaction(db vm.Database, tx *types.Transaction, idx int, blockNumber uint64) (*types.Receipt, []byte, common.Address, error) {
	tv := tx.GetTransactionValue()
	switch tv.GetVmType() {
	case types.TransactionValue_EVM:
		executor.logger.Debug("execute in evm")
		return evm.ExecTransaction(db, tx, idx, blockNumber, executor.logger, executor.namespace)
	case types.TransactionValue_JVM:
		executor.logger.Debug("execute in jvm")
		return jvm.ExecTransaction(db, tx, idx, blockNumber, executor.logger, executor.namespace, executor.jvmCli)
	default:
		executor.logger.Warningf("try to execute a transaction with undefined vm type %s", tv.GetVmType().String())
		return nil, nil, common.Address{}, er.ExecContractErr(1, "undefined vm type")
	}
}


