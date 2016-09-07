package helper

import (
	"hyperchain/event"
	pb "hyperchain/protos"

	"github.com/golang/protobuf/proto"
)

type helper struct {
	msgQ *event.TypeMux
}

type Stack interface {
	InnerBroadcast(msg *pb.Message) error
	Execute(reqBatch *pb.ExeMessage) error
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

// Execute transfers the transactions decided by consensus to outer
func (h *helper) Execute(reqBatch *pb.ExeMessage) error{

	tmpMsg,err:=proto.Marshal(reqBatch)

	if err!=nil {
		return err
	}

	exeEvent := event.NewBlockEvent{
		Payload:	tmpMsg,

	}

	// Post the event to outer
	go h.msgQ.Post(exeEvent)

	return nil
}

// NewHelper initializes a helper object
func NewHelper(m *event.TypeMux) *helper {

	h:=&helper{
		msgQ: m,
	}

	return h
}