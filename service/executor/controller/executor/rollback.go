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
	"github.com/hyperchain/hyperchain/hyperdb/db"
	"github.com/hyperchain/hyperchain/manager/event"
	"github.com/hyperchain/hyperchain/manager/protos"
)

// Rollback is called by manager to reset blockchain to a stable checkpoint status when `viewchange` occurs.
func (executor *Executor) Rollback(ev event.VCResetEvent) {
	// Wait util current validating and committing done
	executor.waitUtilRollbackAvailable()
	defer executor.rollbackDone()

	executor.logger.Noticef("[Namespace = %s] receive vc reset event, required revert to %d", executor.namespace, ev.SeqNo-1)
	batch := executor.db.NewBatch()
	// Revert state
	if err := executor.revertState(batch, ev.SeqNo-1); err != nil {
		return
	}
	// Delete related transaction, receipt, txmeta, and block itself in a specific range
	if err := executor.cutdownChain(batch, ev.SeqNo-1); err != nil {
		executor.logger.Errorf("[Namespace = %s] remove block && transaction in range %d to %d failed.", ev.SeqNo, chain.GetHeightOfChain(executor.namespace))
		return
	}
	// Remove uncommitted data
	if err := executor.clearUncommittedData(batch); err != nil {
		executor.logger.Errorf("[Namespace = %s] remove uncommitted data failed", executor.namespace)
		return
	}
	// Reset chain
	chain.UpdateChainByBlcokNum(executor.namespace, batch, ev.SeqNo-1, false, false)
	batch.Write()
	executor.context.initDemand(ev.SeqNo)
	executor.informConsensus(NOTIFY_VC_DONE, protos.VcResetDone{SeqNo: ev.SeqNo, View: ev.View})
	NotifyViewChange(executor.helper, ev.SeqNo)
}

// CutdownBlock remove a block and reset blockchain status to the last status.
func (executor *Executor) CutdownBlock(number uint64) error {
	// Wait util current validating and committing done
	executor.waitUtilRollbackAvailable()
	defer executor.rollbackDone()

	executor.logger.Noticef("[Namespace = %s] cutdown block, required revert to %d", executor.namespace, number)
	// Revert state
	batch := executor.db.NewBatch()
	if err := executor.revertState(batch, number-1); err != nil {
		return err
	}
	// Remove block related data
	if err := executor.cutdownChainByRange(batch, number, number); err != nil {
		executor.logger.Errorf("remove block && transaction %d", number)
		return err
	}
	// Remove uncommitted data
	if err := executor.clearUncommittedData(batch); err != nil {
		executor.logger.Errorf("remove uncommitted of %d failed", number)
		return err
	}
	// Reset chain data
	chain.UpdateChainByBlcokNum(executor.namespace, batch, number-1, false, false)
	// Flush all modified to disk
	batch.Write()
	executor.logger.Noticef("[Namespace = %s] cut down block #%d success. remove all related transactions, receipts, state changes and block together.", executor.namespace, number)
	executor.context.initDemand(chain.GetHeightOfChain(executor.namespace))
	return nil
}

// cutdownChain cuts down the chain to the target height.
func (executor *Executor) cutdownChain(batch db.Batch, targetHeight uint64) error {
	return executor.cutdownChainByRange(batch, targetHeight+1, chain.GetHeightOfChain(executor.namespace))
}

// cutdownChainByRange removes blocks, txs, receipts in range( [from, to] ).
func (executor *Executor) cutdownChainByRange(batch db.Batch, from, to uint64) error {
	for i := from; i <= to; i += 1 {
		block, err := chain.GetBlockByNumber(executor.namespace, i)
		if err != nil {
			executor.logger.Errorf("miss block %d ,error msg %s", i, err.Error())
			continue
		}

		// Delete all the tx metas and receipts
		for _, tx := range block.Transactions {
			if err := chain.DeleteTransactionMeta(batch, tx.GetHash().Bytes(), false, false); err != nil {
				executor.logger.Errorf("[Namespace = %s] delete useless tx meta in block %d failed, error msg %s", executor.namespace, i, err.Error())
			}
			if err := chain.DeleteReceipt(batch, tx.GetHash().Bytes(), false, false); err != nil {
				executor.logger.Errorf("[Namespace = %s] delete useless receipt in block %d failed, error msg %s", executor.namespace, i, err.Error())
			}
		}
		chain.AddTxDeltaOfMemChain(executor.namespace, uint64(len(block.Transactions)))
		// Delete block
		if err := chain.DeleteBlockByNum(executor.namespace, batch, i, false, false); err != nil {
			executor.logger.Errorf("[Namespace = %s] delete useless block %d failed, error msg %s", executor.namespace, i, err.Error())
		}
	}
	return nil
}

// revertState reverts state from currentNumber related status to a target
func (executor *Executor) revertState(batch db.Batch, targetHeight uint64) error {
	currentHeight := chain.GetHeightOfChain(executor.namespace)
	targetBlk, err := chain.GetBlockByNumber(executor.namespace, targetHeight)
	if err != nil {
		return err
	}
	if err := executor.statedb.RevertToJournal(targetHeight, currentHeight, targetBlk.MerkleRoot, batch); err != nil {
		return err
	}
	return nil
}

// removeUncommittedData removes uncommitted validation result avoid of memory leak.
func (executor *Executor) clearUncommittedData(batch db.Batch) error {
	executor.statedb.Purge()
	return nil
}