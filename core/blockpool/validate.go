package blockpool

import (
	"sync/atomic"
	"hyperchain/event"
	"hyperchain/crypto"
	"sort"
	"hyperchain/core/types"
	"hyperchain/recovery"
	"github.com/golang/protobuf/proto"
	"hyperchain/protos"
	"hyperchain/p2p"
	"sync"
	"hyperchain/hyperdb"
	"hyperchain/tree/pmt"
	"hyperchain/core"
	"hyperchain/core/vm/params"
	"strconv"
	"hyperchain/common"
	"errors"
)

// Validate is an entry of `validate process`
// When a validationEvent received, put it into the validationQueue
// If the demand ValidationEvent arrived, call `PreProcess` function
// IMPORTANT this function called in parallelly, Make sure all the variable are thread-safe
func (pool *BlockPool) Validate(validationEvent event.ExeTxsEvent, commonHash crypto.CommonHash, encryption crypto.Encryption, peerManager p2p.PeerManager) {
	if validationEvent.SeqNo > atomic.LoadUint64(&pool.maxSeqNo) {
		atomic.StoreUint64(&pool.maxSeqNo, validationEvent.SeqNo)
	}

	if _, existed := pool.validationQueue.Get(validationEvent.SeqNo); existed {
		log.Error("receive repeat validation event, ", validationEvent.SeqNo)
		return
	}

	// (1) Check SeqNo
	if validationEvent.SeqNo < atomic.LoadUint64(&pool.demandSeqNo) {
		// Receive repeat ValidationEvent
		log.Error("receive invalid validation event, seqno less than demandseqNo, ", validationEvent.SeqNo)
		return
	} else if validationEvent.SeqNo == atomic.LoadUint64(&pool.demandSeqNo) {
		// Process
		if _, success := pool.PreProcess(validationEvent, commonHash, encryption, peerManager); success {
			atomic.AddUint64(&pool.demandSeqNo, 1)
			log.Notice("current demandSeqNo is, ", pool.demandSeqNo)
		} else {
			log.Error("pre process failed")
			return
		}
		judge := func(key interface{}, iterKey interface{}) bool {
			id := key.(uint64)
			iterId := iterKey.(uint64)
			if id >= iterId {
				return true
			}
			return false
		}
		pool.validationQueue.RemoveWithCond(validationEvent.SeqNo, judge)

		// Process remain event
		for i := validationEvent.SeqNo + 1; i <= atomic.LoadUint64(&pool.maxSeqNo); i += 1 {
			if ret, existed := pool.validationQueue.Get(i); existed {
				ev := ret.(event.ExeTxsEvent)
				if _, success := pool.PreProcess(ev, commonHash, encryption, peerManager); success {
					pool.validationQueue.Remove(i)
					atomic.AddUint64(&pool.demandSeqNo, 1)
					log.Notice("Current demandSeqNo is, ", pool.demandSeqNo)
				} else {
					log.Error("PreProcess Failed")
					return
				}
			} else {
				break
			}
		}
		return
	} else {
		log.Notice("Receive ValidationEvent which is not demand, ", validationEvent.SeqNo, "save into cache temporarily")
		pool.validationQueue.Add(validationEvent.SeqNo, validationEvent)
	}
}

// Process an ValidationEvent
func (pool *BlockPool) PreProcess(validationEvent event.ExeTxsEvent, commonHash crypto.CommonHash, encryption crypto.Encryption, peerManager p2p.PeerManager) (error, bool) {
	var validTxSet []*types.Transaction
	var invalidTxSet []*types.InvalidTransactionRecord
	var index []int
	if validationEvent.IsPrimary {
		invalidTxSet, index = pool.checkSign(validationEvent.Transactions, commonHash, encryption)
	} else {
		validTxSet = validationEvent.Transactions
	}

	if len(index) > 0 {
		sort.Ints(index)
		count := 0
		set := validationEvent.Transactions
		for _, idx := range index {
			idx = idx-count
			if idx == 0 {
				set = set[1:]
				count++
			} else {
				set = append(set[:idx - 1], set[idx + 1:]...)
				count++
			}
		}
		validTxSet = set
	} else {
		validTxSet = validationEvent.Transactions
	}
	err, _, merkleRoot, txRoot, receiptRoot, validTxSet, invalidTxSet := pool.ProcessBlockInVm(validTxSet, invalidTxSet, validationEvent.SeqNo)
	if err != nil {
		log.Error("ProcessBlock Failed!, block number: ", validationEvent.SeqNo)
		return err, false
	}
	hash := commonHash.Hash([]interface{}{
		merkleRoot,
		txRoot,
		receiptRoot,
	})

	if len(validTxSet) != 0 {
		pool.blockCache.Add(hash.Hex(), BlockRecord{
			TxRoot:      txRoot,
			ReceiptRoot: receiptRoot,
			MerkleRoot:  merkleRoot,
			InvalidTxs:  invalidTxSet,
			ValidTxs:    validTxSet,
			SeqNo:       validationEvent.SeqNo,
		})
	}
	log.Info("Invalid Tx number: ", len(invalidTxSet))
	log.Info("Valid Tx number: ", len(validTxSet))
	// Communicate with PBFT
	pool.consenter.RecvLocal(protos.ValidatedTxs{
		Transactions: validTxSet,
		SeqNo:        validationEvent.SeqNo,
		View:         validationEvent.View,
		Hash:         hash.Hex(),
		Timestamp:    validationEvent.Timestamp,
	})

	// empty block generated, throw all invalid transactions back to original node directly
	if validationEvent.IsPrimary && len(validTxSet) == 0 {
		// 1. Remove all cached transaction in this block, because empty block won't enter network
		pool.consenter.RemoveCachedBatch(validationEvent.SeqNo)
		// 2. Throw all invalid transaction back to the origin node
		for _, t := range invalidTxSet {
			payload, err := proto.Marshal(t)
			if err != nil {
				log.Error("Marshal tx error")
			}
			if t.Tx.Id == uint64(peerManager.GetNodeId()) {
				pool.StoreInvalidResp(event.RespInvalidTxsEvent{
					Payload: payload,
				})
				continue
			}
			var peers []uint64
			peers = append(peers, t.Tx.Id)
			peerManager.SendMsgToPeers(payload, peers, recovery.Message_INVALIDRESP)
		}
	}
	return nil, true
}
// check the sender's signature of the transaction
func (pool *BlockPool) checkSign(txs []*types.Transaction, commonHash crypto.CommonHash, encryption crypto.Encryption) ([]*types.InvalidTransactionRecord, []int) {
	var invalidTxSet []*types.InvalidTransactionRecord
	// (1) check signature for each transaction
	var wg sync.WaitGroup
	var index []int
	var mu sync.Mutex
	for i := range txs {
		wg.Add(1)
		go func(i int){
			tx := txs[i]
			if !tx.ValidateSign(encryption, commonHash) {
				log.Notice("Validation, found invalid signature, send from :", tx.Id)
				invalidTxSet = append(invalidTxSet, &types.InvalidTransactionRecord{
					Tx:      tx,
					ErrType: types.InvalidTransactionRecord_SIGFAILED,
					ErrMsg:  []byte("Invalid signature"),
				})
				mu.Lock()
				index = append(index, i)
				mu.Unlock()
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	return invalidTxSet, index
	//return nil, nil
}


// Put all transactions into the virtual machine and execute
// Return the execution result, such as txs' merkle root, receipts' merkle root, accounts' merkle root and so on
func (pool *BlockPool) ProcessBlockInVm(txs []*types.Transaction, invalidTxs []*types.InvalidTransactionRecord, seqNo uint64) (error, []byte, []byte, []byte, []byte, []*types.Transaction, []*types.InvalidTransactionRecord) {
	var validtxs []*types.Transaction
	var (
		env = make(map[string]string)
	)
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		return err, nil, nil, nil, nil, nil, invalidTxs
	}
	txTrie, err := pmt.New(common.Hash{}, db)
	if err != nil {return err, nil, nil, nil, nil, nil, invalidTxs}
	receiptTrie, err := pmt.New(common.Hash{}, db)
	if err != nil {return err, nil, nil, nil, nil, nil, invalidTxs}
	v := pool.lastValidationState.Load()
	initStatus, ok := v.(common.Hash)
	if ok == false {
		return errors.New("Get StateDB Status Failed!"), nil, nil, nil, nil, nil, invalidTxs
	}

	state, err := pool.GetStateInstance(initStatus, db)
	log.Critical("Before", string(state.Dump()))
	if err != nil {return err, nil, nil, nil, nil, nil, invalidTxs}
	env["currentNumber"] = strconv.FormatUint(seqNo, 10)
	env["currentGasLimit"] = "10000000"
	vmenv := core.NewEnvFromMap(core.RuleSet{params.MainNetHomesteadBlock, params.MainNetDAOForkBlock, true}, state, env)

	public_batch := db.NewBatch()
	for i, tx := range txs {
		state.StartRecord(tx.GetTransactionHash(), common.Hash{}, i)
		receipt, _, _, err := core.ExecTransaction(*tx, vmenv)
		if err != nil{
			var errType types.InvalidTransactionRecord_ErrType
			if core.IsValueTransferErr(err) {
				errType = types.InvalidTransactionRecord_OUTOFBALANCE
			} else if core.IsExecContractErr(err) {
				tmp := err.(*core.ExecContractError)
				if tmp.GetType() == 0 {
					errType = types.InvalidTransactionRecord_DEPLOY_CONTRACT_FAILED
				} else if tmp.GetType() == 1{
					errType = types.InvalidTransactionRecord_INVOKE_CONTRACT_FAILED
				} else {
					// For extension
				}
			} else {
				// For extension
			}
			invalidTxs = append(invalidTxs, &types.InvalidTransactionRecord{
				Tx:      tx,
				ErrType: errType,
				ErrMsg:  []byte(err.Error()),
			})
			continue
		}
		var data []byte
		// Persist transaction
		err, data = core.PersistTransaction(public_batch, tx, pool.conf.TransactionVersion, false, false);
		if err != nil {
			log.Error("Put tx data into database failed! error msg, ", err.Error())
			return err, nil, nil, nil, nil, nil, invalidTxs
		}
		txTrie.Update(append(core.TransactionPrefix, tx.GetTransactionHash().Bytes()...), data)

		// Persist receipt
		err, data = core.PersistReceipt(public_batch, receipt, pool.conf.TransactionVersion, false, false)
		if  err != nil {
			log.Error("Put receipt data into database failed! error msg, ", err.Error())
			return err, nil, nil, nil, nil, nil, invalidTxs
		}
		receiptTrie.Update(append(core.ReceiptsPrefix, receipt.TxHash...), data)

		// Persist transaction meta
		// set temporarily
		// for primary node, the seqNo can be invalid. remove the incorrect txmeta info when commit block to avoid this error
		meta := &types.TransactionMeta{
			BlockIndex: seqNo,
			Index:      int64(i),
		}
		if err := core.PersistTransactionMeta(public_batch, meta, tx.GetTransactionHash(), false, false); err != nil {
			log.Error("Put txmeta into database failed! error msg, ", err.Error())
			return err, nil, nil, nil, nil, nil, invalidTxs
		}

		validtxs = append(validtxs, tx)
	}
	root, err := state.Commit()
	log.Critical("After", string(state.Dump()))
	if err != nil {
		log.Error("Commit state db failed! error msg, ", err.Error())
		return err, nil, nil, nil, nil, nil, invalidTxs
	}
	merkleRoot := root.Bytes()
	txRoot := txTrie.Hash().Bytes()
	receiptRoot := receiptTrie.Hash().Bytes()
	pool.lastValidationState.Store(root)
	// IMPORTANT never forget to call batch.Write, otherwise, all data in batch will be lost
	go public_batch.Write()
	log.Notice("MERKLE ROOT", root.Hex())
	log.Notice("TX ROOT", common.Bytes2Hex(txRoot))
	log.Notice("RECEIPT ROOT", common.Bytes2Hex(receiptRoot))
	return nil, nil, merkleRoot, txRoot, receiptRoot, validtxs, invalidTxs
}
