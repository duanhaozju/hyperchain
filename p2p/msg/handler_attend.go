package msg

import "fmt"
import (
	pb "hyperchain/p2p/message"
	"hyperchain/manager/event"
	"hyperchain/p2p/hts"
	"github.com/pkg/errors"
)

type AttendMsgHandler struct {
	mchan chan interface{}
	ev    *event.TypeMux
	shts  *hts.ServerHTS
}

func NewAttendHandler(blackHole chan interface{}, ev *event.TypeMux, shts *hts.ServerHTS) *AttendMsgHandler {
	return &AttendMsgHandler{
		mchan:blackHole,
		ev:ev,
		shts:shts,
	}
}

func (h  *AttendMsgHandler) Process() {
	for msg := range h.mchan {
		fmt.Println("got a Attend message", string(msg.(*pb.Message).Payload))
	}
}

func (h  *AttendMsgHandler) Teardown() {
	//TODO THIS is UN Allowed, because reciver cannot close the mchan
	close(h.mchan)
}

func (h *AttendMsgHandler)Receive() chan <- interface{} {
	return h.mchan
}

func (h *AttendMsgHandler)Execute(msg *pb.Message) (*pb.Message, error) {
	if msg == nil || msg.Payload == nil {
		return nil, errors.New("msg or msg body is nil")
	}
	//decrypt
	payload := h.shts.Decrypt(string(msg.From.UUID), msg.Payload)
	if payload == nil {
		return nil, errors.New("cannot decrypt the msg")
	}
	go h.ev.Post(event.NewPeerEvent{
		Payload:payload,
	})
	rsp := &pb.Message{
		MessageType:pb.MsgType_RESPONSE,
	}
	return rsp, nil
}
