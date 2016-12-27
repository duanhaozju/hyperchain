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
	"hyperchain/core/vm"
)

// Validate is an entry of `validate process`
// When a validationEvent received, put it into the validationQueue
// If the demand ValidationEvent arrived, call `PreProcess` function
// IMPORTANT this function called in parallelly, Make sure all the variable are thread-safe
func (pool *BlockPool) Validate(validationEvent event.ExeTxsEvent, commonHash crypto.CommonHash, encryption crypto.Encryption, peerManager p2p.PeerManager) {
	// check whether this is necessary to update max seqNo
	log.Noticef("begin to validate #%d", validationEvent.SeqNo)
	if validationEvent.SeqNo > atomic.LoadUint64(&pool.maxSeqNo) {
		atomic.StoreUint64(&pool.maxSeqNo, validationEvent.SeqNo)
	}
	// check validation event duplication
	if _, existed := pool.validationQueue.Get(validationEvent.SeqNo); existed {
		log.Error("receive repeat validation event, ", validationEvent.SeqNo)
		return
	}

	// check SeqNo
	if validationEvent.SeqNo < atomic.LoadUint64(&pool.demandSeqNo) {
		// receive invalid validation event
		log.Error("receive invalid validation event, seqno less than demandseqNo, ", validationEvent.SeqNo)
		return
	} else if validationEvent.SeqNo == atomic.LoadUint64(&pool.demandSeqNo) {
		// process
		if _, success := pool.PreProcess(validationEvent, commonHash, encryption, peerManager); success {
			atomic.AddUint64(&pool.demandSeqNo, 1)
			currentDemandSeqNo := atomic.LoadUint64(&pool.demandSeqNo)
			log.Noticef("current demandSeqNo is %d", currentDemandSeqNo)
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

		// process remain event
		for i := validationEvent.SeqNo + 1; i <= atomic.LoadUint64(&pool.maxSeqNo); i += 1 {
			if ret, existed := pool.validationQueue.Get(i); existed {
				ev := ret.(event.ExeTxsEvent)
				if _, success := pool.PreProcess(ev, commonHash, encryption, peerManager); success {
					pool.validationQueue.Remove(i)
					atomic.AddUint64(&pool.demandSeqNo, 1)
					currentDemandSeqNo := atomic.LoadUint64(&pool.demandSeqNo)
					log.Noticef("current demandSeqNo is %d", currentDemandSeqNo)
				} else {
					log.Error("pre process failed")
					return
				}
			} else {
				break
			}
		}
		return
	} else {
		log.Noticef("receive validation event  %d which is not demand, save into cache temporarily", validationEvent.SeqNo)
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
	// remove invalid transaction from transaction list
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
	// process block in virtual machine
	err, validateResult := pool.ProcessBlockInVm(validTxSet, invalidTxSet, validationEvent.SeqNo)
	if err != nil {
		log.Error("process block failed!, block number: ", validationEvent.SeqNo)
		return err, false
	}
	// calculate validation result hash for comparison with others
	hash := commonHash.Hash([]interface{}{
		validateResult.MerkleRoot,
		validateResult.TxRoot,
		validateResult.ReceiptRoot,
	})
	// save validation result into cache
	// load them in commit phase
	if len(validTxSet) != 0 {
		validateResult.SeqNo = validationEvent.SeqNo
		pool.blockCache.Add(hash.Hex(), *validateResult)
	}
	log.Info("invalid transaction number: ", len(invalidTxSet))
	log.Info("valid transaction number: ", len(validTxSet))
	// communicate with PBFT
	pool.consenter.RecvLocal(protos.ValidatedTxs{
		Transactions: validTxSet,
		SeqNo:        validationEvent.SeqNo,
		View:         validationEvent.View,
		Hash:         hash.Hex(),
		Timestamp:    validationEvent.Timestamp,
	})

	// empty block generated, throw all invalid transactions back to original node directly
	if validationEvent.IsPrimary && len(validTxSet) == 0 {
		// 1. Remove all cached transaction in this block (which for transaction duplication check purpose), because empty block won't enter network
		pool.consenter.RemoveCachedBatch(validationEvent.SeqNo)
		// 2. Throw all invalid transaction back to the origin node
		for _, t := range invalidTxSet {
			payload, err := proto.Marshal(t)
			if err != nil {
				log.Error("Marshal tx error")
			}
			// original node is local
			// store invalid transaction directly, instead of send by network
			if t.Tx.Id == uint64(peerManager.GetNodeId()) {
				pool.StoreInvalidResp(event.RespInvalidTxsEvent{
					Payload: payload,
				})
				continue
			}
			var peers []uint64
			peers = append(peers, t.Tx.Id)
			// send back to original node
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


// Put all transactions into the virtual machine and execute
// Return the execution result, such as txs' merkle root, receipts' merkle root, accounts' merkle root and so on
func (pool *BlockPool) ProcessBlockInVm(txs []*types.Transaction, invalidTxs []*types.InvalidTransactionRecord, seqNo uint64) (error, *BlockRecord) {
	var validtxs []*types.Transaction
	var receipts []*types.Receipt
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		return err, &BlockRecord{
			InvalidTxs: invalidTxs,
		}
	}
	// initailize calculator
	// for calculate fingerprint of a batch of transactions and receipts
	if err := pool.initializeTransactionCalculator(); err != nil {
		log.Errorf("validate #%d initialize transaction calculator", seqNo)
		return err, &BlockRecord{
			InvalidTxs: invalidTxs,
		}
	}
	if err := pool.initializeReceiptCalculator(); err != nil {
		log.Errorf("validate #%d initialize receipt calculator", seqNo)
		return err, &BlockRecord{
			InvalidTxs: invalidTxs,
		}
	}
	// load latest state fingerprint
	// for compatibility, doesn't remove the statement below
	v := pool.lastValidationState.Load()
	initStatus, ok := v.(common.Hash)
	if ok == false {
		return errors.New("get state status failed!"), &BlockRecord{
			InvalidTxs: invalidTxs,
		}
	}
	// initialize state
	state, err := pool.GetStateInstance(initStatus, db)
	if err != nil {
		return err, &BlockRecord{
			InvalidTxs: invalidTxs,
		}
	}

	state.MarkProcessStart(seqNo)
	// initialize execution environment ruleset
	vmenv := initEnvironment(state, seqNo)
	// execute transaction one by one
	batch := state.FetchBatch(seqNo)
	for i, tx := range txs {
		state.StartRecord(tx.GetTransactionHash(), common.Hash{}, i)
		receipt, _, _, err := core.ExecTransaction(*tx, vmenv)
		// invalid transaction, check invalid type
		if err != nil {
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
		// persist transaction
		var data []byte
		err, data = core.PersistTransaction(batch, tx, pool.conf.TransactionVersion, false, false);
		if err != nil {
			log.Error("put tx data into database failed! error msg, ", err.Error())
			return err, &BlockRecord{
				InvalidTxs: invalidTxs,
			}
		}
		pool.calculateTransactionsFingerprint(tx, data, false)
		// persist receipt
		err, data = core.PersistReceipt(batch, receipt, pool.conf.TransactionVersion, false, false)
		if  err != nil {
			log.Error("put receipt data into database failed! error msg, ", err.Error())
			return err, &BlockRecord{
				InvalidTxs: invalidTxs,
			}
		}
		pool.calculateReceiptFingerprint(receipt, data, false)
		receipts = append(receipts, receipt)
		// persist transaction meta
		// set temporarily
		// for primary node, the seqNo can be invalid. remove the incorrect txmeta info when commit block to avoid this error
		meta := &types.TransactionMeta{
			BlockIndex: seqNo,
			Index:      int64(i),
		}
		if err := core.PersistTransactionMeta(batch, meta, tx.GetTransactionHash(), false, false); err != nil {
			log.Error("Put txmeta into database failed! error msg, ", err.Error())
			return err, &BlockRecord{
				InvalidTxs: invalidTxs,
			}
		}
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
	log.Criticalf("validate result for #%d, merkle root [%s],  transaction root [%s],  receipt root [%s]",
		seqNo, common.Bytes2Hex(merkleRoot), common.Bytes2Hex(txRoot), common.Bytes2Hex(receiptRoot))
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
	env["currentGasLimit"] = "10000000"
	vmenv := core.NewEnvFromMap(core.RuleSet{params.MainNetHomesteadBlock, params.MainNetDAOForkBlock, true}, state, env)
	return vmenv
}
// initialize transaction calculator
func (pool *BlockPool) initializeTransactionCalculator() error {
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Error("get database handler failed in initializeTransactionCalculator")
		return err
	}
	switch pool.conf.StateType {
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
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Error("get database handler failed in initializeTransactionCalculator")
		return err
	}
	switch pool.conf.StateType {
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
func (pool *BlockPool) calculateTransactionsFingerprint(transaction *types.Transaction, data []byte, flush bool) (common.Hash, error) {
	switch pool.conf.StateType {
	case "rawstate":
		calculator := pool.transactionCalculator.(*pmt.Trie)
		if flush == false {
			// put transaction to buffer temporarily
			calculator.Update(append(core.TransactionPrefix, transaction.GetTransactionHash().Bytes()...), data)
			return common.Hash{}, nil
		} else {
			// calculate hash together
			return calculator.Commit()
		}
	case "hyperstate":
		if flush == false {
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
func (pool *BlockPool)calculateReceiptFingerprint(receipt *types.Receipt, data []byte, flush bool)(common.Hash, error) {
	switch pool.conf.StateType {
	case "rawstate":
		calculator := pool.receiptCalculator.(*pmt.Trie)
		if flush == false {
			// put transaction to buffer temporarily
			calculator.Update(append(core.ReceiptsPrefix, receipt.TxHash...), data)
			return common.Hash{}, nil
		} else {
			// calculate hash together
			return calculator.Commit()
		}
	case "hyperstate":
		if flush == false {
			// put transaction to buffer temporarily
			log.Infof("append receipt %s", receipt.ToReceiptTrans())
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
	switch  pool.conf.StateType {
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
		res, _ := pool.calculateTransactionsFingerprint(nil, nil, true)
		txRoot := res.Bytes()
		res, _ = pool.calculateReceiptFingerprint(nil, nil, true)
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
		res, _ := pool.calculateTransactionsFingerprint(nil, nil, true)
		txRoot := res.Bytes()
		res, _ = pool.calculateReceiptFingerprint(nil, nil, true)
		receiptRoot := res.Bytes()
		// store latest state status
		// actually it's useless
		pool.lastValidationState.Store(root)
		// flush content in batch to disk immediately
		batch.Write()
		return nil, merkleRoot, txRoot, receiptRoot
	}
	return errors.New("miss state type"), nil, nil, nil
}
