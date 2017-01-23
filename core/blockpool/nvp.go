package blockpool

import (
	"hyperchain/core/types"
	"github.com/golang/protobuf/proto"
	"hyperchain/event"
	"hyperchain/common"
	"hyperchain/core"
	"bytes"
	"errors"
)

// TransitVerifiedBlock - transit a verified block to non-verified peers.
func (pool *BlockPool) TransitVerifiedBlock(block *types.Block) {
	data, err := proto.Marshal(block)
	if err != nil {
		log.Errorf("marshal verified block for #%d failed.", block.Number)
		return
	}
	go pool.helper.msgQ.Post(event.VerifiedBlock{
		Payload: data,
	})
}

// ReceiveVerifiedBlock - receive verified block for vp.
// process in serial but may out of order.
func (pool *BlockPool) ReceiveVerifiedBlock(ev event.ReceiveVerifiedBlock) {
	block := &types.Block{}
	err := proto.Unmarshal(ev.Payload, block)
	if err != nil {
		log.Errorf("receive invalid verified block, unmarshal failed.")
		return
	}
	log.Debugf("receive verified block #%d", block.Number)
	if block.Number < pool.demandNumber {
		log.Debugf("receive verified block #%d less than demand number, just ignore.", block.Number)
		return
	} else if block.Number == pool.demandNumber {
		if err := pool.applyVerifiedBlock(block); err != nil {
			log.Errorf("apply verified block #%d failed. %s", block.Number, err)
			return
		}
		if err := pool.applyRemainVerifiedBlock(block.Number); err != nil {
			log.Errorf("apply remain verified block failed. %s", err)
			return
		}
	} else {
		log.Debugf("receive verified block #%d larger than demand number %d, save into cache temporarily", block.Number, pool.demandNumber)
		pool.queue.Add(block.Number, block)
	}
}

// applyVerifiedBlock - apply a verified block to current state and update chain if successfully.
func (pool *BlockPool) applyVerifiedBlock(block *types.Block) error {
	if err := pool.processVerifiedBlock(block); err != nil {
		return err
	}
	judge := func(key interface{}, iterKey interface{}) bool {
		id := key.(uint64)
		iterId := iterKey.(uint64)
		if id >= iterId {
			return true
		}
		return false
	}
	pool.demandNumber = block.Number + 1
	log.Debugf("next demand verified block number #%d", block.Number + 1)
	pool.queue.RemoveWithCond(block.Number, judge)
	return nil
}

// applyRemainVerifiedBlock - apply remain verified block in cache which received earlier than expected.
func (pool *BlockPool) applyRemainVerifiedBlock(number uint64) error {
	remain := number + 1
	for {
		if ret, existed := pool.queue.Get(remain); existed == true {
			block := ret.(*types.Block)
			if err := pool.applyVerifiedBlock(block); err != nil {
				return err
			}
			remain = remain + 1
		} else {
			break
		}
	}
	return nil
}

// processVerifiedBlock - execute a verified block in virtual machine and flush all state changes to disk.
func (pool *BlockPool) processVerifiedBlock(block *types.Block) error {
	// initialize calculator
	// for calculate fingerprint of a batch of transactions and receipts
	if err := pool.initializeTransactionCalculator(); err != nil {
		log.Errorf("apply verified block #%d initialize transaction calculator", pool.tempBlockNumber)
		return err
	}
	if err := pool.initializeReceiptCalculator(); err != nil {
		log.Errorf("appy verified block #%d initialize receipt calculator", pool.tempBlockNumber)
		return err
	}
	state, err := pool.GetStateInstance()
	if err != nil {
		log.Errorf("get state db failed, %s", err)
		return err
	}
	batch := state.FetchBatch(block.Number)
	state.MarkProcessStart(block.Number)
	// initialize execution environment rule set
	env := initEnvironment(state, block.Number)
	// execute transaction one by one
	for i, tx := range block.Transactions {
		state.StartRecord(tx.GetTransactionHash(), common.BytesToHash(block.BlockHash), i)
		receipt, _, _, err := core.ExecTransaction(tx, env)
		// just ignore invalid transactions
		if err != nil {
			log.Warning("invalid transaction found during the nvp synchronization process in #%d", block.Number)
			continue
		}
		pool.calculateTransactionsFingerprint(tx, false)
		pool.calculateReceiptFingerprint(receipt, false)
		// persist transaction here
		if err, _ := core.PersistTransaction(batch, tx, pool.GetTransactionVersion(), false, false); err != nil {
			log.Errorf("persist transaction for index %d in #%d failed.", i, block.Number)
			continue
		}
		// persist transaction meta data
		meta := &types.TransactionMeta{
			BlockIndex: block.Number,
			Index:      int64(i),
		}
		if err := core.PersistTransactionMeta(batch, meta, tx.GetTransactionHash(), false, false); err != nil {
			log.Errorf("persist transaction meta for index %d in #%d failed.", i, block.Number)
			continue
		}
		// persist receipt
		if err, _ := core.PersistReceipt(batch, receipt, pool.GetTransactionVersion(), false, false); err != nil {
			log.Errorf("persist receipt for index %d in #%d failed.", i, block.Number)
			continue
		}
	}
	// submit validation result
	err, merkleRoot, txRoot, receiptRoot := pool.submitValidationResult(state, batch)
	if err != nil {
		log.Error("submit validation result failed.", err.Error())
		return err
	}
	if pool.assertApplyVerifiedBlockValidation(merkleRoot, txRoot, receiptRoot, block) == false {
		log.Errorf("apply verified block #%d got different status", block.Number)
		return errors.New("apply verified block failed")
	}
	// generate new state fingerprint
	log.Debugf("apply verified block #%d, merkle root [%s],  transaction root [%s],  receipt root [%s]",
		block.Number, common.Bytes2Hex(merkleRoot), common.Bytes2Hex(txRoot), common.Bytes2Hex(receiptRoot))
	// persist block
	if err, _ := core.PersistBlock(batch, block, pool.GetBlockVersion(), false, false); err != nil {
		log.Errorf("persist block #%d into database failed! error msg, ", block.Number, err.Error())
		return err
	}
	// persist chain
	core.UpdateChain(batch, block, false, false, false)
	batch.Write()
	state.MarkProcessFinish(block.Number)
	return nil
}

// assertApplyVerifiedBlockValidation - check the execution result with given result in block(`merkle root`), return false if not equal.
func (pool *BlockPool) assertApplyVerifiedBlockValidation(merkleRoot, txRoot, receiptRoot []byte, block *types.Block) bool {
	if bytes.Compare(block.MerkleRoot, merkleRoot) != 0 {
		log.Warningf("mismatch in block merkle root  of #%d, required %s, got %s",
			block.Number, common.Bytes2Hex(block.MerkleRoot), common.Bytes2Hex(merkleRoot))
		return false
	}
	if bytes.Compare(block.TxRoot, txRoot) != 0 {
		log.Warningf("mismatch in block transaction root  of #%d, required %s, got %s",
			block.Number, common.Bytes2Hex(block.TxRoot), common.Bytes2Hex(txRoot))
		return false

	}
	if bytes.Compare(block.ReceiptRoot, receiptRoot) != 0 {
		log.Warningf("mismatch in block receipt root  of #%d, required %s, got %s",
			block.Number, common.Bytes2Hex(block.ReceiptRoot), common.Bytes2Hex(receiptRoot))
		return false
	}
	return true
}
