// implementblock pool
// author: Lizhong kuang
// date: 2016-08-29
// last modified:2016-09-01
package core

import (
	"hyperchain/event"
	"sync"

	"encoding/hex"
	"errors"
	"github.com/golang/protobuf/proto"

	//"fmt"
	"hyperchain/common"
	"hyperchain/consensus"
	"hyperchain/core/state"
	"hyperchain/core/types"
	"hyperchain/core/vm/params"
	"hyperchain/crypto"
	"hyperchain/hyperdb"
	"hyperchain/p2p"
	"hyperchain/recovery"
	"hyperchain/trie"
	"math/big"
	"strconv"
	"time"
)

const (
	maxQueued = 64 // max limit of queued block in pool
)

var (
	tempReceiptsMap map[uint64]types.Receipts
	public_batch    hyperdb.Batch
	batchsize       = 0
)

type BlockPool struct {
	demandNumber uint64
	demandSeqNo  uint64
	maxNum       uint64
	maxSeqNo     uint64

	queue           map[uint64]*types.Block
	validationQueue map[uint64]event.ExeTxsEvent

	consenter consensus.Consenter
	eventMux  *event.TypeMux
	events    event.Subscription
	mu        sync.RWMutex
	seqNoMu   sync.RWMutex
	stateLock sync.Mutex
	wg        sync.WaitGroup // for shutdown sync

	lastValidationState common.Hash
}

func (bp *BlockPool) SetDemandNumber(number uint64) {
	bp.demandNumber = number
}

func NewBlockPool(eventMux *event.TypeMux, consenter consensus.Consenter) *BlockPool {
	tempReceiptsMap = make(map[uint64]types.Receipts)

	pool := &BlockPool{
		eventMux:        eventMux,
		consenter:       consenter,
		queue:           make(map[uint64]*types.Block),
		validationQueue: make(map[uint64]event.ExeTxsEvent),
		events:          eventMux.Subscribe(event.NewBlockPoolEvent{}),
	}

	//pool.wg.Add(1)
	//go pool.eventLoop()

	currentChain := GetChainCopy()
	pool.demandNumber = currentChain.Height + 1
	pool.demandSeqNo = currentChain.Height + 1
	db, _ := hyperdb.GetLDBDatabase()
	blk, _ := GetBlock(db, currentChain.LatestBlockHash)
	pool.lastValidationState = common.BytesToHash(blk.MerkleRoot)
	return pool
}

// this method is used to Exec the transactions, if the err of one execution is not nil, we will
// abandon this transaction. And this method will return the new transactions and its' hash
func (pool *BlockPool) ExecTxs(sequenceNum uint64, transactions []types.Transaction) ([]types.Transaction, common.Hash, error) {
	var (
		receipts        types.Receipts
		env             = make(map[string]string)
		newTransactions []types.Transaction
	)

	// 1.prepare the current enviroment
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		return nil, common.Hash{}, err
	}
	currentBlock, _ := GetBlock(db, GetLatestBlockHash())
	statedb, err := state.New(common.BytesToHash(currentBlock.MerkleRoot), db)
	if err != nil {
		return nil, common.Hash{}, err
	}
	env["currentNumber"] = strconv.FormatUint(currentBlock.Number, 10)
	env["currentGasLimit"] = "10000000"
	vmenv := NewEnvFromMap(RuleSet{params.MainNetHomesteadBlock, params.MainNetDAOForkBlock, true}, statedb, env)

	// 2.exec all the transactions, if the err is nil, save the tx and append to newTransactions
	for i, tx := range transactions {
		statedb.StartRecord(tx.GetTransactionHash(), common.Hash{}, i)
		receipt, _, _, err := ExecTransaction(tx, vmenv)
		if err == nil {
			newTransactions = append(newTransactions, tx)
			receipts = append(receipts, receipt)
		}
	}

	// 3.save the receipts to the tempReceiptsMap
	tempReceiptsMap[sequenceNum] = receipts
	return newTransactions, crypto.NewKeccak256Hash("Keccak256").Hash(newTransactions), nil
}

func (pool *BlockPool) CommitOrRollbackBlockEvent(sequenceNum uint64, transactions []types.Transaction,
	timestamp int64, commitTime int64, CommitStatus bool) error {
	// 1.init a new block
	newBlock := new(types.Block)
	for _, tx := range transactions {
		newBlock.Transactions = append(newBlock.Transactions, &tx)
	}
	newBlock.Timestamp = timestamp
	newBlock.Number = sequenceNum
	// 2. add the block to the chain
	pool.AddBlockWithoutExecTxs(newBlock, crypto.NewKeccak256Hash("Keccak256"), commitTime)

	// 3.if CommitStatus is true,save the receipts to database
	//   or reset the statedb
	if CommitStatus {
		// save the receipts to database
		receiptInst, _ := GetReceiptInst()
		for _, receipt := range tempReceiptsMap[newBlock.Number] {
			receiptInst.PutReceipt(common.BytesToHash(receipt.TxHash), receipt)
		}
	} else {
		// prepare the current enviroment
		db, err := hyperdb.GetLDBDatabase()
		if err != nil {
			return err
		}
		currentBlock, _ := GetBlock(db, GetLatestBlockHash())
		statedb, err := state.New(common.BytesToHash(currentBlock.MerkleRoot), db)

		// reset the statedb
		statedb.Reset(common.BytesToHash(currentBlock.BlockHash))
	}
	// 4.delete the receipts of newBlock.Number
	delete(tempReceiptsMap, newBlock.Number)

	return nil
}

func (pool *BlockPool) eventLoop() {
	defer pool.wg.Done()

	for ev := range pool.events.Chan() {
		switch ev.Data.(type) {
		case event.NewBlockPoolEvent:
			pool.mu.Lock()
			/*if ev.Block != nil && pool.config.IsHomestead(ev.Block.Number()) {
				pool.homestead = true
			}

			pool.resetState()*/
			pool.mu.Unlock()

		}
	}
}

//check block sequence and validate in chain
func (pool *BlockPool) AddBlock(block *types.Block, commonHash crypto.CommonHash) {
	if block.Number == 0 {
		WriteBlock(block, commonHash)
		return
	}

	if block.Number > pool.maxNum {
		pool.maxNum = block.Number
	}
	if _, ok := pool.queue[block.Number]; ok {
		log.Info("replated block number,number is: ", block.Number)
		return
	}

	log.Info("number is ", block.Number)

	currentChain := GetChainCopy()

	if currentChain.Height >= block.Number {
		//todo view change ,delete block and rewrite block

		db, err := hyperdb.GetLDBDatabase()
		if err != nil {
			log.Fatal(err)
		}

		block, _ := GetBlockByNumber(db, block.Number)
		//rollback chain height,latestHash
		UpdateChainByViewChange(block.Number-1, block.ParentHash)
		keyNum := strconv.FormatInt(int64(block.Number), 10)
		DeleteBlock(db, append(blockNumPrefix, keyNum...))
		WriteBlock(block, commonHash)
		pool.demandNumber = GetChainCopy().Height + 1
		log.Notice("replated block number,number is: ", block.Number)
		return
	}
	if pool.demandNumber == block.Number {
		pool.mu.RLock()
		pool.demandNumber += 1
		log.Info("current demandNumber is ", pool.demandNumber)
		WriteBlock(block, commonHash)
		pool.mu.RUnlock()

		for i := block.Number + 1; i <= pool.maxNum; i += 1 {
			if _, ok := pool.queue[i]; ok { //存在}

				pool.mu.RLock()
				pool.demandNumber += 1
				log.Info("current demandNumber is ", pool.demandNumber)
				WriteBlock(pool.queue[i], commonHash)
				delete(pool.queue, i)
				pool.mu.RUnlock()

			} else {
				break
			}
		}
		return
	} else {
		pool.queue[block.Number] = block
	}

}

// WriteBlock need:
// 1. Put block into db
// 2. Put transactions in block into db  (-- cancel --)
// 3. Update chain
// 4. Update balance
func WriteBlock(block *types.Block, commonHash crypto.CommonHash) {
	log.Info("block number is ", block.Number)

	currentChain := GetChainCopy()

	block.ParentHash = currentChain.LatestBlockHash
	db, _ := hyperdb.GetLDBDatabase()
	//batch := db.NewBatch()
	/*
		if err := ProcessBlock(block, commonHash, commitTime); err != nil {
			log.Fatal(err)
		}
	*/
	block.WriteTime = time.Now().UnixNano()
	block.EvmTime = time.Now().UnixNano()

	block.BlockHash = block.Hash(commonHash).Bytes()

	UpdateChain(block, false)

	// update our stateObject and statedb to blockchain

	newChain := GetChainCopy()
	log.Notice("Block number", newChain.Height)
	log.Notice("Block hash", hex.EncodeToString(newChain.LatestBlockHash))

	/*	block.WriteTime = time.Now().UnixNano()
		block.CommitTime = commitTime
		block.BlockHash = block.Hash(commonHash).Bytes()
		data, err := proto.Marshal(block)
		if err != nil {
			log.Critical(err)
			return
		}


		keyFact := append(blockPrefix, block.BlockHash...)
		err = db.Put(keyFact,data)
	*/ /*if err := db.Put(keyFact, data); err != nil {
		return err
	}*/ /*
		keyNum := strconv.FormatInt(int64(block.Number), 10)
		//err = db.Put(append(blockNumPrefix, keyNum...), t.BlockHash)

		err = db.Put(append(blockNumPrefix, keyNum...),block.BlockHash)*/

	PutBlockTx(db, commonHash, block.BlockHash, block)
	if block.Number%10 == 0 && block.Number != 0 {
		WriteChainChan()
	}
	//TxSum.Add(TxSum,big.NewInt(int64(len(block.Transactions))))
	//CommitStatedbToBlockchain()
}

func ProcessBlock(block *types.Block, commonHash crypto.CommonHash, commitTime int64) error {
	var (
		//receipts types.Receipts
		env = make(map[string]string)
	)
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		return err
	}
	parentBlock, _ := GetBlock(db, block.ParentHash)
	statedb, e := state.New(common.BytesToHash(parentBlock.MerkleRoot), db)
	//fmt.Println("[Before Process %d] %s\n", block.Number, string(statedb.Dump()))
	if err != nil {
		return e
	}
	env["currentNumber"] = strconv.FormatUint(block.Number, 10)
	env["currentGasLimit"] = "10000000"
	vmenv := NewEnvFromMap(RuleSet{params.MainNetHomesteadBlock, params.MainNetDAOForkBlock, true}, statedb, env)

	//batch := db.NewBatch()
	public_batch = db.NewBatch()

	// todo run 20 ms in 500 tx
	// TODO just for sendContractTransactionTest
	var code = "0x6000805463ffffffff1916815560a0604052600b6060527f68656c6c6f20776f726c6400000000000000000000000000000000000000000060805260018054918190527f68656c6c6f20776f726c6400000000000000000000000000000000000000001681559060be907fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf66020600261010084871615026000190190931692909204601f01919091048101905b8082111560ce576000815560010160ac565b50506101b8806100d26000396000f35b509056606060405260e060020a60003504633ad14af3811461003c578063569c5f6d146100615780638da9b77214610071578063d09de08a146100da575b005b6000805460043563ffffffff8216016024350163ffffffff199190911617905561003a565b6100f960005463ffffffff165b90565b604080516020818101835260008252600180548451600261010083851615026000190190921691909104601f81018490048402820184019095528481526101139490928301828280156101ac5780601f10610181576101008083540402835291602001916101ac565b61003a6000805463ffffffff19811663ffffffff909116600101179055565b6040805163ffffffff929092168252519081900360200190f35b60405180806020018281038252838181518152602001915080519060200190808383829060006004602084601f0104600302600f01f150905090810190601f1680156101735780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b820191906000526020600020905b81548152906001019060200180831161018f57829003601f168201915b5050505050905061006e56"
	var tx_value1 = &types.TransactionValue{Price: 100000, GasLimit: 100000, Amount: 100, Payload: common.FromHex(code)}
	value, _ := proto.Marshal(tx_value1)
	//receipt, _, _, _ := ExecTransaction(*types.NewTestCreateTransaction(), vmenv)
	var addr common.Address
	for i, tx := range block.Transactions {
		//tx.To = addr.Bytes() //TODO just for test
		if i == 0 {
			tx.To = nil
			tx.Value = value
			_, _, addr, _ = ExecTransaction(*tx, vmenv)
		} else if i <= 125 {
			tx.To = addr.Bytes()
		} else {
			tx.Value = nil
		}

		statedb.StartRecord(tx.GetTransactionHash(), common.Hash{}, i)
		receipt, _, _, _ := ExecTransaction(*tx, vmenv)
		//ExecTransaction(*tx, vmenv)
		//receipts = append(receipts, receipt)

		txKey := tx.Hash(commonHash).Bytes()
		txKeyFact := append(transactionPrefix, txKey...)
		txValue, err := proto.Marshal(tx)
		if err != nil {
			return nil
		}
		data, err := proto.Marshal(receipt)

		if err := public_batch.Put(append(receiptsPrefix, receipt.TxHash...), data); err != nil {
			return err
		}

		public_batch.Put(txKeyFact, txValue)
		batchsize++
	}

	/*receiptInst, _ := GetReceiptInst()
	for _, receipt := range receipts {
		receiptInst.PutReceipt(common.BytesToHash(receipt.TxHash), receipt)
	}*/
	//WriteReceipts(receipts)

	begin := time.Now().UnixNano()
	root, _ := statedb.Commit()

	//batch := db.NewBatch()
	block.MerkleRoot = root.Bytes()

	/*block.WriteTime = time.Now().UnixNano()
	block.CommitTime = commitTime
	block.BlockHash = block.Hash(commonHash).Bytes()
	data, err := proto.Marshal(block)
	if err != nil {
		log.Critical(err)
		return err
	}

	keyFact := append(blockPrefix, block.BlockHash...)
	err = batch.Put(keyFact,data)
	*/ /*if err := db.Put(keyFact, data); err != nil {
		return err
	}*/ /*
		keyNum := strconv.FormatInt(int64(block.Number), 10)
		//err = db.Put(append(blockNumPrefix, keyNum...), t.BlockHash)

		err = batch.Put(append(blockNumPrefix, keyNum...),block.BlockHash)*/
	//batch.Write()
	end := time.Now().UnixNano()
	log.Notice("write time is ", (end-begin)/int64(time.Millisecond))
	//WriteBlockInDB(root,block,commitTime,commonHash)

	/*
		if(batchsize>=500){
			go public_batch.Write()
			batchsize = 0
			public_batch = db.NewBatch()
		}*/
	go public_batch.Write()

	//fmt.Println("[After Process %d] %s\n", block.Number, string(statedb.Dump()))
	return nil
}

func WriteBlockInDB(root common.Hash, block *types.Block, commitTime int64, commonHash crypto.CommonHash) {
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Critical(err)
		return
	}
	batch := db.NewBatch()
	block.MerkleRoot = root.Bytes()

	block.WriteTime = time.Now().UnixNano()
	block.CommitTime = commitTime
	block.BlockHash = block.Hash(commonHash).Bytes()
	data, err := proto.Marshal(block)
	if err != nil {
		log.Critical(err)
		return
	}

	keyFact := append(blockPrefix, block.BlockHash...)
	err = batch.Put(keyFact, data)
	/*if err := db.Put(keyFact, data); err != nil {
		return err
	}*/
	keyNum := strconv.FormatInt(int64(block.Number), 10)
	//err = db.Put(append(blockNumPrefix, keyNum...), t.BlockHash)

	err = batch.Put(append(blockNumPrefix, keyNum...), block.BlockHash)
	batch.Write()
}

func (pool *BlockPool) Validate(validationEvent event.ExeTxsEvent) {
	log.Debug("[Validate]begin")
	if validationEvent.SeqNo > pool.maxSeqNo {
		pool.maxSeqNo = validationEvent.SeqNo
	}
	// TODO Is necessary ?
	if _, ok := pool.validationQueue[validationEvent.SeqNo]; ok {
		log.Error("Receive Repeat ValidationEvent, ", validationEvent.SeqNo)
		return
	}
	// (1) Check SeqNo
	if validationEvent.SeqNo < pool.demandSeqNo {
		// receive repeat ValidationEvent
		log.Error("Receive Repeat ValidationEvent, ", validationEvent.SeqNo)
		return
	} else if validationEvent.SeqNo == pool.demandSeqNo {
		// Process
		pool.seqNoMu.RLock()
		if _, success := pool.PreProcess(validationEvent); success {
			pool.demandSeqNo += 1
			log.Notice("Current demandSeqNo is, ", pool.demandSeqNo)
		}
		pool.seqNoMu.RUnlock()
		// Process remain event
		for i := validationEvent.SeqNo + 1; i <= pool.maxSeqNo; i += 1 {
			if _, ok := pool.validationQueue[i]; ok {
				pool.seqNoMu.RLock()
				// Process
				if _, success := pool.PreProcess(pool.validationQueue[i]); success {
					pool.demandSeqNo += 1
					log.Notice("Current demandSeqNo is, ", pool.demandSeqNo)
				}
				delete(pool.validationQueue, i)
				pool.seqNoMu.RUnlock()
			} else {
				break
			}
		}
		return
	} else {
		log.Notice("Receive ValidationEvent which is not demand, ", validationEvent.SeqNo)
		pool.validationQueue[validationEvent.SeqNo] = validationEvent
	}
}

func (pool *BlockPool) PreProcess(validationEvent event.ExeTxsEvent) (error, bool) {
	var validTxSet []*types.Transaction
	var invalidTxSet []*types.InvalidTransactionRecord
	if validationEvent.IsPrimary {
		validTxSet, invalidTxSet = pool.PreCheck(validationEvent.Transactions)
	} else {
		validTxSet = validationEvent.Transactions
	}
	err, _, merkleRoot, txRoot, receiptRoot, validTxSet, invalidTxSet := pool.ProcessBlock1(validTxSet, invalidTxSet, validationEvent.SeqNo)
	if err != nil {
		return err, false
	}
	blockCache, _ := GetBlockCache()
	blockCache.Record(validationEvent.SeqNo, BlockRecord{
		TxRoot:      txRoot,
		ReceiptRoot: receiptRoot,
		MerkleRoot:  merkleRoot,
		InvalidTxs:  invalidTxSet,
		ValidTxs:    validTxSet,
	})
	log.Info("Invalid Tx number: ", len(invalidTxSet))
	log.Info("Valid Tx number: ", len(validTxSet))
	// Communicate with PBFT
	hash := crypto.NewKeccak256Hash("Keccak256").Hash([]interface{}{
		merkleRoot,
		txRoot,
		receiptRoot,
	})
	pool.consenter.RecvValidatedResult(event.ValidatedTxs{
		Transactions: validTxSet,
		//Digest:       validationEvent.Digest,
		SeqNo:        validationEvent.SeqNo,
		View:         validationEvent.View,
		Hash:         hash.Bytes(),
	})
	return nil, true
}
func (pool *BlockPool) PreCheck(txs []*types.Transaction) ([]*types.Transaction, []*types.InvalidTransactionRecord) {
	var validTxSet []*types.Transaction
	var invalidTxSet []*types.InvalidTransactionRecord
	encryption := crypto.NewEcdsaEncrypto("ecdsa")
	kec256Hash := crypto.NewKeccak256Hash("keccak256")
	// (1) check signature for each transaction
	for _, tx := range txs {
		if !tx.ValidateSign(encryption, kec256Hash) {
			log.Info("Validation, found invalid signature, send from :", tx.Id)
			invalidTxSet = append(invalidTxSet, &types.InvalidTransactionRecord{
				Tx:      tx,
				ErrType: types.InvalidTransactionRecord_SIGFAILED,
			})
		} else {
			validTxSet = append(validTxSet, tx)
		}
	}
	return validTxSet, invalidTxSet
}

func (pool *BlockPool) ProcessBlock1(txs []*types.Transaction, invalidTxs []*types.InvalidTransactionRecord, seqNo uint64) (error, []byte, []byte, []byte, []byte, []*types.Transaction, []*types.InvalidTransactionRecord) {
	log.Debug("[ProcessBlock1] txs: ", len(txs))
	var validtxs []*types.Transaction
	var (
		//receipts types.Receipts
		env = make(map[string]string)
	)
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		return err, nil, nil, nil, nil, nil, invalidTxs
	}
	txTrie, _ := trie.New(common.Hash{}, db)
	receiptTrie, _ := trie.New(common.Hash{}, db)
	statedb, err := state.New(pool.lastValidationState, db)
	if err != nil {
		return err, nil, nil, nil, nil, nil, invalidTxs
	}
	env["currentNumber"] = strconv.FormatUint(seqNo, 10)
	env["currentGasLimit"] = "10000000"
	vmenv := NewEnvFromMap(RuleSet{params.MainNetHomesteadBlock, params.MainNetDAOForkBlock, true}, statedb, env)

	public_batch = db.NewBatch()

	for i, tx := range txs {
		if !CanTransfer(common.BytesToAddress(tx.From), statedb, tx.Amount()) {
			invalidTxs = append(invalidTxs, &types.InvalidTransactionRecord{
				Tx:      tx,
				ErrType: types.InvalidTransactionRecord_OUTOFBALANCE,
			})
			continue
		}
		statedb.StartRecord(tx.GetTransactionHash(), common.Hash{}, i)
		receipt, _, _, _ := ExecTransaction(*tx, vmenv)

		// save to DB
		txValue, _ := proto.Marshal(tx)
		if err := public_batch.Put(append(transactionPrefix, tx.GetTransactionHash().Bytes()...), txValue); err != nil {
			return err, nil, nil, nil, nil, nil, invalidTxs
		}

		receiptValue, _ := proto.Marshal(receipt)
		if err := public_batch.Put(append(receiptsPrefix, receipt.TxHash...), receiptValue); err != nil {
			return err, nil, nil, nil, nil, nil, invalidTxs
		}

		// Update trie
		txTrie.Update(append(transactionPrefix, tx.GetTransactionHash().Bytes()...), txValue)
		receiptTrie.Update(append(receiptsPrefix, receipt.TxHash...), receiptValue)

		validtxs = append(validtxs, tx)
	}

	root, _ := statedb.Commit()

	merkleRoot := root.Bytes()
	txRoot := txTrie.Hash().Bytes()
	receiptRoot := receiptTrie.Hash().Bytes()

	pool.lastValidationState = root

	go public_batch.Write()

	return nil, nil, merkleRoot, txRoot, receiptRoot, validtxs, invalidTxs
}

func (pool *BlockPool) CommitBlock(ev event.CommitOrRollbackBlockEvent, peerManager p2p.PeerManager) {
	blockCache, _ := GetBlockCache()
	record := blockCache.Get(ev.SeqNo)
	if ev.Flag {
		// 1.init a new block
		newBlock := new(types.Block)
		for _, tx := range record.ValidTxs {
			newBlock.Transactions = append(newBlock.Transactions, tx)
		}
		newBlock.MerkleRoot = record.MerkleRoot
		newBlock.TxRoot = record.TxRoot
		newBlock.ReceiptRoot = record.ReceiptRoot
		newBlock.Timestamp = ev.Timestamp
		newBlock.CommitTime = ev.CommitTime
		newBlock.Number = ev.SeqNo
		// 2.save block and update chain
		pool.AddBlock(newBlock, crypto.NewKeccak256Hash("Keccak256"))
		// 3.throw invalid tx back to origin node if current peer is primary
		if ev.IsPrimary {
			for _, t := range record.InvalidTxs {
				payload, err := proto.Marshal(t)
				if err != nil {
					log.Error("Marshal tx error")
				}
				message := &recovery.Message{
					MessageType:  recovery.Message_INVALIDRESP,
					MsgTimeStamp: time.Now().UnixNano(),
					Payload:      payload,
				}
				broadcastMsg, err := proto.Marshal(message)
				if err != nil {
					log.Error("Marshal Message")
				}
				var peers []uint64
				peers = append(peers, t.Tx.Id)
				peerManager.SendMsgToPeers(broadcastMsg, peers, recovery.Message_INVALIDRESP)
			}
		}
	} else {
		db, _ := hyperdb.GetLDBDatabase()
		for _, t := range record.InvalidTxs {
			db.Delete(append(transactionPrefix, t.Tx.GetTransactionHash().Bytes()...))
			db.Delete(append(receiptsPrefix, t.Tx.GetTransactionHash().Bytes()...))
		}
		// TODO
	}
	blockCache.Delete(ev.SeqNo)
}

func (pool *BlockPool) StoreInvalidResp(ev event.RespInvalidTxsEvent) {
	msg := &recovery.Message{}
	invalidTx := &types.InvalidTransactionRecord{}
	err := proto.Unmarshal(ev.Payload, msg)
	err = proto.Unmarshal(msg.Payload, invalidTx)
	if err != nil {
		log.Error("Unmarshal Payload failed")
	}
	// save to db
	db, _ := hyperdb.GetLDBDatabase()
	db.Put(append(invalidTransactionPrefix, invalidTx.Tx.TransactionHash...), msg.Payload)
}

func BuildTree(prefix []byte, ctx []interface{}) ([]byte, error) {
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		return nil, err
	}
	trie, err := trie.New(common.Hash{}, db)
	if err != nil {
		return nil, err
	}
	for _, item := range ctx {
		switch t := item.(type) {
		case *types.Receipt:
			data, err := proto.Marshal(t)
			if err != nil {
				return nil, err
			}
			trie.Update(append(receiptsPrefix, t.TxHash...), data)
		case *types.Transaction:
			data, err := proto.Marshal(t)
			if err != nil {
				return nil, err
			}
			trie.Update(append(transactionPrefix, t.GetTransactionHash().Bytes()...), data)
		default:
			return nil, errors.New("Invalid element type when build tree")
		}
	}
	return trie.Hash().Bytes(), nil
}

func convertR(receipts types.Receipts) (ret []interface{}) {
	ret = make([]interface{}, len(receipts))
	for idx, v := range receipts {
		ret[idx] = v
	}
	return
}
func convertT(txs []*types.Transaction) (ret []interface{}) {
	ret = make([]interface{}, len(txs))
	for idx, v := range txs {
		ret[idx] = v
	}
	//fmt.Printf("[After Process %d] %s\n", block.Number, string(statedb.Dump()))
	return nil
}

//check block sequence and validate in chain
func (pool *BlockPool) AddBlockWithoutExecTxs(block *types.Block, commonHash crypto.CommonHash, commitTime int64) {

	if block.Number == 0 {
		writeBlockWithoutExecTx(block, commonHash)
		return
	}

	if block.Number > pool.maxNum {
		pool.maxNum = block.Number
	}
	if _, ok := pool.queue[block.Number]; ok {
		log.Info("replated block number,number is: ", block.Number)
		return
	}

	log.Info("number is ", block.Number)

	currentChain := GetChainCopy()

	if currentChain.Height >= block.Number {
		//todo view change ,delete block and rewrite block

		db, err := hyperdb.GetLDBDatabase()
		if err != nil {
			log.Fatal(err)
		}

		block, _ := GetBlockByNumber(db, block.Number)
		//rollback chain height,latestHash
		UpdateChainByViewChange(block.Number-1, block.ParentHash)
		keyNum := strconv.FormatInt(int64(block.Number), 10)
		DeleteBlock(db, append(blockNumPrefix, keyNum...))
		writeBlockWithoutExecTx(block, commonHash)
		log.Notice("replated block number,number is: ", block.Number)
		return
	}

	if pool.demandNumber == block.Number {

		pool.mu.RLock()
		pool.demandNumber += 1
		log.Info("current demandNumber is ", pool.demandNumber)

		writeBlockWithoutExecTx(block, commonHash)

		pool.mu.RUnlock()

		for i := block.Number + 1; i <= pool.maxNum; i += 1 {
			if _, ok := pool.queue[i]; ok { //存在}

				pool.mu.RLock()
				pool.demandNumber += 1
				log.Info("current demandNumber is ", pool.demandNumber)
				writeBlockWithoutExecTx(pool.queue[i], commonHash)
				delete(pool.queue, i)
				pool.mu.RUnlock()

			} else {
				break
			}

		}
		return
	} else {
		pool.queue[block.Number] = block
	}
}

// write the block to db but don't exec the txs
func writeBlockWithoutExecTx(block *types.Block, commonHash crypto.CommonHash) {
	log.Info("block number is ", block.Number)
	currentChain := GetChainCopy()
	block.ParentHash = currentChain.LatestBlockHash
	block.WriteTime = time.Now().UnixNano()
	block.BlockHash = block.Hash(commonHash).Bytes()
	block.EvmTime = time.Now().UnixNano()
	UpdateChain(block, false)

	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	if err := PutBlock(db, block.BlockHash, block); err != nil {
		log.Fatal(err)
	}

	newChain := GetChainCopy()
	log.Notice("Block number", newChain.Height)
	log.Notice("Block hash", hex.EncodeToString(newChain.LatestBlockHash))

	if block.Number%10 == 0 && block.Number != 0 {
		WriteChainChan()
	}
}

func CanTransfer(from common.Address, statedb *state.StateDB, value *big.Int) bool {
	return statedb.GetBalance(from).Cmp(value) >= 0
}
