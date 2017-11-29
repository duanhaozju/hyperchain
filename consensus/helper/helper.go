//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package helper

import (
	"time"

	"github.com/hyperchain/hyperchain/consensus"
	"github.com/hyperchain/hyperchain/core/types"
	"github.com/hyperchain/hyperchain/manager/appstat"
	"github.com/hyperchain/hyperchain/manager/event"
	pb "github.com/hyperchain/hyperchain/manager/protos"

	"github.com/golang/protobuf/proto"
	"github.com/hyperchain/hyperchain/core/oplog"
	opLog2 "github.com/hyperchain/hyperchain/core/oplog/proto"
	"github.com/hyperchain/hyperchain/core/fiber"
)

type helper struct {
	opLog       oplog.OpLog
	fiber       fiber.Fiber
	innerMux    *event.TypeMux
	externalMux *event.TypeMux
}

// Stack helps rbftImpl send message to other components of system. RbftImpl
// would generate some messages and post these messages to other components,
// in order to send messages to other vp nodes, let other components validate or
// execute transactions, or send messages to clients.
type Stack interface {
	// InnerBroadcast broadcast the consensus message to all other vp nodes
	InnerBroadcast(msg *pb.Message) error

	// InnerUnicast unicast the transaction message to a specific vp node
	InnerUnicast(msg *pb.Message, to uint64) error

	// Execute transfers the transactions decided by consensus to outer to execute these transactions
	Execute(seqNo uint64, hash string, flag bool, isPrimary bool, time int64) error

	// UpdateState transfers the UpdateStateEvent to outer
	UpdateState(myId uint64, height uint64, blockHash []byte, replicas []event.SyncReplica) error

	// ValidateBatch transfers the ValidationEvent to outer
	ValidateBatch(digest string, txs []*types.Transaction, timeStamp int64, seqNo uint64, view uint64, isPrimary bool) error

	// VcReset reset vid when view change is done, clear the validate cache larger than seqNo
	VcReset(seqNo uint64, view uint64) error

	// InformPrimary send the primary id to update info after negotiate view or view change
	InformPrimary(primary uint64) error

	// BroadcastAddNode broadcast addnode message to others
	BroadcastAddNode(msg *pb.Message) error

	// BroadcastDelNode broadcast delnode message to others
	BroadcastDelNode(msg *pb.Message) error

	// UpdateTable inform to update routing table
	UpdateTable(payload []byte, flag bool) error

	// SendFilterEvent sends event to subscription system, then the system would return message to clients which subscribe this message.
	SendFilterEvent(informType int, message ...interface{}) error

	// GetLatestCommitNumber queries and returns latest committed block number from opLog
	GetLatestCommitNumber() uint64

	// GetLatestCommitHeightAndHash queries and returns latest committed block number and hash from opLog
	GetLatestCommitHeightAndHash() (uint64, string, error)

	// StableCheckpoint sends stable checkpoint ack to executor
	StableCheckpoint(isStable bool, seqNo uint64)
}

// NewHelper initializes a helper object
func NewHelper(innerMux *event.TypeMux, externalMux *event.TypeMux, opLog oplog.OpLog, fiber fiber.Fiber) *helper {

	h := &helper{
		innerMux:    innerMux,
		externalMux: externalMux,
		opLog:       opLog,
		fiber:       fiber,
	}
	return h
}

// InnerBroadcast broadcasts the consensus message between VP nodes
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

// InnerUnicast unicasts message to the specified VP node
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

// ValidateBatch transfers the ValidateEvent to outer
func (h *helper) ValidateBatch(digest string, txs []*types.Transaction, timeStamp int64, seqNo uint64, view uint64, isPrimary bool) error {

	validationEvent := &event.ValidationEvent{
		Digest:       digest,
		Transactions: txs,
		SeqNo:        seqNo,
		View:         view,
		IsPrimary:    isPrimary,
		Timestamp:    timeStamp,
	}

	payload, err := proto.Marshal(validationEvent)
	if err != nil {
		return err
	}

	entry := &opLog2.LogEntry{
		Type:    opLog2.LogEntry_TransactionList,
		Payload: payload,
	}

	if err = h.opLog.Append(entry); err != nil {
		return err
	}

	return nil
}

// VcReset resets vid after in recovery, viewchange or add/delete nodes
func (h *helper) VcReset(seqNo uint64, view uint64) error {

	vcResetEvent := event.VCResetEvent{
		SeqNo: seqNo,
		View:  view,
	}

	// No need to "go h.msgQ.Post...", we'll wait for it to return
	h.innerMux.Post(vcResetEvent)

	return nil
}

// InformPrimary informs the primary id after negotiate or viewchanged
func (h *helper) InformPrimary(primary uint64) error {

	informPrimaryEvent := event.InformPrimaryEvent{
		Primary: primary,
	}

	go h.innerMux.Post(informPrimaryEvent)

	return nil
}

// BroadcastAddNode broadcasts addnode message to others
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

// BroadcastDelNode broadcasts delnode message to others
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

// UpdateTable informs to update routing table
func (h *helper) UpdateTable(payload []byte, flag bool) error {

	updateTable := event.UpdateRoutingTableEvent{
		Payload: payload,
		Type:    flag,
	}

	h.innerMux.Post(updateTable)

	return nil
}

// PostExternal posts event to outer event mux
func (h *helper) PostExternal(ev interface{}) {
	h.externalMux.Post(ev)
}

// sendFilterEvent sends event to subscription system.
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

func (h *helper) GetLatestCommitNumber() uint64 {
	return h.opLog.GetLastBlockNum()
}

func (h *helper) GetLatestCommitHeightAndHash() (uint64, string, error) {
	return h.opLog.GetHeightAndDigest()
}

func (h *helper) StableCheckpoint(isStable bool, seqNo uint64) {
	if isStable {
		h.opLog.SetStableCheckpoint(seqNo)
	} else {
		// TODO
	}

	ack := event.CheckpointAck{
		IsStableCkpt: isStable,
		Cid:          seqNo,
	}
	h.fiber.Send(ack)
}
