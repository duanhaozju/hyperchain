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
	InnerBroadcast(e *pb.Message) error
	Execute(batch *pb.ExeMessage) error
}

func (h *helper) InnerBroadcast(e *pb.Message) error{
	tmpMsg,err:=proto.Marshal(e)
	if err!=nil {
		return err
	}
	wrapMessage:=&event.BroadcastConsensusEvent{
		Payload:tmpMsg,
	}
	h.msgQ.Post(wrapMessage)
	return nil
}


func (h *helper) Execute(reqBatch *pb.ExeMessage) error{
	tmpMsg,err:=proto.Marshal(reqBatch)
	if err!=nil {
		return err
	}
	wrapMessage:=&event.NewBlockEvent{
		Payload:tmpMsg,
	}
	h.msgQ.Post(wrapMessage)
	return nil
}

func NewHelper(m *event.TypeMux) (*helper){
	h:=&helper{
		msgQ:m,
	}
	return h
}