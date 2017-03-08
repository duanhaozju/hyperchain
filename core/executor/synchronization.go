package executor

import (
	"hyperchain/common"
	"hyperchain/core"
	"hyperchain/core/types"
	"hyperchain/tree/bucket"
	edb "hyperchain/core/db_utils"
)

// ApplyBlock - apply all transactions in block into state during the `state update` process.
func (executor *Executor) ApplyBlock(block *types.Block, seqNo uint64) (error, *ValidationResultRecord) {
	if block.Transactions == nil {
		return EmptyPointerErr, nil
	}
	return executor.applyBlock(block, seqNo)
}

func (executor *Executor) applyBlock(block *types.Block, seqNo uint64) (error, *ValidationResultRecord) {
	// initialize calculator
	// for calculate fingerprint of a batch of transactions and receipts
	executor.initTransactionHashCalculator()
	executor.initReceiptHashCalculator()
	// load latest state fingerprint
	// for compatibility, doesn't remove the statement below
	// initialize state
	executor.statedb.Purge()

	tree := executor.statedb.GetTree()
	bucketTree := tree.(*bucket.BucketTree)
	bucketTree.ClearAllCache()

	batch := executor.statedb.FetchBatch(seqNo)
	executor.statedb.MarkProcessStart(executor.getTempBlockNumber())
	// initialize execution environment rule set
	env := initEnvironment(executor.statedb, executor.getTempBlockNumber())
	// execute transaction one by one
	for i, tx := range block.Transactions {
		executor.statedb.StartRecord(tx.GetTransactionHash(), common.Hash{}, i)
		receipt, _, _, err := core.ExecTransaction(tx, env)
		// just ignore invalid transactions
		if err != nil {
			log.Warning("invalid transaction found during the state update process in #%d", seqNo)
			continue
		}
		executor.calculateTransactionsFingerprint(tx, false)
		executor.calculateReceiptFingerprint(receipt, false)

		// different with normal process, because during the state update, block number and seqNo are always same
		// persist transaction here
		if err, _ := edb.PersistTransaction(batch, tx, false, false); err != nil {
			log.Errorf("persist transaction for index %d in #%d failed.", i, seqNo)
			continue
		}
		// persist transaction meta data
		meta := &types.TransactionMeta{
			BlockIndex: seqNo,
			Index:      int64(i),
		}
		if err := edb.PersistTransactionMeta(batch, meta, tx.GetTransactionHash(), false, false); err != nil {
			log.Errorf("persist transaction meta for index %d in #%d failed.", i, seqNo)
			continue
		}
		// persist receipt
		if err, _ := edb.PersistReceipt(batch, receipt, false, false); err != nil {
			log.Errorf("persist receipt for index %d in #%d failed.", i, seqNo)
			continue
		}
	}
	// submit validation result
	err, merkleRoot, txRoot, receiptRoot := executor.submitValidationResult(batch)
	if err != nil {
		log.Error("submit validation result failed.", err.Error())
		return err, nil
	}
	// generate new state fingerprint
	// IMPORTANT doesn't call batch.Write util recv commit event for atomic assurance
	log.Debugf("validate result temp block number #%d, vid #%d, merkle root [%s],  transaction root [%s],  receipt root [%s]",
		executor.getTempBlockNumber(), seqNo, common.Bytes2Hex(merkleRoot), common.Bytes2Hex(txRoot), common.Bytes2Hex(receiptRoot))
	return nil, &ValidationResultRecord{
		TxRoot:      txRoot,
		ReceiptRoot: receiptRoot,
		MerkleRoot:  merkleRoot,
	}
}

// ClearStateUnCommitted - remove all cached stuff
func (executor *Executor) ClearStateUnCommitted() {
	executor.statedb.Purge()
}

// SubmitForStateUpdate - submit all changes in `state update` process
func (executor *Executor) SubmitForStateUpdate(seqNo uint64) error {
	batch := executor.statedb.FetchBatch(seqNo)
	edb.UpdateChainByBlcokNum(executor.namespace, batch, seqNo, false, false)
	batch.Write()
	executor.statedb.MarkProcessFinish(seqNo)
	return nil
}
