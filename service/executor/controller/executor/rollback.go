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
)

// Rollback is called by manager to reset blockchain to a stable checkpoint status when `viewchange` occurs.
// re.SeqNo = previous stable checkpoint + 1
func (e *Executor) Rollback(re *event.RollbackEvent) {
	// Wait util current validating and committing done
	e.waitUtilRollbackAvailable() //TODO(Xiaoyi Wang): modify this to adapt new architecture
	defer e.rollbackDone()

	e.logger.Noticef("receive vc reset event, required revert to %d", re.SeqNo-1)
	batch := e.db.NewBatch()
	// Revert state
	if err := e.revertState(batch, re.SeqNo-1); err != nil {
		return
	}
	// Delete related transaction, receipt, txmeta, and block itself in a specific range
	if err := e.cutdownChain(batch, re.SeqNo-1); err != nil {
		e.logger.Errorf("remove block && transaction in range %d to %d failed.", re.SeqNo, chain.GetHeightOfChain(e.namespace))
		return
	}
	// Remove uncommitted data
	if err := e.clearUncommittedData(batch); err != nil {
		e.logger.Errorf("remove uncommitted data failed")
		return
	}
	// Reset chain
	chain.UpdateChainByBlockNum(e.namespace, batch, re.SeqNo-1, false, false)
	if err := batch.Write(); err != nil {
		e.logger.Error(err)
		return
	}
	e.context.initDemand(re.SeqNo)
	e.context.setDemandOpLogIndex(re.Lid)
}

// cutdownChain cuts down the chain to the target height.
func (e *Executor) cutdownChain(batch db.Batch, targetHeight uint64) error {
	return e.cutdownChainByRange(batch, targetHeight+1, chain.GetHeightOfChain(e.namespace))
}

// cutdownChainByRange removes blocks, txs, receipts in range( [from, to] ).
func (e *Executor) cutdownChainByRange(batch db.Batch, from, to uint64) error {
	for i := from; i <= to; i += 1 {
		block, err := chain.GetBlockByNumber(e.namespace, i)
		if err != nil {
			e.logger.Errorf("miss block %d ,error msg %s", i, err.Error())
			continue
		}

		// Delete all the tx metas and receipts
		for _, tx := range block.Transactions {
			if err := chain.DeleteTransactionMeta(batch, tx.GetHash().Bytes(), false, false); err != nil {
				e.logger.Errorf("delete useless tx meta in block %d failed, error msg %s", i, err.Error())
			}
			if err := chain.DeleteReceipt(batch, tx.GetHash().Bytes(), false, false); err != nil {
				e.logger.Errorf("delete useless receipt in block %d failed, error msg %s", i, err.Error())
			}
		}
		chain.AddTxDeltaOfMemChain(e.namespace, uint64(len(block.Transactions)))
		// Delete block
		if err := chain.DeleteBlockByNum(e.namespace, batch, i, false, false); err != nil {
			e.logger.Errorf("delete useless block %d failed, error msg %s", i, err.Error())
		}
	}
	return nil
}

// revertState reverts state from currentNumber related status to a target
func (e *Executor) revertState(batch db.Batch, targetHeight uint64) error {
	currentHeight := chain.GetHeightOfChain(e.namespace)
	targetBlk, err := chain.GetBlockByNumber(e.namespace, targetHeight)
	if err != nil {
		return err
	}
	if err := e.statedb.RevertToJournal(targetHeight, currentHeight, targetBlk.MerkleRoot, batch); err != nil {
		return err
	}
	return nil
}

// removeUncommittedData removes uncommitted validation result avoid of memory leak.
func (e *Executor) clearUncommittedData(batch db.Batch) error {
	e.statedb.Purge()
	return nil
}
