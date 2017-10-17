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
	edb "hyperchain/core/ledger/chain"
	"hyperchain/hyperdb/db"
	"hyperchain/manager/event"
	"hyperchain/manager/protos"
)

// reset blockchain to a stable checkpoint status when `viewchange` occur
func (executor *Executor) Rollback(ev event.VCResetEvent) {
	executor.waitUtilRollbackAvailable()
	defer executor.rollbackDone()

	executor.logger.Noticef("[Namespace = %s] receive vc reset event, required revert to %d", executor.namespace, ev.SeqNo-1)
	batch := executor.db.NewBatch()
	// revert state
	if err := executor.revertState(batch, ev.SeqNo-1); err != nil {
		return
	}
	// Delete related transaction, receipt, txmeta, and block itself in a specific range
	if err := executor.cutdownChain(batch, ev.SeqNo-1); err != nil {
		executor.logger.Errorf("[Namespace = %s] remove block && transaction in range %d to %d failed.", ev.SeqNo, edb.GetHeightOfChain(executor.namespace))
		return
	}
	// remove uncommitted data
	if err := executor.clearUncommittedData(batch); err != nil {
		executor.logger.Errorf("[Namespace = %s] remove uncommitted data failed", executor.namespace)
		return
	}
	// Reset chain
	edb.UpdateChainByBlcokNum(executor.namespace, batch, ev.SeqNo-1, false, false)
	batch.Write()
	executor.initDemand(ev.SeqNo)
	executor.informConsensus(NOTIFY_VC_DONE, protos.VcResetDone{SeqNo: ev.SeqNo})
	NotifyViewChange(executor.helper, ev.SeqNo)
}

// CutdownBlock remove a block and reset blockchain status to the last status.
func (executor *Executor) CutdownBlock(number uint64) error {
	executor.waitUtilRollbackAvailable()
	defer executor.rollbackDone()

	executor.logger.Noticef("[Namespace = %s] cutdown block, required revert to %d", executor.namespace, number)
	// 2. revert state
	batch := executor.db.NewBatch()
	if err := executor.revertState(batch, number-1); err != nil {
		return err
	}
	// 3. remove block releted data
	if err := executor.cutdownChainByRange(batch, number, number); err != nil {
		executor.logger.Errorf("remove block && transaction %d", number)
		return err
	}
	// 4. remove uncommitted data
	if err := executor.clearUncommittedData(batch); err != nil {
		executor.logger.Errorf("remove uncommitted of %d failed", number)
		return err
	}
	// 5. reset chain data
	edb.UpdateChainByBlcokNum(executor.namespace, batch, number-1, false, false)
	// flush all modified to disk
	batch.Write()
	executor.logger.Noticef("[Namespace = %s] cut down block #%d success. remove all related transactions, receipts, state changes and block together.", executor.namespace, number)
	executor.initDemand(edb.GetHeightOfChain(executor.namespace))
	return nil
}

func (executor *Executor) cutdownChain(batch db.Batch, targetHeight uint64) error {
	return executor.cutdownChainByRange(batch, targetHeight+1, edb.GetHeightOfChain(executor.namespace))
}

// cutdownChainByRange - remove block, tx, receipt in range.
func (executor *Executor) cutdownChainByRange(batch db.Batch, from, to uint64) error {
	for i := from; i <= to; i += 1 {
		block, err := edb.GetBlockByNumber(executor.namespace, i)
		if err != nil {
			executor.logger.Errorf("miss block %d ,error msg %s", i, err.Error())
			continue
		}

		for _, tx := range block.Transactions {
			if err := edb.DeleteTransactionMeta(batch, tx.GetHash().Bytes(), false, false); err != nil {
				executor.logger.Errorf("[Namespace = %s] delete useless tx meta in block %d failed, error msg %s", executor.namespace, i, err.Error())
			}
			if err := edb.DeleteReceipt(batch, tx.GetHash().Bytes(), false, false); err != nil {
				executor.logger.Errorf("[Namespace = %s] delete useless receipt in block %d failed, error msg %s", executor.namespace, i, err.Error())
			}
		}
		edb.AddTxDeltaOfMemChain(executor.namespace, uint64(len(block.Transactions)))
		// delete block
		if err := edb.DeleteBlockByNum(executor.namespace, batch, i, false, false); err != nil {
			executor.logger.Errorf("[Namespace = %s] delete useless block %d failed, error msg %s", executor.namespace, i, err.Error())
		}
	}
	return nil
}

// revertState revert state from currentNumber related status to a target
func (executor *Executor) revertState(batch db.Batch, targetHeight uint64) error {
	currentHeight := edb.GetHeightOfChain(executor.namespace)
	targetBlk, err := edb.GetBlockByNumber(executor.namespace, targetHeight)
	if err != nil {
		return err
	}
	if err := executor.statedb.RevertToJournal(targetHeight, currentHeight, targetBlk.MerkleRoot, batch); err != nil {
		return err
	}
	return nil
}

// removeUncommittedData remove uncommitted validation result avoid of memory leak.
func (executor *Executor) clearUncommittedData(batch db.Batch) error {
	executor.statedb.Purge()
	return nil
}
