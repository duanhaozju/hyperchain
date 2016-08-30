// author: Lizhong kuang
// date: 16-8-29
// last modified: 16-8-29

package manager

import (
	"testing"

	"fmt"
	"hyperchain/event"
	"time"
	"github.com/golang/protobuf/proto"
	"crypto/ecdsa"
	"hyperchain/core/types"
	"hyperchain/crypto"
)


func TestAliveEvent(t *testing.T){
	manager := &ProtocolManager{
		eventMux:    new(event.TypeMux),
		quitSync:    make(chan struct{}),



	}
	manager.aLiveSub = manager.eventMux.Subscribe(event.AliveEvent{})

	go newEvent(manager)

	for obj := range manager.aLiveSub.Chan() {

		switch ev := obj.Data.(type) {
		case event.AliveEvent:
			fmt.Print(ev.Payload)
		}
	}


}
func newEvent(manager *ProtocolManager)  {
	fmt.Println("1")
	for i := 0; i < 5; i += 1 {
		fmt.Println("hahah")
		//var transaction types.Transaction
		transaction:=types.Transaction{
			From:[]byte{0x00, 0x00, 0x03, 0xe8},
			To:[]byte{0x00, 0x00, 0x03},
		}

		payLoad, err := proto.Marshal(&transaction)
		/*fmt.Println("new payload")
		fmt.Println(payLoad)*/
		if err != nil {
			return
		}


		//eventmux := new(event.TypeMux)
		manager.eventMux.Post(event.ConsensusEvent{payLoad})


		//manager.eventMux.Post(event.ConsensusEvent{[]byte{0x00, 0x00, 0x03, 0xe8}})
		time.Sleep(2)

	}

}


func TestDecodeTx(t *testing.T){
	kec256Hash:=crypto.NewKeccak256Hash("keccak256")
	encryption :=crypto.NewEcdsaEncrypto("ecdsa")
	encryption.GeneralKey("124")
	manager := &ProtocolManager{
		eventMux:    new(event.TypeMux),
		quitSync:    make(chan struct{}),
		commonHash:kec256Hash,
		encryption:encryption,


	}
	manager.aLiveSub = manager.eventMux.Subscribe(event.ConsensusEvent{})

	go newEvent(manager)

	for obj := range manager.aLiveSub.Chan() {

		switch ev := obj.Data.(type) {
		case event.ConsensusEvent:
			var transaction types.Transaction
			//decode tx
			proto.Unmarshal(ev.Payload, &transaction)
			fmt.Println("unmarshal payload")
			fmt.Println(transaction)
			//hash tx
			h := transaction.SighHash(manager.commonHash)
			fmt.Println(h[:])
			key, err := manager.encryption.GetKey()
			//if value, ok := key.(*ecdsa.PrivateKey); ok {


				switch key.(type){
			case *ecdsa.PrivateKey:
				actualKey:=key.(*ecdsa.PrivateKey)
				sign, err := manager.encryption.Sign(h[:], actualKey)
				fmt.Println(actualKey)
				fmt.Println(h[:])
				fmt.Println(sign)
				if err != nil {
					fmt.Print(err)
				}

				//fmt.Println(sign)
				//fmt.Println(manager.encryption.UnSign(h[:],sign))
				transaction.Signature = sign
				//encode tx
				payLoad, err := proto.Marshal(&transaction)
				if err != nil {
					return
				}
				fmt.Println("marshal payload")
				fmt.Println(payLoad)

				//manager.consenter.RecvMsg(payLoad)
			}
			if err != nil {
				return
			}
		}
	}


}
