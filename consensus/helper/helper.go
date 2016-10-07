package helper

import (
	"time"

	"hyperchain/event"
	pb "hyperchain/protos"

	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
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
	Execute(reqBatch *pb.ExeMessage) error
	UpdateState(updateState *pb.UpdateStateMessage) error
}

// InnerBroadcast broadcast the consensus message between vp nodes
func (h *helper) InnerBroadcast(msg *pb.Message) error{

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
func (h *helper) InnerUnicast(msg *pb.Message, to uint64) error{

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
func (h *helper) Execute(reqBatch *pb.ExeMessage) error{

	tmpMsg,err:=proto.Marshal(reqBatch)

	if err!=nil {
		return err
	}

	exeEvent := event.NewBlockEvent{
		Payload:	tmpMsg,
		CommitTime:	time.Now().UnixNano(),
	}

	// Post the event to outer
	go h.msgQ.Post(exeEvent)

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
	logger.Error("-------------post UpdateStateEvent----------")
	go h.msgQ.Post(updateStateEvent)

	return nil
}

// NewHelper initializes a helper object
func NewHelper(m *event.TypeMux) *helper {

	h:=&helper{
		msgQ: m,
	}

	return h
}