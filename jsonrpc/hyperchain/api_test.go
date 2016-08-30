package hyperchain

import (
	"testing"
	"fmt"
	"hyperchain/core"
	"log"
	"hyperchain/common"
	"hyperchain/manager"
	"hyperchain/p2p"
	"hyperchain/event"
	"hyperchain/crypto"
)

func TestSendTransaction(t *testing.T) {

	isSuccess := SendTransaction(TxArgs{
		From: "addressFrom",
		To: "addressTo",
		Value: "12",
	})

	fmt.Println(isSuccess)

	balanceIns,err := core.GetBalanceIns()
	if err != nil {
		log.Fatalf("%v", err)
	}
	balanceIns.PutCacheBalance(common.BytesToAddress([]byte("addressFrom")),[]byte("13"))


	eventMux := new(event.TypeMux)
	peerManager := new(p2p.GrpcPeerManager)
	fetcher := core.NewFetcher()
	encryption :=crypto.NewEcdsaEncrypto("ecdsa")
	encryption.GeneralKey(string(8012))
	commonHash:=crypto.NewKeccak256Hash("keccak256")

	manager.NewProtocolManager(peerManager, eventMux, fetcher, nil, encryption, commonHash)

	isSuccess2 := SendTransaction(TxArgs{
		From: "addressFrom",
		To: "addressTo",
		Value: "12",
	})

	fmt.Println(isSuccess2)
}


