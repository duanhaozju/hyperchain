package pbft


import (
	"testing"
	"hyperchain/event"
	//"github.com/golang/protobuf/proto"
	//"hyperchain/protos"
	"hyperchain/consensus/helper"
	"hyperchain/consensus/events"
	//"reflect"
	//"time"
	"time"
)


func TestReceivePrePrePare(t *testing.T){
	//msgQ :=new(event.TypeMux)
	//h:=helper.NewHelper(msgQ)
	//batch:=GetPlugin(3, h)
	//
	//
	//
	//req:=&Request{
	//	Timestamp :time.Now().UnixNano(),
	//	Payload   :[]byte{},
	//	ReplicaId :2,
	//	Signature :[]byte{},
	//}
	//reqbatch:=&RequestBatch{
	//	Batch:[]*Request{req,req},
	//}
	//preprep := &PrePrepare{
	//	View:           1,
	//	SequenceNumber: 2,
	//	BatchDigest:    "1233",
	//	RequestBatch:   reqbatch,
	//	ReplicaId:      2,
	//}
	//msg := pbftMsgHelper(&Message{Payload: &Message_PrePrepare{PrePrepare: preprep}}, 2)
	//pbftPayload, _ := proto.Marshal(msg)
	//
	//
	//pbft:=&BatchMessage{Payload: &BatchMessage_PbftMessage{PbftMessage: pbftPayload}}
	//b3,err:=proto.Marshal(pbft)
	//if err!=nil{
	//	return
	//}
	//msg1:=&protos.Message{
	//	Type      :1 ,
	//	Timestamp :233333,
	//	Payload   : b3,
	//	Id        :22222,
	//}
	//
	//b1,err1:=proto.Marshal(msg1)
	//if err1==nil{
	//	logger.Info("**********>  send message:",reflect.TypeOf(msg1),msg1.Type)
	//	batch.RecvMsg(b1)
	//}

}

func TestClientRequest(t *testing.T){
	//msgQ :=new(event.TypeMux)
	//h:=helper.NewHelper(msgQ)
	//batch:=GetPlugin(0, h)
	//
	//msg0:=&protos.Message{
	//	Type      :0 ,
	//	Timestamp :233333,
	//	Payload   : []byte {'a', 'b', 'c', 'd'},
	//	Id        :22222,
	//}
	//
	//b0,err0:=proto.Marshal(msg0)
	//if err0!=nil{
	//	return
	//}
	//
	//logger.Info("**********>  send message:",reflect.TypeOf(msg0),msg0.Type)
	//batch.RecvMsg(b0)

	time.Sleep(10*time.Second)
	//batch.Close()
}

func TestBatchInit(t *testing.T){
	msgQ :=new(event.TypeMux)
	h:=helper.NewHelper(msgQ)
	batchObj:=&batch{
		localID:	3,
		helperImpl:	h,
	}
	type testEvent struct {} //ToDo for test
	batchObj.manager=events.NewManagerImpl()
	batchObj.manager.SetReceiver(batchObj)
	etf := events.NewTimerFactoryImpl(batchObj.manager)
	batchObj.pbft = newPbftCore(3, config, batchObj, etf)
	batchObj.manager.Start()

	//batchObj.manager.Queue()<-&testEvent{}
}
