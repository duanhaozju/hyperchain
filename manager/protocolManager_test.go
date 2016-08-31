// author: Lizhong kuang
// date: 16-8-29
// last modified: 16-8-29

package manager

import (
	"testing"


	"hyperchain/event"




	"fmt"

	"time"

)



func newEvent(manager *ProtocolManager) {
	for i := 0; i < 5; i += 1 {


		manager.eventMux.Post(event.AliveEvent{true})


	}

}

func receive(manager *ProtocolManager){

	manager.aLiveSub = manager.eventMux.Subscribe(event.AliveEvent{})
	for obj := range manager.aLiveSub.Chan() {
		switch ev := obj.Data.(type) {

		case event.AliveEvent:
			fmt.Println(ev.Payload)
		}

	}
}
func TestAliveEvent(t *testing.T) {
	manager := &ProtocolManager{
		eventMux:    new(event.TypeMux),
		quitSync:    make(chan struct{}),

	}
	go receive(manager)
	for i := 0; i < 100; i += 1 {
		if i==0{
			time.Sleep(1234589112)
		}
		go newEvent(manager)
		go newEvent(manager)
		go newEvent(manager)
		go newEvent(manager)

	}
}

/*func TestCommitNewBlock(t *testing.T) {

	transaction := &types.Transaction{
		TimeStamp:12,
	}
	payLoadT, _ := proto.Marshal(transaction)

	msg := &protos.Message{
		Type: protos.Message_TRANSACTION,
		Payload: payLoadT,
		Timestamp: time.Now().UnixNano(),
		Id: 0,
	}
	var Batch []*protos.Message
	Batch=append(Batch,msg)

     msgList := &protos.ExeMessage{
	     Batch:Batch,
     }
	payload, _ := proto.Marshal(msgList)

	manager := &ProtocolManager{
		eventMux:    new(event.TypeMux),
		quitSync:    make(chan struct{}),

	}
	//manager.commitNewBlock(payload)
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

/*func TestTransformTx(t *testing.T){

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

}*/
