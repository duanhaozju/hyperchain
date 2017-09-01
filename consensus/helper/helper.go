//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package helper

import (
	"time"

	"hyperchain/consensus"
	"hyperchain/core/types"
	"hyperchain/manager/appstat"
	"hyperchain/manager/event"
	pb "hyperchain/manager/protos"

	"github.com/golang/protobuf/proto"
)

type helper struct {
	innerMux    *event.TypeMux
	externalMux *event.TypeMux
}

type Stack interface {
	InnerBroadcast(msg *pb.Message) error
	InnerUnicast(msg *pb.Message, to uint64) error
	Execute(seqNo uint64, hash string, flag bool, isPrimary bool, time int64) error
	UpdateState(myId uint64, height uint64, blockHash []byte, replicas []event.SyncReplica) error
	ValidateBatch(digest string, txs []*types.Transaction, timeStamp int64, seqNo uint64, vid uint64, view uint64, isPrimary bool) error
	VcReset(seqNo uint64) error
	InformPrimary(primary uint64) error
	BroadcastAddNode(msg *pb.Message) error
	BroadcastDelNode(msg *pb.Message) error
	UpdateTable(payload []byte, flag bool) error
	SendFilterEvent(informType int, message ...interface{}) error
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
	go h.innerMux.Post(broadcastEvent)

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
	go h.innerMux.Post(unicastEvent)

	return nil
}

// Execute transfers the transactions decided by consensus to outer
func (h *helper) Execute(seqNo uint64, hash string, flag bool, isPrimary bool, timestamp int64) error {

	writeEvent := event.CommitEvent{
		SeqNo:      seqNo,
		Hash:       hash,
		Timestamp:  timestamp,
		CommitTime: time.Now().UnixNano(),
		Flag:       flag,
		IsPrimary:  isPrimary,
	}

	// Post the event to outer
	// !!! CANNOT use go, it will result in concurrent problems when writing blocks
	h.innerMux.Post(writeEvent)

	return nil
}

// UpdateState transfers the UpdateStateEvent to outer
func (h *helper) UpdateState(myId uint64, height uint64, blockHash []byte, replicas []event.SyncReplica) error {
	updateStateEvent := event.ChainSyncReqEvent{
		Id:              myId,
		TargetHeight:    height,
		TargetBlockHash: blockHash,
		Replicas:        replicas,
	}

	// Post the event to outer
	go h.innerMux.Post(updateStateEvent)

	return nil
}

// UpdateState transfers the UpdateStateEvent to outer
func (h *helper) ValidateBatch(digest string, txs []*types.Transaction, timeStamp int64, seqNo uint64, vid uint64, view uint64, isPrimary bool) error {

	validateEvent := event.ValidationEvent{
		Digest:       digest,
		Vid:          vid,
		Transactions: txs,
		Timestamp:    timeStamp,
		SeqNo:        seqNo,
		View:         view,
		IsPrimary:    isPrimary,
	}

	// Post the event to outer
	h.innerMux.Post(validateEvent)

	return nil
}

// VcReset reset vid when view change is done
func (h *helper) VcReset(seqNo uint64) error {

	vcResetEvent := event.VCResetEvent{
		SeqNo: seqNo,
	}

	// No need to "go h.msgQ.Post...", we'll wait for it to return
	h.innerMux.Post(vcResetEvent)
	//time.Sleep(time.Millisecond * 50)

	return nil
}

// Inform the primary id after negotiate or
func (h *helper) InformPrimary(primary uint64) error {

	informPrimaryEvent := event.InformPrimaryEvent{
		Primary: primary,
	}

	go h.innerMux.Post(informPrimaryEvent)

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
	h.innerMux.Post(broadcastEvent)

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
	h.innerMux.Post(broadcastEvent)

	return nil
}

// Inform to update routing table
func (h *helper) UpdateTable(payload []byte, flag bool) error {

	updateTable := event.UpdateRoutingTableEvent{
		Payload: payload,
		Type:    flag,
	}

	h.innerMux.Post(updateTable)

	return nil
}

// NewHelper initializes a helper object
func NewHelper(innerMux *event.TypeMux, externalMux *event.TypeMux) *helper {

	h := &helper{
		innerMux:    innerMux,
		externalMux: externalMux,
	}

	return h
}

// PostExternal post event to outer event mux
func (h *helper) PostExternal(ev interface{}) {
	h.externalMux.Post(ev)
}

// sendFilterEvent - send event to subscription system.
func (h *helper) SendFilterEvent(informType int, message ...interface{}) error {
	switch informType {
	case consensus.FILTER_View_Change_Finish:
		// NewBlock event
		if len(message) != 1 {
			return nil
		}
		msg, ok := message[0].(string)
		if ok == false {
			return nil
		}
		h.PostExternal(event.FilterSystemStatusEvent{
			Module:  appstat.ExceptionModule_Consenus,
			Status:  appstat.Normal,
			Subtype: appstat.ExceptionSubType_ViewChange,
			Message: msg,
		})
		return nil
	default:
		return nil
	}
}
