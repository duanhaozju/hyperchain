//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package helper

import (
	"testing"
	"time"

	"github.com/hyperchain/hyperchain/consensus"
	"github.com/hyperchain/hyperchain/manager/appstat"
	"github.com/hyperchain/hyperchain/manager/event"
	pb "github.com/hyperchain/hyperchain/manager/protos"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func TestNewHelper(t *testing.T) {
	im := &event.TypeMux{}
	em := &event.TypeMux{}

	h := &helper{
		innerMux:    im,
		externalMux: em,
	}

	help := NewHelper(im, em)
	assert.Equal(t, h, help, "error NewHelper")
}

func TestInnerBroadcast(t *testing.T) {
	im := &event.TypeMux{}
	em := &event.TypeMux{}
	h := NewHelper(im, em)
	msg := &pb.Message{Type: pb.Message_CONSENSUS, Id: 1}

	tmpMsg, _ := proto.Marshal(msg)
	broadcastEvent := event.BroadcastConsensusEvent{
		Payload: tmpMsg,
	}
	sub := im.Subscribe(event.BroadcastConsensusEvent{})

	go h.InnerBroadcast(msg)

	select {
	case e := <-sub.Chan():
		assert.Equal(t, broadcastEvent, e.Data, "Received wrong message from sub.Chan")
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for message to fire")
	}
}

func TestInnerUnicast(t *testing.T) {
	im := &event.TypeMux{}
	em := &event.TypeMux{}
	h := NewHelper(im, em)
	msg := &pb.Message{Type: pb.Message_CONSENSUS, Id: 2}
	tmpMsg, _ := proto.Marshal(msg)

	unicastEvent := event.TxUniqueCastEvent{
		Payload: tmpMsg,
		PeerId:  100,
	}
	sub := im.Subscribe(event.TxUniqueCastEvent{})
	go h.InnerUnicast(msg, 100)

	select {
	case e := <-sub.Chan():
		assert.Equal(t, unicastEvent, e.Data, "Received wrong message from sub.Chan")
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for message to fire")
	}
}

func TestExecute(t *testing.T) {
	im := &event.TypeMux{}
	em := &event.TypeMux{}
	h := NewHelper(im, em)
	timestamp := time.Now().Unix()

	sub := im.Subscribe(event.CommitEvent{})

	go h.Execute(1, "hhhhhaaaaahhhh", true, false, timestamp)

	select {
	case e := <-sub.Chan():
		if commit, ok := e.Data.(event.CommitEvent); !ok {
			t.Error("Received wrong type message from sub.Chan")
		} else {
			assert.Equal(t, "hhhhhaaaaahhhh", commit.Hash)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for message to fire")
	}
}

func TestUpdateState(t *testing.T) {
	im := &event.TypeMux{}
	em := &event.TypeMux{}
	h := NewHelper(im, em)

	updateStateEvent := event.ChainSyncReqEvent{
		Id:              1,
		TargetHeight:    10,
		TargetBlockHash: []byte{1, 2, 3},
		Replicas:        nil,
	}
	sub := im.Subscribe(event.ChainSyncReqEvent{})

	go h.UpdateState(1, 10, []byte{1, 2, 3}, nil)

	select {
	case e := <-sub.Chan():
		assert.Equal(t, updateStateEvent, e.Data, "Received wrong message from sub.Chan")
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for message to fire")
	}
}

func TestValidateBatch(t *testing.T) {
	im := &event.TypeMux{}
	em := &event.TypeMux{}
	h := NewHelper(im, em)
	timestamp := time.Now().Unix()
	validateEvent := event.ValidationEvent{
		Digest:       "something",
		Transactions: nil,
		Timestamp:    timestamp,
		SeqNo:        123,
		View:         123,
		IsPrimary:    true,
	}
	sub := im.Subscribe(event.ValidationEvent{})

	go h.ValidateBatch("something", nil, timestamp, 123, 123, true)

	select {
	case e := <-sub.Chan():
		assert.Equal(t, validateEvent, e.Data, "Received wrong message from sub.Chan")
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for message to fire")
	}
}

func TestVcReset(t *testing.T) {
	im := &event.TypeMux{}
	em := &event.TypeMux{}
	h := NewHelper(im, em)
	vcResetEvent := event.VCResetEvent{
		SeqNo: 12345,
		View:  1,
	}
	sub := im.Subscribe(vcResetEvent)

	go h.VcReset(12345, 1)

	select {
	case e := <-sub.Chan():
		assert.Equal(t, vcResetEvent, e.Data, "Received wrong message from sub.Chan")
	case <-time.After(5 * time.Second):
		t.Fatal("Timed out waiting for message to fire")
	}
}

func TestInformPrimary(t *testing.T) {
	im := &event.TypeMux{}
	em := &event.TypeMux{}
	h := NewHelper(im, em)
	informPrimaryEvent := event.InformPrimaryEvent{
		Primary: 9,
	}
	sub := im.Subscribe(event.InformPrimaryEvent{})
	go h.InformPrimary(9)

	select {
	case e := <-sub.Chan():
		assert.Equal(t, informPrimaryEvent, e.Data, "Received wrong message from sub.Chan")
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for message to fire")
	}
}

func TestBroadcastAddNode(t *testing.T) {
	im := &event.TypeMux{}
	em := &event.TypeMux{}
	h := NewHelper(im, em)

	msg := &pb.Message{Type: pb.Message_CONSENSUS, Id: 2}
	tmpMsg, _ := proto.Marshal(msg)
	broadcastEvent := event.BroadcastNewPeerEvent{
		Payload: tmpMsg,
	}

	sub := im.Subscribe(event.BroadcastNewPeerEvent{})
	go h.BroadcastAddNode(msg)

	select {
	case e := <-sub.Chan():
		assert.Equal(t, broadcastEvent, e.Data, "Received wrong message from sub.Chan")
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for message to fire")
	}

}

func TestBroadcastDelNode(t *testing.T) {
	im := &event.TypeMux{}
	em := &event.TypeMux{}
	h := NewHelper(im, em)
	msg := &pb.Message{Type: pb.Message_CONSENSUS, Id: 2}
	tmpMsg, _ := proto.Marshal(msg)
	broadcastEvent := event.BroadcastDelPeerEvent{
		Payload: tmpMsg,
	}
	sub := im.Subscribe(event.BroadcastDelPeerEvent{})
	go h.BroadcastDelNode(msg)

	select {
	case e := <-sub.Chan():
		assert.Equal(t, broadcastEvent, e.Data, "Received wrong message from sub.Chan")
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for message to fire")
	}
}

func TestUpdateTable(t *testing.T) {
	im := &event.TypeMux{}
	em := &event.TypeMux{}
	h := NewHelper(im, em)
	msg := &pb.Message{Type: pb.Message_CONSENSUS, Id: 2}
	tmpMsg, _ := proto.Marshal(msg)

	broadcastEvent := event.UpdateRoutingTableEvent{
		Payload: tmpMsg,
		Type:    true,
	}
	sub := im.Subscribe(event.UpdateRoutingTableEvent{})
	go h.UpdateTable(tmpMsg, true)

	select {
	case e := <-sub.Chan():
		assert.Equal(t, broadcastEvent, e.Data, "Received wrong message from sub.Chan")
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for message to fire")
	}
}

func TestPostExternal(t *testing.T) {
	im := &event.TypeMux{}
	em := &event.TypeMux{}
	h := NewHelper(im, em)
	informPrimaryEvent := event.InformPrimaryEvent{
		Primary: 9,
	}
	sub := em.Subscribe(event.InformPrimaryEvent{})

	go h.PostExternal(informPrimaryEvent)

	select {
	case e := <-sub.Chan():
		assert.Equal(t, informPrimaryEvent, e.Data, "Received wrong message from sub.Chan")
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for message to fire")
	}
}

func TestSendFilterEvent(t *testing.T) {
	im := &event.TypeMux{}
	em := &event.TypeMux{}
	h := NewHelper(im, em)
	msg := "test"

	filterSystemStatusEvent := event.FilterSystemStatusEvent{
		Module:  appstat.ExceptionModule_Consenus,
		Status:  appstat.Normal,
		Subtype: appstat.ExceptionSubType_ViewChange,
		Message: msg,
	}
	sub := em.Subscribe(event.FilterSystemStatusEvent{})

	go h.SendFilterEvent(consensus.FILTER_View_Change_Finish, msg)

	select {
	case e := <-sub.Chan():
		assert.Equal(t, filterSystemStatusEvent, e.Data, "Received wrong message from sub.Chan")
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for message to fire")
	}
}
