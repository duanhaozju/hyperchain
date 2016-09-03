package api

import (
	"testing"
	"hyperchain/core"
	"hyperchain/common"
	"hyperchain/manager"
	"hyperchain/p2p"
	"hyperchain/event"
	"hyperchain/crypto"
	"github.com/stretchr/testify/assert"
)

//var pm *manager.ProtocolManager

func initPM() *manager.ProtocolManager{
	eventMux := new(event.TypeMux)
	peerManager := new(p2p.GrpcPeerManager)
	fetcher := core.NewFetcher()
	encryption :=crypto.NewEcdsaEncrypto("ecdsa")
	encryption.GeneralKey(string(8012))
	commonHash:=crypto.NewKeccak256Hash("keccak256")
	//blockPool:=core.NewBlockPool(eventMux)

	return manager.NewProtocolManager(nil,peerManager, eventMux, fetcher, nil, encryption, commonHash)
}

func TestSendTransaction(t *testing.T) {

	initPM()

	isSuccess := SendTransaction(TxArgs{
		From: "addressFrom",
		To: "addressTo",
		Value: "12",
	})

	assert.Equal(t,false, isSuccess, "they should be equal")

	balanceIns,err := core.GetBalanceIns()
	if err != nil {
		log.Fatalf("%v", err)
	}
	balanceIns.PutCacheBalance(common.BytesToAddress([]byte("addressFrom")),[]byte("13"))

	isSuccess2 := SendTransaction(TxArgs{
		From: "addressFrom",
		To: "addressTo",
		Value: "12",
	})

	assert.Equal(t,true, isSuccess2, "they should be equal")
}

//func TestGetAllTransactions(t *testing.T) {
//
//	tx := types.NewTransaction([]byte("addressFrom"),[]byte("addressTo"),[]byte("12"))
//	tx.TimeStamp = time.Now().Unix()
//
//	pm := initPM()
//
//	tx.SighHash(pm.commonHash)
//
//}


