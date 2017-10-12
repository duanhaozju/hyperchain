package rbft

import (
	"github.com/facebookgo/ensure"
	//"github.com/gogo/protobuf/proto"
	"hyperchain/common"
	//"hyperchain/manager/protos"
	"strconv"
	"testing"
	//"time"
)

func TestPbftImpl_NewPbft(t *testing.T) {

	//new PBFT
	rbft, conf, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ensure.Nil(t, err)

	ensure.DeepEqual(t, rbft.namespace, "global-"+strconv.Itoa(int(rbft.id)))
	ensure.DeepEqual(t, rbft.activeView, uint32(1))
	ensure.DeepEqual(t, rbft.f, (rbft.N-1)/3)
	ensure.DeepEqual(t, rbft.N, conf.GetInt("self.N"))
	ensure.DeepEqual(t, rbft.h, uint64(0))
	ensure.DeepEqual(t, rbft.id, uint64(conf.GetInt64(common.C_NODE_ID)))
	ensure.DeepEqual(t, rbft.K, uint64(10))
	ensure.DeepEqual(t, rbft.logMultiplier, uint64(4))
	ensure.DeepEqual(t, rbft.L, rbft.logMultiplier*rbft.K)
	ensure.DeepEqual(t, rbft.seqNo, uint64(0))
	ensure.DeepEqual(t, rbft.view, uint64(0))

	//Test Consenter interface

	rbft.Start()

	//pbft.RecvMsg()
	rbft.Close()

}

//pbMsg := &protos.Message{
//Type:      protos.Message_NULL_REQUEST,
//Payload:   nil,
//Timestamp: time.Now().UnixNano(),
//Id:        id,
//}
//func TestProcessNullRequest(t *testing.T) {
//	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
//	ensure.Nil(t, err)
//	pbMsg := &protos.Message{
//		Type:      protos.Message_NULL_REQUEST,
//		Payload:   nil,
//		Timestamp: time.Now().UnixNano(),
//		Id:        1,
//	}
//	event, err := proto.Marshal(pbMsg)
//	rbft.RecvMsg(event)
//	time.Sleep(3 * time.Second)
//	err = CleanData(rbft.namespace)
//	ensure.Nil(t, err)
//
//}

//type mockEvent struct {
//	info string
//}
//
//type mockReceiver struct {
//	processEventImpl func(event event) event
//}
//
//func (mr *mockReceiver) ProcessEvent(event event) event {
//	if mr.processEventImpl != nil {
//		return mr.processEventImpl(event)
//	}
//	return nil
//}

//func TestPostRequestEvent(t *testing.T)  {
//	pbft := new(pbftImpl)
//
//	tx := &types.Transaction{Id:123}
//	evts := make(chan event)
//
//	pbft.batchManager = events.NewManagerImpl()
//	pbft.batchManager.SetReceiver(&mockReceiver{
//		processEventImpl : func(event event) event{
//			evts <- event
//			return nil
//		},
//	})
//	pbft.batchManager.Start()
//	pbft.postRequestEvent(tx)
//
//	go func() {
//		select {
//		case e := <- evts:
//			if !reflect.DeepEqual(e, tx) {
//				t.Error("error postRequestEvent ")
//			}
//		case <- time.After(3*time.Second) :
//			t.Error("error postRequestEvent event not received")
//		}
//	}()
//}

//func TestPostPbftEvent(t *testing.T)  {
//	pbft := new(pbftImpl)
//	mEvent := &mockEvent{info:"pbftEvent"}
//
//	evts := make(chan event)
//	pbft.pbftEventManager = events.NewManagerImpl()
//	pbft.pbftEventManager.SetReceiver(&mockReceiver{
//		processEventImpl : func(event event) event{
//			evts <- event
//			return nil
//		},
//	})
//	pbft.pbftEventManager.Start()
//
//	pbft.postPbftEvent(mEvent)
//
//	go func() {
//		select {
//		case e := <- evts:
//			if !reflect.DeepEqual(e, mEvent) {
//				t.Error("error postPbftEvent ")
//			}
//		case <- time.After(3*time.Second) :
//			t.Error("error postPbftEvent event not received")
//		}
//	}()
//}
