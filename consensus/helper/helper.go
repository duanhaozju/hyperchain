package helper

import (
	"time"

	"hyperchain/event"
	pb "hyperchain/protos"

	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"hyperchain/core/types"
)

var logger *logging.Logger // package-level logger

func init() {
	logger = logging.MustGetLogger("consensus/help")
}

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
		Payload:	tmpMsg,
		PeerId:		to,
	}

	// Post the event to outer
	go h.msgQ.Post(unicastEvent)

	return nil
}

// Execute transfers the transactions decided by consensus to outer
func (h *helper) Execute(seqNo uint64, hash string, flag bool, isPrimary bool, timestamp int64) error {

	writeEvent := event.CommitOrRollbackBlockEvent {
		SeqNo:		seqNo,
		Hash:		hash,
		Timestamp:	timestamp,
		CommitTime:	time.Now().UnixNano(),
		Flag:		flag,
		IsPrimary:	isPrimary,
	}

	// Post the event to outer
	h.msgQ.Post(writeEvent)

	return nil
}

// UpdateState transfers the UpdateStateEvent to outer
func (h *helper) UpdateState(updateState *pb.UpdateStateMessage) error {

	tmpMsg, err := proto.Marshal(updateState)

	if err != nil {
		return err
	}

	updateStateEvent := event.SendCheckpointSyncEvent {
		Payload:	tmpMsg,
	}

	// Post the event to outer
	go h.msgQ.Post(updateStateEvent)

	return nil
}

// UpdateState transfers the UpdateStateEvent to outer
func (h *helper) ValidateBatch(txs []*types.Transaction, timeStamp int64, seqNo uint64, view uint64, isPrimary bool) error {

	validateEvent := event.ExeTxsEvent {
		Transactions:	txs,
		Timestamp:      timeStamp,
		SeqNo:		seqNo,
		View:		view,
		IsPrimary:	isPrimary,
	}

	// Post the event to outer
	h.msgQ.Post(validateEvent)

	return nil
}

// VcReset reset vid when view change is done
func (h *helper) VcReset(seqNo uint64) error {
	vcResetEvent := event.VCResetEvent{
		SeqNo: seqNo,
	}

	// No need to "go h.msgQ.Post...", we'll wait for it to return
	h.msgQ.Post(vcResetEvent)

	time.Sleep(10 * time.Millisecond)

	return nil
}

// NewHelper initializes a helper object
func NewHelper(m *event.TypeMux) *helper {

	h := &helper{
		msgQ: m,
	}

	return h
}