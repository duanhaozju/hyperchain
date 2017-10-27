package msg

import "fmt"
import (
	"github.com/hyperchain/hyperchain/manager/event"
	pb "github.com/hyperchain/hyperchain/p2p/message"
)

type HelloMsgHandler struct {
	mchan chan interface{}
	ev    *event.TypeMux
}

func NewHelloHandler(blackHole chan interface{}, ev *event.TypeMux) *HelloMsgHandler {
	return &HelloMsgHandler{
		mchan: blackHole,
		ev:    ev,
	}
}

func (hellloMsgh *HelloMsgHandler) Process() {
	for msg := range hellloMsgh.mchan {
		fmt.Println("got a hello message", string(msg.(*pb.Message).Payload))
	}
}

func (hellloMsgh *HelloMsgHandler) Teardown() {
	//TODO THIS is UN Allowed, because reciver cannot close the mchan
	close(hellloMsgh.mchan)
}

func (helloMsgh *HelloMsgHandler) Receive() chan<- interface{} {
	return helloMsgh.mchan
}

func (helloMsgh *HelloMsgHandler) Execute(msg *pb.Message) (*pb.Message, error) {
	fmt.Printf("Got a hello message %+v \n", msg)
	rsp := &pb.Message{
		MessageType: pb.MsgType_RESPONSE,
	}
	return rsp, nil
}
