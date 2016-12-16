package blockpool

import (
	"sync/atomic"
	"hyperchain/hyperdb"
	"hyperchain/event"
	"hyperchain/common"
	"hyperchain/core"
)


// reset blockchain to a stable checkpoint status when `viewchange` occur
func (pool *BlockPool) ResetStatus(ev event.VCResetEvent) {
	tmpDemandNumber := atomic.LoadUint64(&pool.demandNumber)
	// 1. Reset demandNumber , demandSeqNo and lastValidationState
	atomic.StoreUint64(&pool.demandNumber, ev.SeqNo)
	atomic.StoreUint64(&pool.demandSeqNo, ev.SeqNo)
	atomic.StoreUint64(&pool.maxSeqNo, ev.SeqNo-1)

	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Error("Get Database Instance Failed! error msg,", err.Error())
		return
	}

	block, err := core.GetBlockByNumber(db, ev.SeqNo - 1)
	if err != nil {
		return
	}
	pool.lastValidationState.Store(common.BytesToHash(block.MerkleRoot))
	// 2. Delete related transaction, receipt, txmeta, and block itself in a specific range
	pool.removeDataInRange(ev.SeqNo, tmpDemandNumber)

	// 3. Delete from blockcache
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
	// 5. Reset chain
	isGenesis := (block.Number == 0)
	core.UpdateChain(block, isGenesis)
}

// remove a block and reset blockchain status to the last status
func (pool *BlockPool) CutdownBlock(number uint64) {
	// 1. reset demand number and demand seqNo
	atomic.StoreUint64(&pool.demandNumber, number)
	atomic.StoreUint64(&pool.demandSeqNo, number)
	atomic.StoreUint64(&pool.maxSeqNo, number - 1)
	// 2. remove block releted data
	pool.removeDataInRange(number, number + 1)
	// 3. reset state root hash
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Error("Get Database Instance Failed! error msg,", err.Error())
		return
	}
	block, err := core.GetBlockByNumber(db, number - 1)
	if err != nil {
		log.Errorf("miss block %d ,error msg %s", number - 1, err.Error())
		return
	}
	pool.lastValidationState.Store(common.BytesToHash(block.MerkleRoot))
	// 4. reset chain data
	core.UpdateChainByBlcokNum(db, block.Number)
}

// remove transaction receipt txmeta and block itself in a specific range
// range is [from, to)
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


