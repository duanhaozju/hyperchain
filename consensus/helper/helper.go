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

func (h *helper) InnerBroadcast(msg *pb.Message) error{
	tmpMsg, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	broadcastEvent := &event.BroadcastConsensusEvent{
		Payload: tmpMsg,
	}
	h.msgQ.Post(broadcastEvent)
	return nil
}

func (h *helper) Execute(reqBatch *pb.ExeMessage) error{
	tmpMsg,err:=proto.Marshal(reqBatch)
	if err!=nil {
		return err
	}
	exeEvent:=&event.NewBlockEvent{
		Payload:tmpMsg,
	}
	h.msgQ.Post(exeEvent)
	return nil
}

func NewHelper(m *event.TypeMux) *helper {
	h:=&helper{
		msgQ:m,
	}
	return h
}