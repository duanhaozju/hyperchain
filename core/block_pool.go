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
	"hyperchain/core/state"
	"hyperchain/core/types"
	"hyperchain/core/vm/params"
	"hyperchain/crypto"
	"hyperchain/hyperdb"
	"hyperchain/trie"
	"strconv"
	"time"


)

const (
	maxQueued = 64 // max limit of queued block in pool
)

var (
	tempReceiptsMap map[uint64]types.Receipts
	public_batch hyperdb.Batch
	batchsize = 0

)

type BlockPool struct {
	demandNumber uint64
	maxNum       uint64

	queue     map[uint64]*types.Block
	eventMux  *event.TypeMux
	events    event.Subscription
	mu        sync.RWMutex
	stateLock sync.Mutex
	wg        sync.WaitGroup // for shutdown sync
}

func (bp *BlockPool) SetDemandNumber(number uint64) {
	bp.demandNumber = number
}

func NewBlockPool(eventMux *event.TypeMux) *BlockPool {
	tempReceiptsMap = make(map[uint64]types.Receipts)

	pool := &BlockPool{
		eventMux: eventMux,

		queue: make(map[uint64]*types.Block),

		events: eventMux.Subscribe(event.NewBlockPoolEvent{}),
	}

	//pool.wg.Add(1)
	//go pool.eventLoop()

	currentChain := GetChainCopy()
	pool.demandNumber = currentChain.Height + 1
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
func (pool *BlockPool) AddBlock(block *types.Block, commonHash crypto.CommonHash, commitTime int64) {
	if block.Number == 0 {
		WriteBlock(block, commonHash, commitTime)
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
		WriteBlock(block, commonHash, commitTime)
		pool.demandNumber = GetChainCopy().Height + 1
		log.Notice("replated block number,number is: ", block.Number)
		return
	}

	if pool.demandNumber == block.Number {

		pool.mu.RLock()
		pool.demandNumber += 1
		log.Info("current demandNumber is ", pool.demandNumber)

		WriteBlock(block, commonHash, commitTime)

		pool.mu.RUnlock()

		for i := block.Number + 1; i <= pool.maxNum; i += 1 {
			if _, ok := pool.queue[i]; ok { //存在}

				pool.mu.RLock()
				pool.demandNumber += 1
				log.Info("current demandNumber is ", pool.demandNumber)
				WriteBlock(pool.queue[i], commonHash, commitTime)
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
func WriteBlock(block *types.Block, commonHash crypto.CommonHash, commitTime int64) {
	log.Info("block number is ", block.Number)

	currentChain := GetChainCopy()

	block.ParentHash = currentChain.LatestBlockHash
	db, _ := hyperdb.GetLDBDatabase()
	//batch := db.NewBatch()
	if err := ProcessBlock(block,commonHash,commitTime ); err != nil {
		log.Fatal(err)
	}
	block.WriteTime = time.Now().UnixNano()
	block.CommitTime = commitTime
	block.BlockHash = block.Hash(commonHash).Bytes()

	UpdateChain(block, false)

	// update our stateObject and statedb to blockchain
	block.EvmTime = time.Now().UnixNano()



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
	*//*if err := db.Put(keyFact, data); err != nil {
		return err
	}*//*
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

func ProcessBlock(block *types.Block,commonHash crypto.CommonHash,commitTime int64) error {
	var (
		//receipts types.Receipts
		env      = make(map[string]string)
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
	begin_exectx:=time.Now().UnixNano()
	for i, tx := range block.Transactions {
		//tx.To = addr.Bytes() //TODO just for test
		if(i==0){
			tx.To = nil
			tx.Value = value
			_, _, addr, _ = ExecTransaction(*tx, vmenv)
			continue
		}else if(i<100) {
			tx.To = addr.Bytes()
		}else {
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
	end_exectx:=time.Now().UnixNano()
	log.Notice("exec tx time is ",(end_exectx-begin_exectx)/int64(time.Millisecond))
	/*receiptInst, _ := GetReceiptInst()
	for _, receipt := range receipts {
		receiptInst.PutReceipt(common.BytesToHash(receipt.TxHash), receipt)
	}*/
	//WriteReceipts(receipts)

	begin:=time.Now().UnixNano()
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
	*//*if err := db.Put(keyFact, data); err != nil {
		return err
	}*//*
	keyNum := strconv.FormatInt(int64(block.Number), 10)
	//err = db.Put(append(blockNumPrefix, keyNum...), t.BlockHash)

	err = batch.Put(append(blockNumPrefix, keyNum...),block.BlockHash)*/
	//batch.Write()
	end:=time.Now().UnixNano()
	log.Notice("write time is ",(end-begin)/int64(time.Millisecond))
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

func WriteBlockInDB(root common.Hash,block *types.Block,commitTime int64,commonHash crypto.CommonHash)  {
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
	err = batch.Put(keyFact,data)
	/*if err := db.Put(keyFact, data); err != nil {
		return err
	}*/
	keyNum := strconv.FormatInt(int64(block.Number), 10)
	//err = db.Put(append(blockNumPrefix, keyNum...), t.BlockHash)

	err = batch.Put(append(blockNumPrefix, keyNum...),block.BlockHash)
	batch.Write()
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
		WriteBlockWithoutExecTx(block, commonHash, commitTime)
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
		WriteBlockWithoutExecTx(block, commonHash, commitTime)
		log.Notice("replated block number,number is: ", block.Number)
		return
	}

	if pool.demandNumber == block.Number {

		pool.mu.RLock()
		pool.demandNumber += 1
		log.Info("current demandNumber is ", pool.demandNumber)

		WriteBlockWithoutExecTx(block, commonHash, commitTime)

		pool.mu.RUnlock()

		for i := block.Number + 1; i <= pool.maxNum; i += 1 {
			if _, ok := pool.queue[i]; ok { //存在}

				pool.mu.RLock()
				pool.demandNumber += 1
				log.Info("current demandNumber is ", pool.demandNumber)
				WriteBlockWithoutExecTx(pool.queue[i], commonHash, commitTime)
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
func WriteBlockWithoutExecTx(block *types.Block, commonHash crypto.CommonHash, commitTime int64) {
	log.Info("block number is ", block.Number)
	currentChain := GetChainCopy()
	block.ParentHash = currentChain.LatestBlockHash
	block.WriteTime = time.Now().UnixNano()
	block.CommitTime = commitTime
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
