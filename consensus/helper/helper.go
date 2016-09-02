package helper

import (
	"hyperchain/event"
	pb "hyperchain/protos"

	"github.com/golang/protobuf/proto"
	"fmt"
	"hyperchain/manager"
)
type helper struct {
	msgQ *event.TypeMux
}

type Stack interface {
	InnerBroadcast(msg *pb.Message) error
	Execute(reqBatch *pb.ExeMessage, now uint64, pre uint64) error
}

func (h *helper) InnerBroadcast(msg *pb.Message) error{
	fmt.Println("enter innerbroad cast#######")
	tmpMsg, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	broadcastEvent := event.BroadcastConsensusEvent{
		Payload: tmpMsg,
	}
	//fmt.Println("broadcast")
	//manager.GetEventObject().Post(event.NewTxEvent{Payload: []byte{0x00, 0x00, 0x03, 0xe8}})

	go manager.GetEventObject().Post(broadcastEvent)
	//h.msgQ.Post(broadcastEvent)
	return nil
}

func (h *helper) Execute(reqBatch *pb.ExeMessage, now uint64, pre uint64) error{
	tmpMsg,err:=proto.Marshal(reqBatch)
	if err!=nil {
		return err
	}
	exeEvent := event.NewBlockEvent{
		Payload:	tmpMsg,
		Now:		now,
		Pre:		pre,
	}
	go manager.GetEventObject().Post(exeEvent)
	//h.msgQ.Post(exeEvent)
	return nil
}

func NewHelper(m *event.TypeMux) *helper {
	h:=&helper{
		msgQ:m,
	}
	return h
}