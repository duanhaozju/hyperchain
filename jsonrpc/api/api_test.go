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
	"hyperchain/accounts"
)

//var pm *manager.ProtocolManager

func initPM() *manager.ProtocolManager{
	eventMux := new(event.TypeMux)
	peerManager := new(p2p.GrpcPeerManager)
	fetcher := core.NewFetcher()
	commonHash:=crypto.NewKeccak256Hash("keccak256")
	//blockPool:=core.NewBlockPool(eventMux)

	//init encryption object
	encryption := crypto.NewEcdsaEncrypto("ecdsa")
	encryption.GeneralKey("8012")

	scryptN := accounts.StandardScryptN
	scryptP := accounts.StandardScryptP
	keydir := "../keystore/"
	am := accounts.NewAccountManager(keydir,encryption, scryptN, scryptP)

	return manager.NewProtocolManager(nil, peerManager, eventMux, fetcher, nil, am, commonHash)
}

func TestSendTransaction(t *testing.T) {

	initPM()

	balanceIns,err := core.GetBalanceIns()
	if err != nil {
		log.Fatalf("%v", err)
	}
	balanceIns.DeleteCacheBalance(common.BytesToAddress([]byte("000000000000000000000000000000adressFrom")))

	isSuccess := SendTransaction(TxArgs{
		From: "000000000000000000000000000000adressFrom",
		To: "00000000000000000000000000000000adressTo",
		Value: "12",
	})
	assert.Equal(t,false, isSuccess, "they should be equal")


	balanceIns.PutCacheBalance(common.BytesToAddress([]byte("000000000000000000000000000000adressFrom")),[]byte("13"))

	isSuccess2 := SendTransaction(TxArgs{
		From: "000000000000000000000000000000adressFrom",
		To: "00000000000000000000000000000000adressTo",
		Value: "12",
	})

	assert.Equal(t,true, isSuccess2, "they should be equal")
}

//func TestGetAllTransactions(t *testing.T) {
//
//	tx := types.NewTransaction([]byte("00000000000000000000000000000adressFrom"),[]byte("addressTo"),[]byte("12"))
//	tx.TimeStamp = time.Now().Unix()
//
//	pm := initPM()
//
//	tx.SighHash(pm.commonHash)
//
//}


