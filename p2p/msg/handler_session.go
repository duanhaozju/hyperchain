package msg

import "fmt"
import (
	pb "hyperchain/p2p/message"
	"hyperchain/manager/event"
)


type SessionMsgHandler struct {
	mchan chan interface{}
	evmux *event.TypeMux
}

func NewSessionHandler(blackHole chan interface{},eventHub *event.TypeMux)*SessionMsgHandler{
	return &SessionMsgHandler{
		mchan:blackHole,
		evmux:eventHub,
	}
}

func (session  *SessionMsgHandler) Process() {
	for msg := range session.mchan {
		 fmt.Println("got a hello message", string(msg.(pb.Message).Payload))
		}
}

func (session  *SessionMsgHandler)  Teardown() {
	close(session.mchan)
}

func (session  *SessionMsgHandler) Receive() chan<- interface{}{
	return session.mchan
}

func (session  *SessionMsgHandler) Execute(msg *pb.Message) (*pb.Message,error){
	go session.evmux.Post(event.SessionEvent{
		Message:msg.Payload,
	})
	rsp  := &pb.Message{
		MessageType:pb.MsgType_RESPONSE,
	}
	return rsp,nil
}
