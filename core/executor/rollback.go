package executor

import (
	"bytes"
	"github.com/deckarep/golang-set"
	"github.com/pkg/errors"
	"hyperchain/common"
	"hyperchain/core/hyperstate"
	"hyperchain/event"
	"hyperchain/hyperdb"
	"hyperchain/protos"
	"hyperchain/tree/bucket"
	"math/big"
	"hyperchain/hyperdb/db"
	edb "hyperchain/core/db_utils"
)

// reset blockchain to a stable checkpoint status when `viewchange` occur
func (executor *Executor) Rollback(ev event.VCResetEvent) {
	executor.waitUtilRollbackAvailable()
	defer executor.rollbackDone()

	log.Noticef("[Namespace = %s] receive vc reset event, required revert to %d", executor.namespace, ev.SeqNo-1)
	db, err := hyperdb.GetDBDatabaseByNamespace(executor.namespace)
	if err != nil {
		log.Errorf("[Namespace = %s] get database handler failed. error : ", executor.namespace, err.Error())
		return
	}
	batch := db.NewBatch()
	// revert state
	if err := executor.revertState(batch, ev.SeqNo - 1); err != nil {
		return
	}
	// Delete related transaction, receipt, txmeta, and block itself in a specific range
	if err := executor.cutdownChain(batch, ev.SeqNo - 1); err != nil {
		log.Errorf("[Namespace = %s] remove block && transaction in range %d to %d failed.", ev.SeqNo, edb.GetHeightOfChain(executor.namespace))
		return
	}
	// remove uncommitted data
	if err := executor.clearUncommittedData(batch); err != nil {
		log.Errorf("[Namespace = %s] remove uncommitted data failed", executor.namespace)
		return
	}
	// Reset chain
	edb.UpdateChainByBlcokNum(executor.namespace, batch, ev.SeqNo - 1, false, false)
	batch.Write()
	executor.initDemand(ev.SeqNo)
	executor.informConsensus(CONSENSUS_LOCAL, protos.VcResetDone{SeqNo: ev.SeqNo})
}

// CutdownBlock remove a block and reset blockchain status to the last status.
func (executor *Executor) CutdownBlock(number uint64) error {
	executor.waitUtilRollbackAvailable()
	defer executor.rollbackDone()

	log.Noticef("[Namespace = %s] cutdown block, required revert to %d", executor.namespace, number)
	// 2. revert state
	db, err := hyperdb.GetDBDatabaseByNamespace(executor.namespace)
	if err != nil {
		log.Error("Get Database Instance Failed! error msg,", err.Error())
		return err
	}
	batch := db.NewBatch()
	if err := executor.revertState(batch, number - 1); err != nil {
		return err
	}
	// 3. remove block releted data
	if err := executor.cutdownChainByRange(batch, number, number); err != nil {
		log.Errorf("remove block && transaction %d", number)
		return err
	}
	// 4. remove uncommitted data
	if err := executor.clearUncommittedData(batch); err != nil {
		log.Errorf("remove uncommitted of %d failed", number)
		return err
	}
	// 5. reset chain data
	edb.UpdateChainByBlcokNum(executor.namespace, batch, number - 1, false, false)
	// flush all modified to disk
	batch.Write()
	log.Noticef("[Namespace = %s] cut down block #%d success. remove all related transactions, receipts, state changes and block together.", executor.namespace, number)
	executor.initDemand(edb.GetHeightOfChain(executor.namespace))
	return nil
}

func (executor *Executor) cutdownChain(batch db.Batch, targetHeight uint64) error {
	return executor.cutdownChainByRange(batch, targetHeight + 1, edb.GetHeightOfChain(executor.namespace))
}

// cutdownChainByRange - remove block, tx, receipt in range.
func (executor *Executor) cutdownChainByRange(batch db.Batch, from, to uint64) error  {
	for i := from; i <= to; i += 1 {
		block, err := edb.GetBlockByNumber(executor.namespace, i)
		if err != nil {
			log.Errorf("miss block %d ,error msg %s", i, err.Error())
			continue
		}

		for _, tx := range block.Transactions {
			if err := edb.DeleteTransaction(batch, tx.GetTransactionHash().Bytes(), false, false); err != nil {
				log.Errorf("[Namespace = %s] delete useless tx in block %d failed, error msg %s", executor.namespace, i, err.Error())
			}
			if err := edb.DeleteReceipt(batch, tx.GetTransactionHash().Bytes(), false, false); err != nil {
				log.Errorf("[Namespace = %s] delete useless receipt in block %d failed, error msg %s", executor.namespace, i, err.Error())
			}
			if err := edb.DeleteTransactionMeta(batch, tx.GetTransactionHash().Bytes(), false, false); err != nil {
				log.Errorf("[Namespace = %s] delete useless txmeta in block %d failed, error msg %s", executor.namespace, i, err.Error())
			}
		}
		edb.SetTxDeltaOfMemChain(executor.namespace, uint64(len(block.Transactions)))
		// delete block
		if err := edb.DeleteBlockByNum(executor.namespace, batch, i, false, false); err != nil {
			log.Errorf("[Namespace = %s] delete useless block %d failed, error msg %s", executor.namespace, i, err.Error())
		}
	}
	return nil
}

// revertState revert state from currentNumber related status to a target
func (executor *Executor) revertState(batch db.Batch, targetHeight uint64) error {
	currentHeight := edb.GetHeightOfChain(executor.namespace)
	targetBlk, err := edb.GetBlockByNumber(executor.namespace, targetHeight)
	if err != nil {
		log.Errorf("[Namespace = %s] get block #%d failed", executor.namespace, targetHeight)
		return err
	}
	db, err := hyperdb.GetDBDatabaseByNamespace(executor.namespace)
	if err != nil {
		log.Errorf("[Namespace = %s] get database handler failed.", executor.namespace)
		return err
	}
	dirtyStateObjectSet := mapset.NewSet()
	stateObjectStorageHashs := make(map[common.Address][]byte)
	// revert state change with changeset [targetNumber+1, currentNumber]
	// undo changes in reverse
	journalCache := hyperstate.NewJournalCache(db)
	for i := currentHeight; i >= targetHeight + 1; i -= 1 {
		log.Debugf("[Namespace = %s] undo changes for #%d", executor.namespace, i)
		j, err := db.Get(hyperstate.CompositeJournalKey(uint64(i)))
		if err != nil {
			log.Warningf("[Namespace = %s] get journal in database for #%d failed. make sure #%d doesn't have state change",
				executor.namespace, i, i)
			continue
		}
		journal, err := hyperstate.UnmarshalJournal(j)
		if err != nil {
			log.Errorf("[Namespace = %s] unmarshal journal for #%d failed", executor.namespace, i)
			continue
		}
		// undo journal in reverse
		for j := len(journal.JournalList) - 1; j >= 0; j -= 1 {
			log.Debugf("[Namespace = %s] journal %s", executor.namespace, journal.JournalList[j].String())
			t := executor.statedb.(*hyperstate.StateDB)
			journal.JournalList[j].Undo(t, journalCache, batch, true)
			if journal.JournalList[j].GetType() == hyperstate.StorageHashChangeType {
				tmp := journal.JournalList[j].(*hyperstate.StorageHashChange)
				// use struct instead of pointer since different pointers may represent same stateObject
				dirtyStateObjectSet.Add(*tmp.Account)
				stateObjectStorageHashs[*tmp.Account] = tmp.Prev
			}
		}
		// remove persisted journals
		batch.Delete(hyperstate.CompositeJournalKey(uint64(i)))
	}
	if err := journalCache.Flush(batch); err != nil {
		log.Errorf("flush modified content failed. %s", err.Error())
		return err
	}
	// revert related stateObject storage bucket tree
	for addr := range dirtyStateObjectSet.Iter() {
		address := addr.(common.Address)
		prefix, _ := hyperstate.CompositeStorageBucketPrefix(address.Bytes())
		bucketTree := bucket.NewBucketTree(string(prefix))
		bucketTree.Initialize(hyperstate.SetupBucketConfig(executor.GetBucketSize(STATEOBJECT), executor.GetBucketLevelGroup(STATEOBJECT), executor.GetBucketCacheSize(STATEOBJECT)))
		bucketTree.ClearAllCache()
		// don't flush into disk util all operations finish
		bucketTree.PrepareWorkingSet(journalCache.GetWorkingSet(hyperstate.WORKINGSET_TYPE_STATEOBJECT, address), big.NewInt(0))
		//bucketTree.RevertToTargetBlock(batch, big.NewInt(currentNumber), big.NewInt(targetNumber), false, false)
		hash, _ := bucketTree.ComputeCryptoHash()
		bucketTree.AddChangesForPersistence(batch, big.NewInt(int64(targetHeight)))
		log.Debugf("[Namespace = %s] re-compute %s storage hash %s", executor.namespace, address.Hex(), common.Bytes2Hex(hash))
		stateObjectHash := stateObjectStorageHashs[address]
		if common.BytesToHash(hash).Hex() != common.BytesToHash(stateObjectHash).Hex() {
			log.Errorf("[Namespace = %s] after revert to #%d, state object %s revert failed, required storage hash %s, got %s",
				executor.namespace, targetHeight, address.Hex(), common.Bytes2Hex(stateObjectHash), common.Bytes2Hex(hash))
			return errors.New("revert state failed.")
		}
	}
	// revert state bucket tree
	tree := executor.statedb.GetTree()
	bucketTree := tree.(*bucket.BucketTree)
	bucketTree.Initialize(hyperstate.SetupBucketConfig(executor.GetBucketSize(STATEDB), executor.GetBucketLevelGroup(STATEDB), executor.GetBucketCacheSize(STATEDB)))
	bucketTree.ClearAllCache()
	// don't flush into disk util all operations finish
	bucketTree.PrepareWorkingSet(journalCache.GetWorkingSet(hyperstate.WORKINGSET_TYPE_STATE, common.Address{}), big.NewInt(0))
	//bucketTree.RevertToTargetBlock(batch, big.NewInt(currentNumber), big.NewInt(targetNumber), false, false)
	currentRootHash, err := bucketTree.ComputeCryptoHash()
	if err != nil {
		log.Errorf("[Namespace = %s] re-compute state bucket tree hash failed, error :%s", executor.namespace, err.Error())
		return err
	}
	bucketTree.AddChangesForPersistence(batch, big.NewInt(int64(targetHeight)))
	log.Debugf("re-compute state hash %s", common.Bytes2Hex(currentRootHash))
	if bytes.Compare(currentRootHash, targetBlk.MerkleRoot) != 0 {
		log.Errorf("[Namespace = %s] revert to a different state, required %s, but current state %s",
			executor.namespace, common.Bytes2Hex(targetBlk.MerkleRoot), common.Bytes2Hex(currentRootHash))
		return errors.New("revert state failed")
	}
	// revert state instance oldest and root
	executor.statedb.ResetToTarget(uint64(targetHeight +1), common.BytesToHash(targetBlk.MerkleRoot))
	executor.recordStateHash(common.BytesToHash(targetBlk.MerkleRoot))
	log.Debugf("[Namespace = %s] revert state from #%d to #%d success", executor.namespace, currentHeight, targetHeight)
	return nil
}

// removeUncommittedData remove uncommitted validation result avoid of memory leak.
func (executor *Executor) clearUncommittedData(batch db.Batch) error {
	executor.statedb.Purge()
	return nil
}

