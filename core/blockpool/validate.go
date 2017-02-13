package blockpool

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"hyperchain/common"
	"hyperchain/core"
	"hyperchain/core/types"
	"hyperchain/core/vm"
	"hyperchain/core/vm/params"
	"hyperchain/crypto"
	"hyperchain/event"
	"hyperchain/hyperdb"
	"hyperchain/p2p"
	"hyperchain/protos"
	"hyperchain/recovery"
	"hyperchain/tree/pmt"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
)

func (pool *BlockPool) Validate(validationEvent event.ExeTxsEvent, peerManager p2p.PeerManager) {
	pool.validateQueue <- validationEvent
	if pool.peerManager == nil {
		pool.peerManager = peerManager
	}
	atomic.AddInt32(&pool.validateQueueLen, 1)
}
func (pool *BlockPool) validateBackendLoop() {
	for ev := range pool.validateQueue {
		if atomic.LoadInt32(&pool.validateBehaveFlag) == VALIDATEBEHAVETYPE_NORMAL {
			if success := pool.consumeValidateEvent(ev); success == false {
				log.Errorf("validate #%d failed, system crush down.", ev.SeqNo)
			}
			atomic.AddInt32(&pool.validateQueueLen, -1)
		} else {
			log.Noticef("drop validation event %d", ev.SeqNo)
			pool.dropValdiateEvent(ev)
			atomic.AddInt32(&pool.validateQueueLen, -1)
		}
	}
}

func (pool *BlockPool) consumeValidateEvent(validationEvent event.ExeTxsEvent) bool {
	atomic.StoreInt32(&pool.validateInProgress, PROGRESS_TRUE)
	defer atomic.StoreInt32(&pool.validateInProgress, PROGRESS_FALSE)
	if pool.validateEventCheck(validationEvent) == false {
		log.Errorf("receive validation event %d while %d is required, save into cache temporarily.", validationEvent.SeqNo, pool.demandSeqNo)
		pool.validateEventQueue.Add(validationEvent.SeqNo, validationEvent)
		return true
	}
	if _, success := pool.PreProcess(validationEvent); success == false {
		return false
	}
	pool.increaseDemandSeqNo()
	if pool.validateEventQueue.Len() > 0 {
		// there is still some events remain.
		for  {
			if pool.validateEventQueue.Contains(pool.demandSeqNo) {
				res, _ := pool.validateEventQueue.Get(pool.demandSeqNo)
				validationEvent = res.(event.ExeTxsEvent)
				if _, success := pool.PreProcess(validationEvent); success == false {
					return false
				} else {
					pool.increaseDemandSeqNo()
					judge := func(key interface{}, iterKey interface{}) bool {
						id := key.(uint64)
						iterId := iterKey.(uint64)
						if id >= iterId {
							return true
						}
						return false
					}
					pool.validateEventQueue.RemoveWithCond(validationEvent.SeqNo, judge)
				}
			} else {
				break
			}
		}
	}
	return true
}

// dropValdiateEvent - this function do nothing but consume a validation event.
func (pool *BlockPool) dropValdiateEvent(validationEvent event.ExeTxsEvent) {
	atomic.StoreInt32(&pool.validateInProgress, PROGRESS_TRUE)
	atomic.StoreInt32(&pool.validateInProgress, PROGRESS_FALSE)
}

// Process an ValidationEvent
func (pool *BlockPool) PreProcess(validationEvent event.ExeTxsEvent) (error, bool) {
	var validTxSet []*types.Transaction
	var invalidTxSet []*types.InvalidTransactionRecord
	var index []int
	if validationEvent.IsPrimary {
		invalidTxSet, index = pool.checkSign(validationEvent.Transactions, pool.commonHash, pool.encryption)
	} else {
		validTxSet = validationEvent.Transactions
	}
	// remove invalid transaction from transaction list
	if len(index) > 0 {
		sort.Ints(index)
		count := 0
		set := validationEvent.Transactions
		for _, idx := range index {
			idx = idx - count
			set = append(set[:idx], set[idx+1:]...)
			count++
		}
		validTxSet = set
	} else {
		validTxSet = validationEvent.Transactions
	}
	// process block in virtual machine
	err, validateResult := pool.ProcessBlockInVm(validTxSet, invalidTxSet, validationEvent.SeqNo)
	if err != nil {
		log.Error("process block failed!, block number: ", validationEvent.SeqNo)
		return err, false
	}
	// calculate validation result hash for comparison with others
	hash := pool.commonHash.Hash([]interface{}{
		validateResult.MerkleRoot,
		validateResult.TxRoot,
		validateResult.ReceiptRoot,
	})
	// save validation result into cache
	// load them in commit phase
	// update some variant
	if len(validateResult.ValidTxs) != 0 {
		validateResult.VID = validationEvent.SeqNo
		validateResult.SeqNo = pool.tempBlockNumber
		// regard the batch as a valid block
		// increase tempBlockNumber for next validation
		// tempBlockNumber doesn't has concurrency dangerous
		pool.tempBlockNumber += 1
		pool.blockCache.Add(hash.Hex(), *validateResult)
	}
	log.Debug("invalid transaction number: ", len(validateResult.InvalidTxs))
	log.Debug("valid transaction number: ", len(validateResult.ValidTxs))
	// communicate with PBFT
	pool.consenter.RecvLocal(protos.ValidatedTxs{
		Transactions: validateResult.ValidTxs,
		SeqNo:        validationEvent.SeqNo,
		View:         validationEvent.View,
		Hash:         hash.Hex(),
		Timestamp:    validationEvent.Timestamp,
	})

	// empty block generated, throw all invalid transactions back to original node directly
	if validationEvent.IsPrimary && len(validateResult.ValidTxs) == 0 {
		// 1. Remove all cached transaction in this block (which for transaction duplication check purpose), because empty block won't enter network
		msg := protos.RemoveCache{Vid: validationEvent.SeqNo}
		pool.consenter.RecvLocal(msg)
		// 2. Throw all invalid transaction back to the origin node
		for _, t := range validateResult.InvalidTxs {
			payload, err := proto.Marshal(t)
			if err != nil {
				log.Error("Marshal tx error")
			}
			// original node is local
			// store invalid transaction directly
			if t.Tx.Id == uint64(pool.peerManager.GetNodeId()) {
				pool.StoreInvalidResp(event.RespInvalidTxsEvent{
					Payload: payload,
				})
				continue
			}
			var peers []uint64
			peers = append(peers, t.Tx.Id)
			// send back to original node
			pool.peerManager.SendMsgToPeers(payload, peers, recovery.Message_INVALIDRESP)
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
		go func(i int) {
			tx := txs[i]
			if !tx.ValidateSign(encryption, commonHash) {
				log.Notice("validation, found invalid signature, send from :", tx.Id)
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
}

// ProcessBlockInVm - Put all transactions into the virtual machine and execute
// Return the execution result, such as txs' merkle root, receipts' merkle root, accounts' merkle root and so on.
func (pool *BlockPool) ProcessBlockInVm(txs []*types.Transaction, invalidTxs []*types.InvalidTransactionRecord, seqNo uint64) (error, *BlockRecord) {
	var validtxs []*types.Transaction
	var receipts []*types.Receipt
	// initialize calculator
	// for calculate fingerprint of a batch of transactions and receipts
	if err := pool.initializeTransactionCalculator(); err != nil {
		log.Errorf("validate #%d initialize transaction calculator", pool.tempBlockNumber)
		return err, &BlockRecord{
			InvalidTxs: invalidTxs,
		}
	}
	if err := pool.initializeReceiptCalculator(); err != nil {
		log.Errorf("validate #%d initialize receipt calculator", pool.tempBlockNumber)
		return err, &BlockRecord{
			InvalidTxs: invalidTxs,
		}
	}
	// load latest state fingerprint
	// for compatibility, doesn't remove the statement below
	// initialize state
	state, err := pool.GetStateInstance()
	if err != nil {
		return err, &BlockRecord{
			InvalidTxs: invalidTxs,
		}
	}

	state.MarkProcessStart(pool.tempBlockNumber)
	// initialize execution environment rule set
	env := initEnvironment(state, pool.tempBlockNumber)
	// execute transaction one by one
	batch := state.FetchBatch(pool.tempBlockNumber)
	for i, tx := range txs {
		state.StartRecord(tx.GetTransactionHash(), common.Hash{}, i)
		receipt, _, _, err := core.ExecTransaction(tx, env)
		// invalid transaction, check invalid type
		if err != nil {
			var errType types.InvalidTransactionRecord_ErrType
			if core.IsValueTransferErr(err) {
				errType = types.InvalidTransactionRecord_OUTOFBALANCE
			} else if core.IsExecContractErr(err) {
				tmp := err.(*core.ExecContractError)
				if tmp.GetType() == 0 {
					errType = types.InvalidTransactionRecord_DEPLOY_CONTRACT_FAILED
				} else if tmp.GetType() == 1 {
					errType = types.InvalidTransactionRecord_INVOKE_CONTRACT_FAILED
				}
			}
			invalidTxs = append(invalidTxs, &types.InvalidTransactionRecord{
				Tx:      tx,
				ErrType: errType,
				ErrMsg:  []byte(err.Error()),
			})
			continue
		}
		pool.calculateTransactionsFingerprint(tx, false)
		pool.calculateReceiptFingerprint(receipt, false)
		receipts = append(receipts, receipt)
		validtxs = append(validtxs, tx)
	}
	// submit validation result
	err, merkleRoot, txRoot, receiptRoot := pool.submitValidationResult(state, batch)
	if err != nil {
		log.Error("Commit state db failed! error msg, ", err.Error())
		return err, &BlockRecord{
			InvalidTxs: invalidTxs,
		}
	}
	// generate new state fingerprint
	// IMPORTANT doesn't call batch.Write util recv commit event for atomic assurance
	log.Noticef("validate result temp block number #%d, vid #%d, merkle root [%s],  transaction root [%s],  receipt root [%s]",
		pool.tempBlockNumber, seqNo, common.Bytes2Hex(merkleRoot), common.Bytes2Hex(txRoot), common.Bytes2Hex(receiptRoot))
	return nil, &BlockRecord{
		TxRoot:      txRoot,
		ReceiptRoot: receiptRoot,
		MerkleRoot:  merkleRoot,
		Receipts:    receipts,
		ValidTxs:    validtxs,
		InvalidTxs:  invalidTxs,
	}
}


// initialize transaction execution environment
func initEnvironment(state vm.Database, seqNo uint64) vm.Environment {
	env := make(map[string]string)
	env["currentNumber"] = strconv.FormatUint(seqNo, 10)
	env["currentGasLimit"] = "200000000"
	vmenv := core.NewEnvFromMap(core.RuleSet{params.MainNetHomesteadBlock, params.MainNetDAOForkBlock, true}, state, env)
	return vmenv
}

// initialize transaction calculator
func (pool *BlockPool) initializeTransactionCalculator() error {
	db, err := hyperdb.GetDBDatabase()
	if err != nil {
		log.Error("get database handler failed in initializeTransactionCalculator")
		return err
	}
	switch pool.GetStateType() {
	case "rawstate":
		tree, err := pmt.New(common.Hash{}, db)
		if err != nil {
			log.Error("initialize pmt failed in initializeTransactionCalculator")
			return err
		}
		pool.transactionCalculator = tree
	case "hyperstate":
		pool.transactionBuffer = nil
	}
	return nil
}

// initialize receipt calculator
func (pool *BlockPool) initializeReceiptCalculator() error {
	db, err := hyperdb.GetDBDatabase()
	if err != nil {
		log.Error("get database handler failed in initializeTransactionCalculator")
		return err
	}
	switch pool.GetStateType() {
	case "rawstate":
		tree, err := pmt.New(common.Hash{}, db)
		if err != nil {
			log.Error("initialize pmt failed in initializeTransactionCalculator")
			return err
		}
		pool.receiptCalculator = tree
	case "hyperstate":
		pool.receiptBuffer = nil
	}
	return nil
}

// calculate a batch of transaction
func (pool *BlockPool) calculateTransactionsFingerprint(transaction *types.Transaction, flush bool) (common.Hash, error) {
	if transaction == nil && flush == false {
		return common.Hash{}, errors.New("empty pointer")
	}
	switch pool.GetStateType() {
	case "rawstate":
		calculator := pool.transactionCalculator.(*pmt.Trie)
		if flush == false {
			err, data := core.WrapperTransaction(transaction, pool.GetTransactionVersion())
			if err != nil {
				log.Error("Invalid Transaction struct to marshal! error msg, ", err.Error())
				return common.Hash{}, err
			}
			// put transaction to buffer temporarily
			calculator.Update(append(core.TransactionPrefix, transaction.GetTransactionHash().Bytes()...), data)
			return common.Hash{}, nil
		} else {
			// calculate hash together
			return calculator.Commit()
		}
	case "hyperstate":
		if flush == false {
			err, data := core.WrapperTransaction(transaction, pool.GetTransactionVersion())
			if err != nil {
				log.Error("Invalid Transaction struct to marshal! error msg, ", err.Error())
				return common.Hash{}, err
			}
			// put transaction to buffer temporarily
			pool.transactionBuffer = append(pool.transactionBuffer, data)
			return common.Hash{}, nil
		} else {
			// calculate hash together
			kec256Hash := crypto.NewKeccak256Hash("keccak256")
			hash := kec256Hash.Hash(pool.transactionBuffer)
			pool.transactionBuffer = nil
			return hash, nil
		}
	}
	return common.Hash{}, nil
}

// calculate a batch of receipt
func (pool *BlockPool) calculateReceiptFingerprint(receipt *types.Receipt, flush bool) (common.Hash, error) {
	// 1. marshal receipt to byte slice
	if receipt == nil && flush == false {
		log.Error("empty recepit pointer")
		return common.Hash{}, errors.New("empty pointer")
	}
	switch pool.GetStateType() {
	case "rawstate":
		calculator := pool.receiptCalculator.(*pmt.Trie)
		if flush == false {
			// process
			err, data := core.WrapperReceipt(receipt, pool.GetTransactionVersion())
			if err != nil {
				log.Error("Invalid receipt struct to marshal! error msg, ", err.Error())
				return common.Hash{}, err
			}

			// put transaction to buffer temporarily
			calculator.Update(append(core.ReceiptsPrefix, receipt.TxHash...), data)
			return common.Hash{}, nil
		} else {
			// calculate hash together
			return calculator.Commit()
		}
	case "hyperstate":
		if flush == false {
			// process
			err, data := core.WrapperReceipt(receipt, pool.GetTransactionVersion())
			if err != nil {
				log.Error("Invalid receipt struct to marshal! error msg, ", err.Error())
				return common.Hash{}, err
			}
			// put transaction to buffer temporarily
			pool.receiptBuffer = append(pool.receiptBuffer, data)
			return common.Hash{}, nil
		} else {
			// calculate hash together
			kec256Hash := crypto.NewKeccak256Hash("keccak256")
			hash := kec256Hash.Hash(pool.receiptBuffer)
			pool.receiptBuffer = nil
			return hash, nil
		}
	}
	return common.Hash{}, nil
}

func (pool *BlockPool) submitValidationResult(state vm.Database, batch hyperdb.Batch) (error, []byte, []byte, []byte) {
	switch pool.GetStateType() {
	case "hyperstate":
		// flush all state change
		root, err := state.Commit()
		state.Reset()
		if err != nil {
			log.Error("Commit state db failed! error msg, ", err.Error())
			return err, nil, nil, nil
		}
		// generate new state fingerprint
		merkleRoot := root.Bytes()
		// generate transactions and receipts fingerprint
		res, _ := pool.calculateTransactionsFingerprint(nil, true)
		txRoot := res.Bytes()
		res, _ = pool.calculateReceiptFingerprint(nil, true)
		receiptRoot := res.Bytes()
		// store latest state status
		// actually it's useless
		pool.lastValidationState.Store(root)
		return nil, merkleRoot, txRoot, receiptRoot
		// IMPORTANT doesn't call batch.Write util recv commit event for atomic assurance
	case "rawstate":
		// flush all state change
		root, err := state.Commit()
		if err != nil {
			log.Error("Commit state db failed! error msg, ", err.Error())
			return err, nil, nil, nil
		}
		// generate new state fingerprint
		merkleRoot := root.Bytes()
		// generate transactions and receipts fingerprint
		res, _ := pool.calculateTransactionsFingerprint(nil, true)
		txRoot := res.Bytes()
		res, _ = pool.calculateReceiptFingerprint(nil, true)
		receiptRoot := res.Bytes()
		// store latest state status
		pool.lastValidationState.Store(root)
		batch.Write()
		return nil, merkleRoot, txRoot, receiptRoot
	}
	return errors.New("miss state type"), nil, nil, nil
}

func (pool *BlockPool) validateEventCheck(ev event.ExeTxsEvent) bool {
	if ev.SeqNo != pool.demandSeqNo {
		log.Errorf("got a validation event is not required. required %d but got %d", pool.demandSeqNo, ev.SeqNo)
		return false
	}
	return true
}

func (pool *BlockPool) increaseDemandSeqNo() {
	pool.demandSeqNo += 1
	log.Noticef("demand seqNo %d", pool.demandSeqNo)
}

