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
	"github.com/hyperchain/hyperchain/core/ledger/chain"
	"github.com/hyperchain/hyperchain/core/types"
	"github.com/hyperchain/hyperchain/core/vm"

	"github.com/golang/protobuf/proto"
	"time"
)

// RunInSandBox is used to run transaction in a simulator,
// in which way the execution result will not been added to database.
func (executor *Executor) RunInSandBox(tx *types.Transaction, snapshotId string) error {
	var (
		statedb  vm.Database
		err      error
		callback func()
	)
	if snapshotId == "" {
		statedb, err = executor.newStateDb()
	} else {
		statedb, err, callback = executor.initHistoryStateDb(snapshotId)
	}
	if err != nil {
		return err
	}
	// Release history db handler
	defer func() {
		if callback != nil {
			callback()
		}
	}()
	// Init execution environment
	fakeBlockNumber := chain.GetHeightOfChain(executor.namespace) + 1
	// Execute the tx
	receipt, _, _, err := executor.ExecTransaction(statedb, tx, 0, fakeBlockNumber, time.Now().Unix())
	if err != nil {
		errType := executor.classifyInvalid(err)
		t := &types.InvalidTransactionRecord{
			Tx:      tx,
			ErrType: errType,
			ErrMsg:  []byte(err.Error()),
		}
		payload, err := proto.Marshal(t)
		if err != nil {
			executor.logger.Error("Marshal tx error")
			return err
		}
		// Persist execution result to local
		executor.StoreInvalidTransaction(payload)
		return nil
	} else {
		// Persist execution result to local
		_, err := chain.PersistReceipt(executor.db.NewBatch(), receipt, true, true)
		if err != nil {
			executor.logger.Error("Put receipt data into database failed! error msg, ", err.Error())
			return err
		}
		return nil
	}
}
