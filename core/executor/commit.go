package executor

import (
	"encoding/hex"
	"github.com/golang/protobuf/proto"
	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/event"
	"hyperchain/hyperdb"
	"hyperchain/p2p"
	"hyperchain/protos"
	"time"
	"hyperchain/hyperdb/db"
	edb "hyperchain/core/db_utils"

)

func (executor *Executor) CommitBlock(ev event.CommitOrRollbackBlockEvent, peerManager p2p.PeerManager) {
	executor.addCommitEvent(ev)
	if executor.peerManager == nil {
		executor.peerManager = peerManager
	}
}

func (executor *Executor) listenCommitEvent() {
	for {
		ev := executor.fetchCommitEvent()
		if success := executor.processCommitEvent(ev, executor.processCommitDone); success == false {
			log.Errorf("[Namespace = %s] commit block #%d failed, system crush down.", executor.namespace, ev.SeqNo)
		}
	}
}

// processCommitEvent - consume commit event from channel.
func (executor *Executor) processCommitEvent(ev event.CommitOrRollbackBlockEvent, done func()) bool {
	executor.markCommitBusy()
	defer executor.markCommitIdle()
	defer done()
	if !executor.commitValidationCheck(ev) {
		log.Errorf("[Namespace = %s] commit event %d not satisfy the demand", executor.namespace, ev.SeqNo)
		return false
	}
	block := executor.constructBlock(ev)
	if block == nil {
		log.Errorf("[Namespace = %s] construct new block for %d commit event failed.", executor.namespace, ev.SeqNo)
		return false
	}
	record := executor.getValidateRecord(ev.Hash)
	if record == nil {
		log.Errorf("[Namespace = %s] no validation record for #%d found", executor.namespace, ev.SeqNo)
		return false
	}
	if err := executor.writeBlock(block, record); err != nil {
		log.Errorf("write block for #%d failed. err %s", ev.SeqNo, err.Error())
		return false
	}
	// throw all invalid transactions back.
	if ev.IsPrimary {
		executor.throwInvalidTransactionBack(record.InvalidTxs)
	}
	executor.incDemandNumber()
	executor.cache.validationResultCache.Remove(ev.Hash)
	return true
}

// writeBlock - flush a block into disk.
func (executor *Executor) writeBlock(block *types.Block, record *ValidationResultRecord) error {
	batch := executor.statedb.FetchBatch(record.SeqNo)
	if err := executor.persistTransactions(batch, block.Transactions, block.Number); err != nil {
		log.Errorf("[Namespace = %s] persist transactions of #%d failed.", executor.namespace, block.Number)
		return err
	}
	if err := executor.persistReceipts(batch, record.Receipts, block.Number, common.BytesToHash(block.BlockHash)); err != nil {
		log.Errorf("[Namespace = %s] persist receipts of #%d failed.", executor.namespace, block.Number)
		return err
	}
	if err, _ := edb.PersistBlock(batch, block, false, false); err != nil {
		log.Errorf("[Namespace = %s] persist block #%d into database failed.", executor.namespace, block.Number, err.Error())
		return err
	}
	edb.UpdateChain(executor.namespace, batch, block, false, false, false)
	batch.Write()
	executor.statedb.MarkProcessFinish(record.SeqNo)

	if block.Number % 10 == 0 && block.Number != 0 {
		edb.WriteChainChan(executor.namespace)
	}
	log.Noticef("[Namespace = %s] Block number %d", executor.namespace, block.Number)
	log.Noticef("[Namespace = %s] Block hash %s", executor.namespace, hex.EncodeToString(block.BlockHash))
	// remove Cached Transactions which used to check transaction duplication
	executor.informConsensus(CONSENSUS_LOCAL, protos.RemoveCache{Vid: record.VID})
	return nil
}

// getValidateRecord - get validate record with given hash identification.
// nil will be return if no record been found.
func (executor *Executor) getValidateRecord(hash string) *ValidationResultRecord {
	ret, existed := executor.fetchValidationResult(hash)
	if !existed {
		log.Noticef("[Namespace = %s] no validation result found when commit block, hash %s", executor.namespace, hash)
		return nil
	}
	return ret
}

// generateBlock - generate a block with given data.
func (executor *Executor) constructBlock(ev event.CommitOrRollbackBlockEvent) *types.Block {
	record := executor.getValidateRecord(ev.Hash)
	if record == nil {
		return nil
	}
	// 1.generate a new block with the argument in cache
	newBlock := &types.Block{
		ParentHash:  edb.GetLatestBlockHash(executor.namespace),
		MerkleRoot:  record.MerkleRoot,
		TxRoot:      record.TxRoot,
		ReceiptRoot: record.ReceiptRoot,
		Timestamp:   ev.Timestamp,
		CommitTime:  ev.Timestamp,
		Number:      ev.SeqNo,
		WriteTime:   time.Now().UnixNano(),
		EvmTime:     time.Now().UnixNano(),
	}
	newBlock.Transactions = make([]*types.Transaction, len(record.ValidTxs))
	copy(newBlock.Transactions, record.ValidTxs)
	newBlock.BlockHash = newBlock.Hash(executor.commonHash).Bytes()
	return newBlock
}

// commitValidationCheck - check whether this commit event satisfy demand.
func (executor *Executor) commitValidationCheck(ev event.CommitOrRollbackBlockEvent) bool {
	// 1. check whether this ev is the demand one
	if !executor.isDemandNumber(ev.SeqNo){
		log.Errorf("[Namespace = %s] receive a commit event %d which is not demand, drop it.", executor.namespace, ev.SeqNo)
		return false
	}
	// 2. check whether validation result exist
	record := executor.getValidateRecord(ev.Hash)
	if record == nil {
		return false
	}
	// 3. check whether ev's seqNo equal to record seqNo which act as block number
	vid := record.VID
	tempBlockNumber := record.SeqNo
	if tempBlockNumber != ev.SeqNo {
		log.Errorf("[Namespace = %s] miss match temp block number<#%d>and actually block number<#%d> for vid #%d validation. commit for block #%d failed",
			executor.namespace, tempBlockNumber, ev.SeqNo, vid, ev.SeqNo)
		return false
	}
	return true
}

func (executor *Executor) persistTransactions(batch db.Batch, transactions []*types.Transaction, blockNumber uint64) error {
	for i, transaction := range transactions {
		if err, _ := edb.PersistTransaction(batch, transaction, false, false); err != nil {
			return err
		}
		// persist transaction meta data
		meta := &types.TransactionMeta{
			BlockIndex: blockNumber,
			Index:      int64(i),
		}
		if err := edb.PersistTransactionMeta(batch, meta, transaction.GetTransactionHash(), false, false); err != nil {
			return err
		}
	}
	return nil
}

// re assign block hash and block number to transaction logs
// during the validation, block number and block hash can be incorrect
func (executor *Executor) persistReceipts(batch db.Batch, receipts []*types.Receipt, blockNumber uint64, blockHash common.Hash) error {
	for _, receipt := range receipts {
		logs, err := receipt.GetLogs()
		if err != nil {
			return err
		}
		for _, log := range logs {
			log.BlockHash = blockHash
			log.BlockNumber = blockNumber
		}
		receipt.SetLogs(logs)
		if err, _ := edb.PersistReceipt(batch, receipt, false, false); err != nil {
			return err
		}
	}
	return nil
}

// save the invalid transaction into database for client query
func (executor *Executor) StoreInvalidTransaction(ev event.RespInvalidTxsEvent) {
	invalidTx := &types.InvalidTransactionRecord{}
	err := proto.Unmarshal(ev.Payload, invalidTx)
	if err != nil {
		log.Error("unmarshal invalid transaction record payload failed")
	}
	// save to db
	log.Noticef("[Namespace = %s]invalid transaction", common.BytesToHash(invalidTx.Tx.TransactionHash).Hex())
	db, err := hyperdb.GetDBDatabaseByNamespace(executor.namespace)
	if err != nil {
		return
	}
	err, _ = edb.PersistInvalidTransactionRecord(db.NewBatch(), invalidTx, true, true)
	if err != nil {
		log.Error("save invalid transaction record failed,", err.Error())
		return
	}
}
