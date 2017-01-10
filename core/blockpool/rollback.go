package blockpool

import (
	"sync/atomic"
	"hyperchain/hyperdb"
	"hyperchain/event"
	"hyperchain/common"
	"hyperchain/core"
	"hyperchain/core/hyperstate"
	"github.com/deckarep/golang-set"
	"hyperchain/tree/bucket"
	"math/big"
	"bytes"
	"github.com/pkg/errors"
	"hyperchain/protos"
)


// reset blockchain to a stable checkpoint status when `viewchange` occur
func (pool *BlockPool) ResetStatus(ev event.VCResetEvent) {
	log.Debugf("receive vc reset event, required revert to %d", ev.SeqNo - 1)
	tmpDemandNumber := atomic.LoadUint64(&pool.demandNumber)
	// 1. Reset demandNumber , demandSeqNo and maxSeqNo
	atomic.StoreUint64(&pool.demandNumber, ev.SeqNo)
	atomic.StoreUint64(&pool.demandSeqNo, ev.SeqNo)
	atomic.StoreUint64(&pool.maxSeqNo, ev.SeqNo-1)
	pool.tempBlockNumber = ev.SeqNo
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Error("Get Database Instance Failed! error msg,", err.Error())
		return
	}
	block, err := core.GetBlockByNumber(db, ev.SeqNo - 1)
	if err != nil {
		return
	}
	// 2 revert state
	if err := pool.revertState(int64(tmpDemandNumber - 1), int64(ev.SeqNo - 1), block.MerkleRoot); err != nil {
		log.Errorf("revert state from %d to %d failed", tmpDemandNumber - 1, ev.SeqNo - 1)
		return
	}
	// 3. Delete related transaction, receipt, txmeta, and block itself in a specific range
	pool.removeDataInRange(ev.SeqNo, tmpDemandNumber)
	// 4. remove uncommitted data
	if err := pool.removeUncommittedData(); err != nil {
		log.Errorf("remove uncommitted during the state reset failed, revert state from %d to %d failed",
			tmpDemandNumber - 1, ev.SeqNo - 1)
		return
	}
	// 5. Reset chain
	isGenesis := (block.Number == 0)
	core.UpdateChain(db.NewBatch(), block, isGenesis, true, true)
	log.Debugf("revert state from %d to target %d success", tmpDemandNumber - 1, ev.SeqNo - 1)
	// 6. Told consensus reset finish
	msg := protos.VcResetDone{SeqNo: ev.SeqNo}
	pool.consenter.RecvLocal(msg)
}

// CutdownBlock remove a block and reset blockchain status to the last status.
func (pool *BlockPool) CutdownBlock(number uint64) {
	// 1. reset demand number  demand seqNo and maxSeqNo
	atomic.StoreUint64(&pool.demandNumber, number)
	atomic.StoreUint64(&pool.demandSeqNo, number)
	atomic.StoreUint64(&pool.maxSeqNo, number - 1)
	pool.tempBlockNumber = number
	// 2. revert state
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Error("Get Database Instance Failed! error msg,", err.Error())
		return
	}
	block, err := core.GetBlockByNumber(db, number - 1)
	if err != nil {
		return
	}
	pool.revertState(int64(number), int64(number - 1), block.MerkleRoot)
	// 3. remove block releted data
	pool.removeDataInRange(number, number + 1)
	// 4. remove uncommitted data
	if err := pool.removeUncommittedData(); err != nil {
		log.Errorf("remove uncommitted of %d failed", number)
		return
	}
	// 5. reset chain data
	core.UpdateChainByBlcokNum(db, block.Number, true, true)
}

// removeDataInRange remove transaction receipt txmeta and block itself in a specific range
// range is [from, to).
func (pool *BlockPool) removeDataInRange(from, to uint64) {
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Error("Get Database Instance Failed! error msg,", err.Error())
		return
	}
	// delete tx, txmeta and receipt
	for i := from; i < to; i += 1 {
		block, err := core.GetBlockByNumber(db, i)
		if err != nil {
			log.Errorf("miss block %d ,error msg %s", i, err.Error())
			continue
		}

		for _, tx := range block.Transactions {
			if err := core.DeleteTransaction(db, tx.GetTransactionHash().Bytes()); err != nil {
				log.Errorf("delete useless tx in block %d failed, error msg %s", i, err.Error())
			}
			if err := core.DeleteReceipt(db, tx.GetTransactionHash().Bytes()); err != nil {
				log.Errorf("delete useless receipt in block %d failed, error msg %s", i, err.Error())
			}
			if err := core.DeleteTransactionMeta(db, tx.GetTransactionHash().Bytes()); err != nil {
				log.Errorf("delete useless txmeta in block %d failed, error msg %s", i, err.Error())
			}
		}
		// delete block
		if err := core.DeleteBlockByNum(db, i); err != nil {
			log.Errorf("ViewChange, delete useless block %d failed, error msg %s", i, err.Error())
		}
	}
}


// revertState revert state from currentNumber related status to a target
// different process logic of different state implement
// undo from currentNumber -> targetNumber + 1.
func (pool *BlockPool) revertState(currentNumber int64, targetNumber int64, targetRootHash []byte) error {
	switch pool.conf.StateType {
	case "hyperstate":
		db, err := hyperdb.GetLDBDatabase()
		if err != nil {
			log.Error("get database handler failed")
			return err
		}
		dirtyStateObjectSet := mapset.NewSet()
		// get latest state instance
		latestBlock, err := core.GetBlockByNumber(db, uint64(currentNumber))
		if err != nil {
			log.Errorf("get latest block = #%d failed.", currentNumber)
			return err
		}
		state, err := pool.GetStateInstance(common.BytesToHash(latestBlock.MerkleRoot), db)
		if err != nil {
			log.Errorf("get latest state = #%d failed.", currentNumber)
			return err
		}
		// revert state change with changeset [targetNumber+1, currentNumber]
		// IMPORTANT undo changes in reverse
		for i := currentNumber; i >= targetNumber + 1; i -= 1 {
			log.Debugf("undo changes for #%d", i)
			j, err := db.Get(hyperstate.CompositeJournalKey(uint64(i)))
			if err != nil {
				log.Warningf("get journal in database for #%d failed. make sure #%d doesn't have state change", i, i)
				continue
			}
			journal, err := hyperstate.UnmarshalJournal(j)
			if err != nil {
				log.Errorf("unmarshal journal for #%d failed", i)
				continue
			}
			// IMPORTANT undo journal in reverse
			for j := len(journal.JournalList) - 1; j >= 0; j -= 1 {
				log.Debugf("journal %s", journal.JournalList[j].String())
				tmpState := state.(*hyperstate.StateDB)
				journal.JournalList[j].Undo(tmpState, true)
				if journal.JournalList[j].GetType() == hyperstate.StorageHashChangeType {
					tmp := journal.JournalList[j].(*hyperstate.StorageHashChange)
					// use struct instead of pointer
					dirtyStateObjectSet.Add(*tmp.Account)
				}
			}
			// remove persisted journals
			db.Delete(hyperstate.CompositeJournalKey(uint64(i)))
		}
		log.Debugf("revert to #%d, %s", targetNumber, string(state.Dump()))

		// revert related stateObject storage bucket tree
		for addr := range dirtyStateObjectSet.Iter() {
			address := addr.(common.Address)
			prefix, _ := hyperstate.CompositeStorageBucketPrefix(address.Bytes())
			bucketTree := bucket.NewBucketTree(string(prefix))
			bucketTree.Initialize(hyperstate.SetupBucketConfig(pool.bucketTreeConf.StorageSize, pool.bucketTreeConf.StorageLevelGroup))
			bucketTree.RevertToTargetBlock(big.NewInt(currentNumber), big.NewInt(targetNumber))
			hash, _ := bucketTree.ComputeCryptoHash()
			log.Debugf("re-compute %s storage hash %s", address.Hex(), common.Bytes2Hex(hash))
			obj, err := db.Get(hyperstate.CompositeAccountKey(address.Bytes()))
			if err != nil {
				log.Debugf("missing state object %s, it may be deleted", address.Hex())
				continue
			}
			account := &hyperstate.Account{}
			err = hyperstate.Unmarshal(obj, account)
			if err != nil {
				log.Errorf("unmarshal state object %s failed", address.Hex())
				continue
			}
			if bytes.Compare(account.Root.Bytes(), hash) != 0 {
				log.Errorf("after revert to #%d, state object %s revert failed, required storage hash %s, got %s",
					targetNumber, address.Hex(), account.Root.Hex(), common.Bytes2Hex(hash))
			}
		}
		// revert state bucket tree
		tree := state.GetTree()
		bucketTree := tree.(*bucket.BucketTree)
		bucketTree.Initialize(hyperstate.SetupBucketConfig(pool.bucketTreeConf.StorageSize, pool.bucketTreeConf.StorageLevelGroup))
		bucketTree.RevertToTargetBlock(big.NewInt(currentNumber), big.NewInt(targetNumber))
		currentRootHash, err := bucketTree.ComputeCryptoHash()
		if err != nil {
			log.Errorf("re-compute state bucket tree hash failed, error :%s", err.Error())
			return err
		}
		if bytes.Compare(currentRootHash, targetRootHash) != 0 {
			log.Errorf("revert to a different state, required %s, but current state %s",
				common.Bytes2Hex(targetRootHash), common.Bytes2Hex(currentRootHash))
			return errors.New("revert state failed")
		}
		// revert state instance oldest and root
		state.ResetToTarget(uint64(targetNumber+1), common.BytesToHash(targetRootHash))
		log.Noticef("revert state from #%d to #%d success", currentNumber, targetNumber)
	case "rawstate":
		// there is no need to revert state, because PMT can assure the correction
		pool.lastValidationState.Store(common.BytesToHash(targetRootHash))
	}
	return nil
}

// removeUncommittedData remove uncommitted validation result avoid of memory leak.
func (pool *BlockPool) removeUncommittedData() error {
	switch pool.conf.StateType {
	case "hyperstate":
		db, err := hyperdb.GetLDBDatabase()
		if err != nil {
			log.Error("get database failed")
			return err
		}
		state, _ := pool.GetStateInstance(common.Hash{}, db)
		state.Purge()
	case "rawstate":
		db, err := hyperdb.GetLDBDatabase()
		if err != nil {
			log.Error("get database failed")
			return err
		}
		keys := pool.blockCache.Keys()
		for _, key := range keys {
			ret, _ := pool.blockCache.Get(key)
			if ret == nil {
				continue
			}
			record := ret.(BlockRecord)
			for i, tx := range record.ValidTxs {
				if err := core.DeleteTransaction(db, tx.GetTransactionHash().Bytes()); err != nil {
					log.Errorf("ViewChange, delete useless tx in cache %d failed, error msg %s", i, err.Error())
				}

				if err := core.DeleteReceipt(db, tx.GetTransactionHash().Bytes()); err != nil {
					log.Errorf("ViewChange, delete useless receipt in cache %d failed, error msg %s", i, err.Error())
				}
				if err := core.DeleteTransactionMeta(db, tx.GetTransactionHash().Bytes()); err != nil {
					log.Errorf("ViewChange, delete useless txmeta in cache %d failed, error msg %s", i, err.Error())
				}
			}
		}
		// clear cache, all data in cache is useless because consensus module will resend those validation event
		// IMPORTANT
		// if validation cache is not clear, new validation event could be ignored, which leads to some event
		// will never be execute!
		pool.blockCache.Purge()
		// 4. Purge validationQueue
		pool.validationQueue.Purge()
	}
	return nil
}