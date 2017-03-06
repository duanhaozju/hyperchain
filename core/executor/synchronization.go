package executor

import (
	"hyperchain/common"
	"hyperchain/core"
	"hyperchain/core/types"
	"errors"
	"hyperchain/tree/bucket"
)

// ApplyBlock - apply all transactions in block into state during the `state update` process.
func (executor *Executor) ApplyBlock(block *types.Block, seqNo uint64) (error, *BlockRecord) {
	if block.Transactions == nil {
		return errors.New("empty block"), nil
	}
	return executor.applyBlock(block, seqNo)
}

func (executor *Executor) applyBlock(block *types.Block, seqNo uint64) (error, *BlockRecord) {
	// initialize calculator
	// for calculate fingerprint of a batch of transactions and receipts
	if err := executor.initializeTransactionCalculator(); err != nil {
		log.Errorf("validate #%d initialize transaction calculator", executor.tempBlockNumber)
		return err, nil
	}
	if err := executor.initializeReceiptCalculator(); err != nil {
		log.Errorf("validate #%d initialize receipt calculator", executor.tempBlockNumber)
		return err, nil
	}
	// load latest state fingerprint
	// for compatibility, doesn't remove the statement below
	// initialize state
	state, err := executor.GetStateInstance()
	if err != nil {
		return err, nil
	}
	state.Purge()

	tree := state.GetTree()
	bucketTree := tree.(*bucket.BucketTree)
	bucketTree.ClearAllCache()

	batch := state.FetchBatch(seqNo)
	state.MarkProcessStart(executor.tempBlockNumber)
	// initialize execution environment rule set
	env := initEnvironment(state, executor.tempBlockNumber)
	// execute transaction one by one
	for i, tx := range block.Transactions {
		state.StartRecord(tx.GetTransactionHash(), common.Hash{}, i)
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
		if err, _ := core.PersistTransaction(batch, tx, executor.GetTransactionVersion(), false, false); err != nil {
			log.Errorf("persist transaction for index %d in #%d failed.", i, seqNo)
			continue
		}
		// persist transaction meta data
		meta := &types.TransactionMeta{
			BlockIndex: seqNo,
			Index:      int64(i),
		}
		if err := core.PersistTransactionMeta(batch, meta, tx.GetTransactionHash(), false, false); err != nil {
			log.Errorf("persist transaction meta for index %d in #%d failed.", i, seqNo)
			continue
		}
		// persist receipt
		if err, _ := core.PersistReceipt(batch, receipt, executor.GetTransactionVersion(), false, false); err != nil {
			log.Errorf("persist receipt for index %d in #%d failed.", i, seqNo)
			continue
		}
	}
	// submit validation result
	err, merkleRoot, txRoot, receiptRoot := executor.submitValidationResult(state, batch)
	if err != nil {
		log.Error("submit validation result failed.", err.Error())
		return err, nil
	}
	// generate new state fingerprint
	// IMPORTANT doesn't call batch.Write util recv commit event for atomic assurance
	log.Debugf("validate result temp block number #%d, vid #%d, merkle root [%s],  transaction root [%s],  receipt root [%s]",
		executor.tempBlockNumber, seqNo, common.Bytes2Hex(merkleRoot), common.Bytes2Hex(txRoot), common.Bytes2Hex(receiptRoot))
	return nil, &BlockRecord{
		TxRoot:      txRoot,
		ReceiptRoot: receiptRoot,
		MerkleRoot:  merkleRoot,
	}
}

// ClearStateUnCommitted - remove all cached stuff
func (executor *Executor) ClearStateUnCommitted() {
	switch executor.GetStateType() {
	case "hyperstate":
		state, err := executor.GetStateInstance()
		if err != nil {
			return
		}
		state.Purge()
	case "rawstate":

	}
}

// SubmitForStateUpdate - submit all changes in `state update` process
func (executor *Executor) SubmitForStateUpdate(seqNo uint64) error {
	switch executor.GetStateType() {
	case "rawstate":

	case "hyperstate":
		state, err := executor.GetStateInstance()
		if err != nil {
			log.Errorf("submit for state update #%d failed", seqNo)
			return err
		}
		batch := state.FetchBatch(seqNo)
		core.UpdateChainByBlcokNum(batch, seqNo, false, false)
		batch.Write()
		state.MarkProcessFinish(seqNo)
		return nil
	default:
		return nil
	}
	return nil
}
