//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package helper

import (
	"time"

	"hyperchain/event"
	pb "hyperchain/protos"

	"github.com/golang/protobuf/proto"
	"hyperchain/core/types"
)

type helper struct {
	msgQ *event.TypeMux
}

type Stack interface {
	InnerBroadcast(msg *pb.Message) error
	InnerUnicast(msg *pb.Message, to uint64) error
	Execute(seqNo uint64, hash string, flag bool, isPrimary bool, time int64) error
	UpdateState(updateState *pb.UpdateStateMessage) error
	ValidateBatch(txs []*types.Transaction, timeStamp int64, seqNo uint64, view uint64, isPrimary bool) error
	VcReset(seqNo uint64) error
	InformPrimary(primary uint64) error
	BroadcastAddNode(msg *pb.Message) error
	BroadcastDelNode(msg *pb.Message) error
	UpdateTable(payload []byte, flag bool) error
	ClearValidateCache() error
}

// InnerBroadcast broadcast the consensus message between vp nodes
func (h *helper) InnerBroadcast(msg *pb.Message) error {

	tmpMsg, err := proto.Marshal(msg)

	if err != nil {
		return err
	}

	broadcastEvent := event.BroadcastConsensusEvent{
		Payload: tmpMsg,
	}

	// Post the event to outer
	go h.msgQ.Post(broadcastEvent)

	return nil
}

// InnerUnicast unicast the transaction message between to primary
func (h *helper) InnerUnicast(msg *pb.Message, to uint64) error {

	tmpMsg, err := proto.Marshal(msg)

	if err != nil {
		return err
	}

	unicastEvent := event.TxUniqueCastEvent{
		Payload: tmpMsg,
		PeerId:  to,
	}

	// Post the event to outer
	go h.msgQ.Post(unicastEvent)

	return nil
}

// Execute transfers the transactions decided by consensus to outer
func (h *helper) Execute(seqNo uint64, hash string, flag bool, isPrimary bool, timestamp int64) error {

	writeEvent := event.CommitOrRollbackBlockEvent{
		SeqNo:      seqNo,
		Hash:       hash,
		Timestamp:  timestamp,
		CommitTime: time.Now().UnixNano(),
		Flag:       flag,
		IsPrimary:  isPrimary,
	}

	// Post the event to outer
	// !!! CANNOT use go, it will result in concurrent problems when writing blocks
	h.msgQ.Post(writeEvent)

	return nil
}

// UpdateState transfers the UpdateStateEvent to outer
func (h *helper) UpdateState(updateState *pb.UpdateStateMessage) error {
	tmpMsg, err := proto.Marshal(updateState)

	if err != nil {
		return err
	}

	updateStateEvent := event.SendCheckpointSyncEvent{
		Payload: tmpMsg,
	}

	// Post the event to outer
	go h.msgQ.Post(updateStateEvent)

	return nil
}

// UpdateState transfers the UpdateStateEvent to outer
func (h *helper) ValidateBatch(txs []*types.Transaction, timeStamp int64, seqNo uint64, view uint64, isPrimary bool) error {

	validateEvent := event.ExeTxsEvent{
		Transactions: txs,
		Timestamp:    timeStamp,
		SeqNo:        seqNo,
		View:         view,
		IsPrimary:    isPrimary,
	}

	// Post the event to outer
	go h.msgQ.Post(validateEvent)

	return nil
}

// VcReset reset vid when view change is done
func (h *helper) VcReset(seqNo uint64) error {

	vcResetEvent := event.VCResetEvent{
		SeqNo: seqNo,
	}

	// No need to "go h.msgQ.Post...", we'll wait for it to return
	h.msgQ.Post(vcResetEvent)
	//time.Sleep(time.Millisecond * 50)

	return nil
}

// Inform the primary id after negotiate or
func (h *helper) InformPrimary(primary uint64) error {

	informPrimaryEvent := event.InformPrimaryEvent{
		Primary: primary,
	}

	go h.msgQ.Post(informPrimaryEvent)

	return nil
}

// Broadcast addnode message to others
func (h *helper) BroadcastAddNode(msg *pb.Message) error {

	tmpMsg, err := proto.Marshal(msg)

	if err != nil {
		return err
	}

	broadcastEvent := event.BroadcastNewPeerEvent{
		Payload: tmpMsg,
	}

	// Post the event to outer
	go h.msgQ.Post(broadcastEvent)

	return nil
}

// Broadcast delnode message to others
func (h *helper) BroadcastDelNode(msg *pb.Message) error {

	tmpMsg, err := proto.Marshal(msg)

	if err != nil {
		return err
	}

	broadcastEvent := event.BroadcastDelPeerEvent{
		Payload: tmpMsg,
	}

	// Post the event to outer
	go h.msgQ.Post(broadcastEvent)

	return nil
}

// Inform to update routing table
func (h *helper) UpdateTable(payload []byte, flag bool) error {

	updateTable := event.UpdateRoutingTableEvent{
		Payload: payload,
		Type:    flag,
	}

	h.msgQ.Post(updateTable)

	return nil
}

// Inform to update routing table
func (h *helper) ClearValidateCache() error {

	remove := event.RemoveCacheEvent{}
	h.msgQ.Post(remove)

	return nil
}

// NewHelper initializes a helper object
func NewHelper(m *event.TypeMux) *helper {

	h := &helper{
		msgQ: m,
	}

	return h
}
