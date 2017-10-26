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
package executor

import (
	"hyperchain/common"
	"hyperchain/core/errors"
	"hyperchain/core/types"
	"hyperchain/core/vm"
	"hyperchain/core/vm/evm"
	"hyperchain/core/vm/jcee/go"
)

// ExecTransaction executes the single transaction in evm or jvm.
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
		return nil, nil, common.Address{}, errors.ExecContractErr(1, "undefined vm type")
	}
}
