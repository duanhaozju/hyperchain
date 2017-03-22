// author: Xiaoyi Wang
// email: wangxiaoyi@hyperchain.cn
// date: 16/11/1
// last modified: 16/11/1
// last Modified Author: Xiaoyi Wang
// change log: 1. new test for helper

package helper

import (
	"testing"
	pb "hyperchain/protos"
	"hyperchain/event"
	"reflect"
	"github.com/golang/protobuf/proto"
	"time"
)

func TestNewHelper(t *testing.T) {
	m := &event.TypeMux{}
	h := &helper{
		msgQ: m,
	}

	help := NewHelper(m)
	if !reflect.DeepEqual(help, h) {
		t.Error("error NewHelper")
	}
}

func TestInnerBroadcast(t *testing.T) {
	mux := &event.TypeMux{}
	h := NewHelper(mux)
	msg := &pb.Message{Type:pb.Message_CONSENSUS, Id:1}

	tmpMsg, _ := proto.Marshal(msg)
	broadcastEvent := event.BroadcastConsensusEvent{
		Payload: tmpMsg,
	}
	sub := mux.Subscribe(event.BroadcastConsensusEvent{})

	h.InnerBroadcast(msg)
	go func() {
		select {
		case e := <-sub.Chan():
			if ! reflect.DeepEqual(e.Data, broadcastEvent) {
				t.Fatal("Received wrong message from sub.Chan")
			}
		case <-time.After(1 * time.Second):
			t.Fatal("Timed out waiting for message to fire")
		}
	}()
}

func TestInnerUnicast(t *testing.T) {
	mux := &event.TypeMux{}
	h := NewHelper(mux)
	msg := &pb.Message{Type:pb.Message_CONSENSUS, Id:2}
	tmpMsg, _ := proto.Marshal(msg)

	unicastEvent := event.TxUniqueCastEvent{
		Payload:    tmpMsg,
		PeerId:        100,
	}
	sub := mux.Subscribe(event.TxUniqueCastEvent{})
	h.InnerUnicast(msg, 100)
	go func() {
		select {
		case e := <-sub.Chan():
			if !reflect.DeepEqual(e.Data, unicastEvent) {
				t.Fatal("Received wrong message from sub.Chan")
			}
		case <-time.After(1 * time.Second):
			t.Fatal("Timed out waiting for message to fire")
		}
	}()
}

func TestExecute(t *testing.T) {
	mux := &event.TypeMux{}
	h := NewHelper(mux)
	timestamp := time.Now().Unix()

	sub := mux.Subscribe(event.CommitOrRollbackBlockEvent{})
	go func() {
		select {
		case <-sub.Chan(): {

		}
		case <-time.After(1 * time.Second):
			t.Fatal("Timed out waiting for message to fire")
		}
	}()
	h.Execute(1, "hhhhhaaaaahhhh", true, false, timestamp)

}

func TestUpdateState(t *testing.T) {
	mux := &event.TypeMux{}
	h := NewHelper(mux)
	updateState := &pb.UpdateStateMessage{Id:12, SeqNo:21}

	tmpMsg, _ := proto.Marshal(updateState)

	updateStateEvent := event.ChainSyncReqEvent{
		Payload:	tmpMsg,
	}
	sub := mux.Subscribe(event.ChainSyncReqEvent{})
	go func() {
		select {
		case e := <-sub.Chan():
			if !reflect.DeepEqual(e.Data, updateStateEvent) {
				t.Fatal("Received wrong message from sub.Chan")
			}
		case <-time.After(1 * time.Second):
			t.Fatal("Timed out waiting for message to fire")
		}
	}()
	h.UpdateState(updateState)
}

func TestValidateBatch(t *testing.T) {
	mux := &event.TypeMux{}
	h := NewHelper(mux)
	timestamp := time.Now().Unix()
	validateEvent := event.ExeTxsEvent {
		Transactions:	nil,
		Timestamp:      timestamp,
		SeqNo:		123,
		View:		123,
		IsPrimary:	true,
	}
	sub := mux.Subscribe(event.ExeTxsEvent{})
	go func() {
		select {
		case e := <-sub.Chan():
			if !reflect.DeepEqual(e.Data, validateEvent) {
				t.Fatal("Received wrong message from sub.Chan")
			}
		case <-time.After(1 * time.Second):
			t.Fatal("Timed out waiting for message to fire")
		}
	}()

	h.ValidateBatch(nil, timestamp, 123, 123, true)
}

func TestVcReset(t *testing.T) {
	mux := &event.TypeMux{}
	h := NewHelper(mux)
	vcResetEvent := event.VCResetEvent{
		SeqNo: 12345,
	}
	sub := mux.Subscribe(vcResetEvent)
	go func(){
		select {
		case e := <-sub.Chan():
			if !reflect.DeepEqual(e.Data, vcResetEvent) {
				t.Fatal("Received wrong message from sub.Chan")
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Timed out waiting for message to fire")
		}
	}()
	h.VcReset(12345)
}

func TestInformPrimary(t *testing.T) {
	mux := &event.TypeMux{}
	h := NewHelper(mux)
	informPrimaryEvent := event.InformPrimaryEvent{
		Primary: 9,
	}
	sub := mux.Subscribe(event.InformPrimaryEvent{})
	go func() {
		select {
		case e := <-sub.Chan():
			if !reflect.DeepEqual(e.Data, informPrimaryEvent) {
				t.Fatal("Received wrong message from sub.Chan")
			}
		case <-time.After(1 * time.Second):
			t.Fatal("Timed out waiting for message to fire")
		}
	}()
	h.InformPrimary(9)

}