// test blockpool
// author: Yijun Xie
// date: 2016-11-2
package blockpool

import (
	"fmt"
	//"hyperchain/common"
	"hyperchain/consensus"
	"hyperchain/core"
	"hyperchain/core/types"
	"hyperchain/crypto"
	"hyperchain/event"
	"hyperchain/p2p"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
)

var transactionCases = []*types.Transaction{
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

func poolInit() *BlockPool {
	core.InitDB("G:/hyperchainDB", 8030)
	eventMux := new(event.TypeMux)
	consenter := new(consensus.Consenter)
	pool := NewBlockPool(eventMux, *consenter)
	return pool
}

func poolParameters() (crypto.CommonHash, crypto.Encryption, p2p.PeerManager) {
	encryption := crypto.NewEcdsaEncrypto("ecdsa")
	kec256Hash := crypto.NewKeccak256Hash("keccak256")
	path := "../../config/local_peerconfig.json"
	grpcPeerMgr := p2p.NewGrpcManager(path, 1)
	return kec256Hash, encryption, grpcPeerMgr
}

func TestSet(t *testing.T) {
	pool := poolInit()
	pool.SetDemandNumber(2)
	pool.SetDemandSeqNo(2)
	log.Notice("pool's demandnumber is ", pool.demandNumber, "pool's demandseqno is ", pool.demandSeqNo)
}

//preprocess have't test
func TestValidate(t *testing.T) {
	pool := poolInit()

	pool.maxSeqNo = 10
	validationEvent := new(event.ExeTxsEvent)
	validationEvent.SeqNo = 15
	fmt.Println("validationevent is", validationEvent)

	kec256Hash, encryption, grpcPeerMgr := poolParameters()
	//test blockpool.maxseqno<validationEvent.seqNo branch
	pool.Validate(*validationEvent, kec256Hash, encryption, grpcPeerMgr)

	//test event already exisels
	validationEvent.SeqNo = 5
	pool.validationQueue.Add(validationEvent.SeqNo, validationEvent)
	pool.Validate(*validationEvent, kec256Hash, encryption, grpcPeerMgr)

	//check seqNo
	//test validationEvent.SeqNo < pool.demandSeqNo
	validationEvent.SeqNo = 0
	pool.Validate(*validationEvent, kec256Hash, encryption, grpcPeerMgr)

	//test validationEvent.SeqNo > pool.demandSeqNo
	validationEvent.SeqNo = 2
	pool.Validate(*validationEvent, kec256Hash, encryption, grpcPeerMgr)
	log.Notice("pool validationqueue is", pool.validationQueue)

	//	//test validationEvent.SeqNo == pool.demandSeqNo
	//	//	validationEvent.SeqNo = 1
	//	//	pool.Validate(*validationEvent, kec256Hash, encryption, grpcPeerMgr)
}

//func TestPreProcess(t *testing.T) {
//	pool := poolInit()
//	kec256Hash, encryption, _ := poolParameters()
//	validationEvent := new(event.ExeTxsEvent)
//	validationEvent.IsPrimary = false
//	validationEvent.SeqNo = 1
//	validationEvent.Transactions = transactionCases
//	err, flag := pool.PreProcess(*validationEvent, kec256Hash, encryption, nil)
//	if err != nil {
//		fmt.Errorf("process failed")
//	}
//	fmt.Println("flag is ", flag)
//}

func TestChecksign(t *testing.T) {
	var validTxSet []*types.Transaction
	var invalidTxSet []*types.InvalidTransactionRecord
	pool := poolInit()
	kec256Hash, encryption, _ := poolParameters()

	validTxSet, invalidTxSet = pool.CheckSign(transactionCases, kec256Hash, encryption)
	fmt.Println("valitxset is", validTxSet)
	fmt.Println("invalidatxset is", invalidTxSet)

}

func TestProcessBlockInVm(t *testing.T) {
	pool := poolInit()

	var invalidTxSet []*types.InvalidTransactionRecord
	err, _, merkleRoot, txRoot, receiptRoot, validtxs, invalidTxs := pool.ProcessBlockInVm(transactionCases, invalidTxSet, 1)
	if err != nil {
		log.Fatal("Put txmeta into database failed!")
	}
	log.Notice("merkleRoot is-->", merkleRoot, "txRoot is-->", txRoot, "receiptRoot is-->", receiptRoot, "validtxs is--->", validtxs, "invalidtxs is-->", invalidTxs)

}

func TestCommitBlock(t *testing.T) {

	var invalidTxSet []*types.InvalidTransactionRecord
	pool := poolInit()
	kec256Hash, _, grpcPeerMgr := poolParameters()
	validationEvent := new(event.ExeTxsEvent)
	validationEvent.IsPrimary = true
	validationEvent.SeqNo = 1
	validationEvent.Transactions = transactionCases
	//flag=true
	ev := new(event.CommitOrRollbackBlockEvent)
	ev.Flag = true
	ev.IsPrimary = false
	ev.SeqNo = 1
	hash := string([]byte{1, 2, 3, 4})
	ev.Hash = hash

	blockrecord := new(BlockRecord)
	blockrecord.InvalidTxs = invalidTxSet
	blockrecord.ValidTxs = transactionCases
	blockrecord.MerkleRoot = []byte{240, 53, 154, 92, 81, 30, 204, 173, 81, 8, 32, 30, 78, 239, 68, 145, 65, 201, 28, 93, 30, 243, 166, 130, 147, 197, 106, 124, 239, 110, 182, 169}
	blockrecord.ReceiptRoot = []byte{191, 216, 46, 62, 236, 4, 189, 185, 85, 225, 217, 180, 159, 235, 128, 197, 92, 218, 32, 85, 107, 228, 23, 248, 138, 236, 41, 174, 162, 208, 176, 49}
	blockrecord.TxRoot = []byte{135, 189, 209, 233, 54, 241, 174, 81, 206, 197, 43, 197, 240, 130, 139, 42, 127, 66, 148, 177, 248, 60, 95, 71, 83, 41, 150, 103, 65, 191, 50, 37}
	blockrecord.SeqNo = 1
	pool.blockCache.Add(ev.Hash, *blockrecord)
	log.Notice("pool's blockcache is ", pool.blockCache)
	pool.CommitBlock(*ev, kec256Hash, grpcPeerMgr)

	//flag=false
	ev1 := new(event.CommitOrRollbackBlockEvent)
	ev1.Flag = true
	ev1.IsPrimary = false
	ev1.SeqNo = 2
	hash1 := string([]byte{1, 4, 3, 9})
	ev1.Hash = hash1
	pool.CommitBlock(*ev1, kec256Hash, grpcPeerMgr)
	log.Notice("pool's blockcache is ", pool.blockCache)
}

func TestWriteBlock(t *testing.T) {

	pool := poolInit()
	log.Notice("pool test", pool.demandNumber)
	kec256Hash := crypto.NewKeccak256Hash("keccak256")

	newBlock := new(types.Block)
	newBlock.Transactions = make([]*types.Transaction, len(transactionCases))
	copy(newBlock.Transactions, transactionCases)
	currentChain := core.GetChainCopy()
	newBlock.ParentHash = currentChain.LatestBlockHash
	newBlock.MerkleRoot = []byte{240, 53, 154, 92, 81, 30, 204, 173, 81, 8, 32, 30, 78, 239, 68, 145, 65, 201, 28, 93, 30, 243, 166, 130, 147, 197, 106, 124, 239, 110, 182, 169}
	newBlock.TxRoot = []byte{135, 189, 209, 233, 54, 241, 174, 81, 206, 197, 43, 197, 240, 130, 139, 42, 127, 66, 148, 177, 248, 60, 95, 71, 83, 41, 150, 103, 65, 191, 50, 37}
	newBlock.ReceiptRoot = []byte{191, 216, 46, 62, 236, 4, 189, 185, 85, 225, 217, 180, 159, 235, 128, 197, 92, 218, 32, 85, 107, 228, 23, 248, 138, 236, 41, 174, 162, 208, 176, 49}

	newBlock.Number = 1
	newBlock.WriteTime = time.Now().UnixNano()
	newBlock.EvmTime = time.Now().UnixNano()
	newBlock.BlockHash = newBlock.Hash(kec256Hash).Bytes()

	newBlock2 := new(types.Block)
	newBlock2 = newBlock
	newBlock2.Number = 2
	WriteBlock(newBlock, kec256Hash, 1, false)
	WriteBlock(newBlock2, kec256Hash, 3, true)
}

func TestStoreInvalidResp(t *testing.T) {
	pool := poolInit()
	ev := new(event.RespInvalidTxsEvent)
	invalidTx := &types.InvalidTransactionRecord{}
	invalidTx.Tx = transactionCases[0]
	ev.Payload, _ = proto.Marshal(invalidTx)
	pool.StoreInvalidResp(*ev)

}

func TestResetStatus(t *testing.T) {
	ev := new(event.VCResetEvent)
	ev.SeqNo = 2
	pool := poolInit()
	pool.ResetStatus(*ev)
}
