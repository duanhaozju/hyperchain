package msg

import (
	pb "hyperchain/p2p/message"
	"hyperchain/manager/event"
	"fmt"
)

type NVPAttendMsgHandler struct {
	mchan chan  interface{}
	ev *event.TypeMux
}

func NewNVPAttendHandler(blackHole chan interface{},ev *event.TypeMux)*NVPAttendMsgHandler{
	return &NVPAttendMsgHandler{
		mchan:blackHole,
		ev:ev,
	}
}

//Process
func (h  *NVPAttendMsgHandler) Process() {
	for msg := range h.mchan {
		 fmt.Println("got a Attend message", string(msg.(*pb.Message).Payload))
	}
}

//Teardown
func (h  *NVPAttendMsgHandler) Teardown() {
	//TODO THIS is UN Allowed, because reciver cannot close the mchan
	close(h.mchan)
}

//Receive
func (h *NVPAttendMsgHandler)Receive() chan<- interface{}{
	return h.mchan
}

//Execute
func (h *NVPAttendMsgHandler)Execute(msg *pb.Message) (*pb.Message,error){
	fmt.Printf("GOT A NVP ATTEND MSG hostname(%s), type: %s ",msg.From.Hostname,msg.MessageType)
	rsp  := &pb.Message{
		MessageType:pb.MsgType_RESPONSE,
	}
	return rsp,nil
}
