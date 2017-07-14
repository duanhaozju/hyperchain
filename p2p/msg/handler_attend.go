package msg

import "fmt"
import (
	pb "hyperchain/p2p/message"
	"hyperchain/manager/event"
)

type AttendMsgHandler struct {
	mchan chan  interface{}
	ev *event.TypeMux
}

func NewAttendHandler(blackHole chan interface{},ev *event.TypeMux)*AttendMsgHandler{
	return &AttendMsgHandler{
		mchan:blackHole,
		ev:ev,
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

func (h *AttendMsgHandler)Receive() chan<- interface{}{
	return h.mchan
}

func (h *AttendMsgHandler)Execute(msg *pb.Message) (*pb.Message,error){
	fmt.Println("got a new peer event ATTEND Msg")
	go h.ev.Post(event.NewPeerEvent{
		Payload:msg.Payload,
	})
	rsp  := &pb.Message{
		MessageType:pb.MsgType_RESPONSE,
	}
	return rsp,nil
}
