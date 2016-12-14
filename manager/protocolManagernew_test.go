// author: Yijun Xie
// date: 2016-11-03
// last modified:
package manager

import (
	"hyperchain/accounts"
	"hyperchain/consensus"
	"hyperchain/core"
	"hyperchain/core/blockpool"
	"hyperchain/core/types"
	"hyperchain/crypto"
	"hyperchain/event"
	"hyperchain/p2p"
	"hyperchain/p2p/peermessage"
	"hyperchain/recovery"
	//"sync"
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

func poolParameters() (crypto.CommonHash, crypto.Encryption, p2p.PeerManager) {
	encryption := crypto.NewEcdsaEncrypto("ecdsa")
	kec256Hash := crypto.NewKeccak256Hash("keccak256")
	path := "../config/local_peerconfig.json"
	grpcPeerMgr := p2p.NewGrpcManager(path, 1)

	aliveChan := make(chan bool)
	eventMux := new(event.TypeMux)
	go grpcPeerMgr.Start(aliveChan, eventMux)

	return kec256Hash, encryption, grpcPeerMgr
}

func initmanager() *ProtocolManager {
	core.InitDB("G:/hyperchainDB", 8030)
	eventMux := new(event.TypeMux)
	consenter := new(consensus.Consenter)
	pool := blockpool.NewBlockPool(eventMux, *consenter)

	path := "../config/local_peerconfig.json"
	grpcPeerMgr := p2p.NewGrpcManager(path, 1)
	kec256Hash := crypto.NewKeccak256Hash("keccak256")

	am := new(accounts.AccountManager)

	syncReplica := false

	return NewProtocolManager(pool, grpcPeerMgr, eventMux, *consenter, am, kec256Hash, 1, syncReplica)

}

//test new can't success ,because need p2p module support
//func TestNew(t *testing.T) {
//	core.InitDB("G:/hyperchainDB", 8030)
//	eventMux := new(event.TypeMux)
//	consenter := new(consensus.Consenter)
//	pool := blockpool.NewBlockPool(eventMux, *consenter)

//	path := "../config/local_peerconfig.json"
//	grpcPeerMgr := p2p.NewGrpcManager(path, 1)
//	kec256Hash := crypto.NewKeccak256Hash("keccak256")

//	am := new(accounts.AccountManager)

//	syncReplica := false
//	New(eventMux, pool, grpcPeerMgr, *consenter, am, kec256Hash, 1, 1, syncReplica)
//}

func TestNewProtocolManager(t *testing.T) {
	initmanager()
}

func TestGetEventObject(t *testing.T) {
	eventMux := new(event.TypeMux)
	eventMuxAll = eventMux
	eventMullTest := GetEventObject()
	if eventMullTest == eventMuxAll {
		log.Notice("geteventobject() is right")
	}
}

func TestStart(t *testing.T) {
	manager := initmanager()
	manager.Start()
}

func TestUpdateRequire(t *testing.T) {
	manager := initmanager()

	kec256Hash := crypto.NewKeccak256Hash("keccak256")
	newBlock := new(types.Block)
	newBlock.Transactions = make([]*types.Transaction, len(transactionCases))
	copy(newBlock.Transactions, transactionCases)
	currentChain := core.GetChainCopy()
	newBlock.ParentHash = currentChain.LatestBlockHash
	newBlock.MerkleRoot = []byte{240, 53, 154, 92, 81, 30, 204, 173, 81, 8, 32, 30, 78, 239, 68, 145, 65, 201, 28, 93, 30, 243, 166, 130, 147, 197, 106, 124, 239, 110, 182, 169}
	newBlock.TxRoot = []byte{135, 189, 209, 233, 54, 241, 174, 81, 206, 197, 43, 197, 240, 130, 139, 42, 127, 66, 148, 177, 248, 60, 95, 71, 83, 41, 150, 103, 65, 191, 50, 37}
	newBlock.ReceiptRoot = []byte{191, 216, 46, 62, 236, 4, 189, 185, 85, 225, 217, 180, 159, 235, 128, 197, 92, 218, 32, 85, 107, 228, 23, 248, 138, 236, 41, 174, 162, 208, 176, 49}

	newBlock.Number = 6
	newBlock.WriteTime = time.Now().UnixNano()
	newBlock.EvmTime = time.Now().UnixNano()
	newBlock.BlockHash = newBlock.Hash(kec256Hash).Bytes()

	manager.updateRequire(newBlock)
}

func TestRecordReplicaStatus(t *testing.T) {
	manager := initmanager()
	status := &recovery.ReplicaStatus{}
	chain := &types.Chain{}
	addr := &peermessage.PeerAddress{}
	addr.IP = "192.168.0.1"
	addr.ID = 20160028
	addr.Port = 8080
	addr.Hash = "0x736d49571ac5ac521b19abfe96d6d04ce15c8c5625b98958af9b93a1dfe594e6"
	tempaddr, _ := proto.Marshal(addr)
	status.Addr = tempaddr
	chain.LatestBlockHash = []byte{135, 189, 209, 233, 54, 241, 174, 81, 206, 197, 43, 197, 240, 130, 139, 42}
	chain.ParentBlockHash = []byte{191, 216, 46, 62, 236, 4, 189, 185, 85, 225, 217, 180, 159, 235, 128, 197, 92}
	chain.Height = 5
	tempchain, _ := proto.Marshal(chain)

	status.Addr = tempaddr
	status.Chain = tempchain

	ev := new(event.ReplicaStatusEvent)
	ev.Payload, _ = proto.Marshal(status)
	manager.RecordReplicaStatus(*ev)
	log.Notice("manager's replicaStatus is", manager.replicaStatus)
}

//func TestPackReplicaStatus(t *testing.T) {
//	manager := initmanager()

//	path := "../config/local_peerconfig.json"
//	grpcPeerMgr := p2p.NewGrpcManager(path, 1)

//	aliveChan := make(chan bool)
//	eventMux := new(event.TypeMux)
//	var syncmux chan bool

//	go func() {
//		grpcPeerMgr.Start(aliveChan, eventMux)
//		syncmux <- true
//	}()
//	<-syncmux
//	log.Notice("get a grpc peer")
//	b1, b2 := manager.packReplicaStatus()
//	log.Notice("b1 is", b1, "b2 is", b2)

//}
