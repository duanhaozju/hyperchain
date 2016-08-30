// author: Lizhong kuang
// date: 16-8-29
// last modified: 16-8-29

package manager

import (
	"testing"


	"hyperchain/event"

	"github.com/golang/protobuf/proto"

	"hyperchain/core/types"
	"hyperchain/crypto"

)


func TestAliveEvent(t *testing.T){
	manager := &ProtocolManager{
		eventMux:    new(event.TypeMux),
		quitSync:    make(chan struct{}),



	}
	manager.aLiveSub = manager.eventMux.Subscribe(event.AliveEvent{})

	//go newEvent(manager)



}
/*func newEvent(manager *ProtocolManager)  {
	for i := 0; i < 5; i += 1 {


		//var transaction types.Transaction

		transaction:=types.Transaction{
			From:[]byte("0x3e514052d804d4934e2ad0fc0818b9fc1431201d"),
			To:[]byte{0x00, 0x00, 0x03},
		}

		payLoad, err := proto.Marshal(&transaction)

		if err != nil {
			return
		}



		manager.eventMux.Post(event.ConsensusEvent{payLoad})




	}

}*/


/*func TestDecodeTx(t *testing.T){
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
			//hash tx
			h := transaction.SighHash(manager.commonHash)

			key, err := manager.encryption.GetKey()




				switch key.(type){
			case *ecdsa.PrivateKey:
				actualKey:=key.(*ecdsa.PrivateKey)
				sign, err := manager.encryption.Sign(h[:], actualKey)

				fmt.Println(common.Bytes2Hex(manager.encryption.PrivKeyToAddress(*actualKey)[:]))

				if err != nil {
					fmt.Print(err)
				}

				fmt.Println(manager.encryption.UnSign(h[:],sign))
				transaction.Signature = sign
				//encode tx
				payLoad, err := proto.Marshal(&transaction)
				if err != nil {
					return
				}
				fmt.Println("marshal payload")
				fmt.Println(payLoad)


			}
			if err != nil {
				return
			}

		}
	}


}*/

func TestTransformTx(t *testing.T){

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
	transaction:=&types.Transaction{
		From:[]byte{0x00, 0x00, 0x03, 0xe8},
		To:[]byte{0x00, 0x00, 0x03},
	}

	payLoad, err := proto.Marshal(transaction)

	if err != nil {
		return
	}
	manager.transformTx(payLoad)

}
