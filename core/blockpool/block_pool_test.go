// test blockpool
// author: Yijun Xie
// date: 2016-11-2
//last modify:2016-11-11
package blockpool

import (
	//"hyperchain/consensus"
	"github.com/golang/protobuf/proto"
	"hyperchain/consensus/controller"
	"hyperchain/core"
	"hyperchain/core/state"
	"hyperchain/core/types"
	"hyperchain/crypto"
	"hyperchain/event"
	"hyperchain/hyperdb"
	"hyperchain/membersrvc"
	"hyperchain/p2p"
	"testing"
	"time"
)

var transactionCases = []*types.Transaction{
	&types.Transaction{
		From:      []byte("000f1a7a08ccc48e5d30f80850cf1cf283aa3abd"),
		To:        []byte("e93b92f1da08f925bdee44e91e7768380ae83307"),
		Value:     []byte("10"),
		Id:        2,
		Timestamp: time.Now().UnixNano() - int64(time.Second),
		Signature: []byte{220, 7, 96, 100, 52, 63, 100, 74, 105, 33, 100, 42, 108, 113, 16, 68, 141, 165, 46, 231, 162, 109, 190, 100, 74, 158, 214, 209, 22, 243, 15, 95, 4, 31, 201, 185, 173, 246, 44, 168, 108, 104, 158, 1, 68, 160, 125, 167, 162, 238, 165, 206, 32, 133, 13, 106, 184, 145, 72, 156, 205, 80, 123, 105, 0},
	},
	&types.Transaction{
		From:      []byte("000f1a7a08ccc48e5d30f80850cf1cf283aa3abd"),
		To:        []byte("6201cb0448964ac597faf6fdf1f472edf2a22b89"),
		Value:     []byte("5"),
		Id:        1,
		Timestamp: time.Now().UnixNano(),
		Signature: []byte{201, 159, 75, 78, 48, 122, 73, 143, 135, 118, 176, 13, 79, 73, 69, 224, 139, 169, 178, 159, 188, 1, 195, 37, 203, 87, 232, 78, 58, 219, 31, 143, 35, 251, 186, 159, 249, 101, 238, 142, 117, 219, 167, 249, 245, 160, 8, 70, 174, 124, 108, 91, 217, 131, 15, 242, 62, 218, 114, 8, 191, 230, 76, 120, 0},
	},
}

var transactionCases1 = []*types.Transaction{
	&types.Transaction{
		From:      []byte("0000000000000000000000000000000000000001"),
		To:        []byte("0000000000000000000000000000000000000003"),
		Value:     []byte("100"),
		Timestamp: time.Now().UnixNano() - int64(time.Second),
		Signature: []byte("signature1"),
	},
	&types.Transaction{
		From:  []byte("0000000000000000000000000000000000000001"),
		To:    []byte("0000000000000000000000000000000000000002"),
		Value: []byte("100"), Timestamp: time.Now().UnixNano(),
		Signature: []byte("signature2"),
	},
	&types.Transaction{
		From:      []byte("0000000000000000000000000000000000000002"),
		To:        []byte("0000000000000000000000000000000000000003"),
		Value:     []byte("700"),
		Timestamp: time.Now().UnixNano(),
		Signature: []byte("signature3"),
	},
}

func TestSet(t *testing.T) {
	//	core.InitDB("G:/hyperchainDB", 8030)
	//	eventMux := new(event.TypeMux)
	//	consenter := new(consensus.Consenter)
	//	pool := NewBlockPool(eventMux, *consenter)
	pooltest := new(BlockPool)
	pooltest.SetDemandNumber(1)
	pooltest.SetDemandSeqNo(1)
	log.Notice("pool's demandnumber is", pooltest.demandNumber, "pool's demandseqo is", pooltest.demandSeqNo)
}

var pool *BlockPool
var encryption crypto.Encryption
var kec256Hash crypto.CommonHash
var grpcPeerMgr p2p.PeerManager

func init() {
	core.InitDB("G:/hyperchainDB", 8030)

	eventMux := new(event.TypeMux)
	consenter := controller.NewConsenter(uint64(1), eventMux, "../../config/pbft.yaml")
	//consenter := new(consensus.Consenter)
	pool = NewBlockPool(eventMux, consenter)
	pool.maxNum = 10
	log.Notice("poo's init demandnumber is", pool.demandNumber)

	encryption = crypto.NewEcdsaEncrypto("ecdsa")
	kec256Hash = crypto.NewKeccak256Hash("keccak256")
	membersrvc.Start("../../config/test/local_membersrvc.yaml", 1)

	//up 4 peers
	path := "../../config/local_peerconfig.json"
	grpcPeerMgr = p2p.NewGrpcManager(path, 1)
	aliveChan := make(chan bool)
	go grpcPeerMgr.Start(aliveChan, eventMux)

	grpcPeerMgr2 := p2p.NewGrpcManager(path, 2)

	go grpcPeerMgr2.Start(aliveChan, eventMux)

	grpcPeerMgr3 := p2p.NewGrpcManager(path, 3)

	go grpcPeerMgr3.Start(aliveChan, eventMux)

	grpcPeerMgr4 := p2p.NewGrpcManager(path, 4)

	go grpcPeerMgr4.Start(aliveChan, eventMux)

	nodeCount := 0
	for flag := range aliveChan {
		if flag {
			log.Info("A peer has connected")
			nodeCount += 1
		}
		if nodeCount >= 4 {
			break
		}
	}
}

//func TestNewBlockPool(t *testing.T) {
//	log.Notice("pool's demandnumber is ", pool.demandNumber)
//}

func TestValidate(t *testing.T) {
	//test validationEvent.SeqNo > pool.maxSeqNo
	validationEvent1 := new(event.ExeTxsEvent)
	validationEvent1.SeqNo = 15
	log.Notice("validationevent is", validationEvent1)
	pool.Validate(*validationEvent1, kec256Hash, encryption, grpcPeerMgr)

	//test event already exisels
	validationEvent2 := new(event.ExeTxsEvent)
	validationEvent2.SeqNo = 5
	pool.validationQueue.Add(validationEvent2.SeqNo, *validationEvent2)
	log.Notice("pool's validationqueue is", pool.validationQueue)
	pool.Validate(*validationEvent2, kec256Hash, encryption, grpcPeerMgr)
	pool.validationQueue.Remove(5)
	log.Notice("pool's validationqueue is", pool.validationQueue)

	//test validationEvent.SeqNo < pool.demandSeqNo
	validationEvent3 := new(event.ExeTxsEvent)
	validationEvent3.SeqNo = 1
	pool.Validate(*validationEvent3, kec256Hash, encryption, grpcPeerMgr)

	//test validationEvent.SeqNo > pool.demandSeqNo
	validationEvent4 := new(event.ExeTxsEvent)
	validationEvent4.SeqNo = pool.demandSeqNo + 1
	pool.Validate(*validationEvent4, kec256Hash, encryption, grpcPeerMgr)
	log.Notice("pool validationqueue is", pool.validationQueue)

	//test validationEvent.SeqNo == pool.demandSeqNo
	validationEvent := new(event.ExeTxsEvent)
	validationEvent.IsPrimary = true
	validationEvent.SeqNo = pool.demandSeqNo
	validationEvent.Transactions = transactionCases
	validationEvent.Timestamp = time.Now().UnixNano()
	pool.Validate(*validationEvent, kec256Hash, encryption, grpcPeerMgr)
	log.Notice("pool's demandseqo is", pool.demandSeqNo)
}

func TestPreProcess_one(t *testing.T) {
	//test isprimary
	validationEvent := new(event.ExeTxsEvent)
	validationEvent.IsPrimary = true
	validationEvent.SeqNo = 2
	validationEvent.Transactions = transactionCases
	validationEvent.Timestamp = time.Now().UnixNano()
	pool.PreProcess(*validationEvent, kec256Hash, encryption, grpcPeerMgr)

}

func TestPreProcess_two(t *testing.T) {
	//test is not primary
	validationEvent1 := new(event.ExeTxsEvent)
	validationEvent1.IsPrimary = false
	validationEvent1.SeqNo = 2
	validationEvent1.Transactions = transactionCases1
	validationEvent1.Timestamp = time.Now().UnixNano()
	pool.PreProcess(*validationEvent1, kec256Hash, encryption, grpcPeerMgr)
}

func TestProcessBlockInVm(t *testing.T) {
	//core.InitDB("G:/hyperchainDB", 8030)
	//eventMux := new(event.TypeMux)
	//consenter := new(consensus.Consenter)
	//pool = NewBlockPool(eventMux, *consenter)
	var invalidTxSet []*types.InvalidTransactionRecord
	err, _, merkleRoot, txRoot, receiptRoot, validtxs, invalidTxs := pool.ProcessBlockInVm(transactionCases, invalidTxSet, 1)
	if err != nil {
		log.Notice("Put txmeta into database failed!")
	}
	log.Notice("merkleRoot is-->", merkleRoot, "txRoot is-->", txRoot, "receiptRoot is-->", receiptRoot, "validtxs is--->", validtxs, "invalidtxs is-->", invalidTxs)

}

func TestCommitBlock(t *testing.T) {
	var invalidTxSet []*types.InvalidTransactionRecord
	invalidTxSet = []*types.InvalidTransactionRecord{
		&types.InvalidTransactionRecord{
			Tx:      transactionCases1[0],
			ErrType: types.InvalidTransactionRecord_OUTOFBALANCE,
			ErrMsg:  []byte("out of balance"),
		},
	}
	//flag=true,isprimary=true
	ev := new(event.CommitOrRollbackBlockEvent)
	ev.Flag = true
	ev.IsPrimary = true
	ev.SeqNo = pool.demandNumber
	hash := string([]byte{1, 2, 3, 4})
	ev.Hash = hash

	blockrecord := new(BlockRecord)
	blockrecord.InvalidTxs = invalidTxSet
	blockrecord.ValidTxs = transactionCases
	blockrecord.MerkleRoot = []byte{240, 53, 154, 92, 81, 30, 204, 173, 81, 8, 32, 30, 78, 239, 68, 145, 65, 201, 28, 93, 30, 243, 166, 130, 147, 197, 106, 124, 239, 110, 182, 169}
	blockrecord.ReceiptRoot = []byte{191, 216, 46, 62, 236, 4, 189, 185, 85, 225, 217, 180, 159, 235, 128, 197, 92, 218, 32, 85, 107, 228, 23, 248, 138, 236, 41, 174, 162, 208, 176, 49}
	blockrecord.TxRoot = []byte{135, 189, 209, 233, 54, 241, 174, 81, 206, 197, 43, 197, 240, 130, 139, 42, 127, 66, 148, 177, 248, 60, 95, 71, 83, 41, 150, 103, 65, 191, 50, 37}
	blockrecord.SeqNo = pool.demandNumber
	pool.blockCache.Add(ev.Hash, *blockrecord)
	log.Notice("pool's blockcache is ", pool.blockCache)
	pool.CommitBlock(*ev, kec256Hash, grpcPeerMgr)
	log.Notice("pool's blockcache is ", pool.blockCache)

	//flag=true,isprimary=false
	ev1 := new(event.CommitOrRollbackBlockEvent)
	ev1.Flag = true
	ev1.IsPrimary = false
	ev1.SeqNo = pool.demandNumber
	hash1 := string([]byte{1, 4, 3, 9})
	ev1.Hash = hash1
	pool.blockCache.Add(ev1.Hash, *blockrecord)
	pool.CommitBlock(*ev1, kec256Hash, grpcPeerMgr)
	log.Notice("pool's blockcache is ", pool.blockCache)
	log.Notice("pool's demandnumber is", pool.demandNumber)
}

func TestWriteBlock(t *testing.T) {
	newBlock := new(types.Block)
	newBlock.Transactions = make([]*types.Transaction, len(transactionCases))
	copy(newBlock.Transactions, transactionCases)
	currentChain := core.GetChainCopy()
	newBlock.ParentHash = currentChain.LatestBlockHash
	newBlock.MerkleRoot = []byte{240, 53, 154, 92, 81, 30, 204, 173, 81, 8, 32, 30, 78, 239, 68, 145, 65, 201, 28, 93, 30, 243, 166, 130, 147, 197, 106, 124, 239, 110, 182, 169}
	newBlock.TxRoot = []byte{135, 189, 209, 233, 54, 241, 174, 81, 206, 197, 43, 197, 240, 130, 139, 42, 127, 66, 148, 177, 248, 60, 95, 71, 83, 41, 150, 103, 65, 191, 50, 37}
	newBlock.ReceiptRoot = []byte{191, 216, 46, 62, 236, 4, 189, 185, 85, 225, 217, 180, 159, 235, 128, 197, 92, 218, 32, 85, 107, 228, 23, 248, 138, 236, 41, 174, 162, 208, 176, 49}
	//pool.lastValidationState=[32]byte{240, 53, 154, 92, 81, 30, 204, 173, 81, 8, 32, 30, 78, 239, 68, 145, 65, 201, 28, 93, 30, 243, 166, 130, 147, 197, 106, 124, 239, 110, 182, 169}

	newBlock.Number = pool.demandNumber + 1
	newBlock.WriteTime = time.Now().UnixNano()
	newBlock.EvmTime = time.Now().UnixNano()
	newBlock.BlockHash = newBlock.Hash(kec256Hash).Bytes()
	WriteBlock(newBlock, kec256Hash, newBlock.Number-1, true)
}

func TestStoreInvalidResp(t *testing.T) {
	ev := new(event.RespInvalidTxsEvent)
	invalidTx := &types.InvalidTransactionRecord{}
	invalidTx.Tx = transactionCases1[0]
	ev.Payload, _ = proto.Marshal(invalidTx)
	pool.StoreInvalidResp(*ev)
}

func TestResetStatus(t *testing.T) {
	var invalidTxSet []*types.InvalidTransactionRecord
	invalidTxSet = []*types.InvalidTransactionRecord{
		&types.InvalidTransactionRecord{
			Tx:      transactionCases1[0],
			ErrType: types.InvalidTransactionRecord_OUTOFBALANCE,
			ErrMsg:  []byte("out of balance"),
		},
	}
	ev := new(event.CommitOrRollbackBlockEvent)
	ev.Flag = true
	ev.IsPrimary = true
	ev.SeqNo = pool.demandNumber - 1
	hash := string([]byte{5, 6, 7, 8})
	ev.Hash = hash
	blockrecord := new(BlockRecord)
	blockrecord.InvalidTxs = invalidTxSet
	blockrecord.ValidTxs = transactionCases
	blockrecord.MerkleRoot = []byte{240, 53, 154, 92, 81, 30, 204, 173, 81, 8, 32, 30, 78, 239, 68, 145, 65, 201, 28, 93, 30, 243, 166, 130, 147, 197, 106, 124, 239, 110, 182, 169}
	blockrecord.ReceiptRoot = []byte{191, 216, 46, 62, 236, 4, 189, 185, 85, 225, 217, 180, 159, 235, 128, 197, 92, 218, 32, 85, 107, 228, 23, 248, 138, 236, 41, 174, 162, 208, 176, 49}
	blockrecord.TxRoot = []byte{135, 189, 209, 233, 54, 241, 174, 81, 206, 197, 43, 197, 240, 130, 139, 42, 127, 66, 148, 177, 248, 60, 95, 71, 83, 41, 150, 103, 65, 191, 50, 37}
	blockrecord.SeqNo = pool.demandNumber - 1
	pool.blockCache.Add(ev.Hash, *blockrecord)
	ev1 := new(event.VCResetEvent)
	ev1.SeqNo = pool.demandNumber - 2
	pool.ResetStatus(*ev1)
	log.Notice("pool 's lastvalidationstate is", pool.lastValidationState)

	db, err := hyperdb.GetDBDatabase()
	statedb, err := state.New(pool.lastValidationState, db)

	if err != nil {
		log.Notice("statedb is", statedb)
	}
}
