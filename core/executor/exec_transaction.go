package executor

import (
	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/core/vm"
	"hyperchain/core/vm/evm"
)

func (executor *Executor) ExecTransaction(db vm.Database, tx *types.Transaction, idx int, blockNumber uint64) (*types.Receipt, []byte, common.Address, error) {
	return evm.ExecTransaction(db, tx, idx, blockNumber, executor.logger, executor.namespace)
}


