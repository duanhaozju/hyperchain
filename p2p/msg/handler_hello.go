package msg

import "fmt"
import pb "hyperchain/p2p/message"

type HelloMsgHandler struct {
	mchan chan *pb.Message
}

func NewHelloHandler(blackHole chan *pb.Message)*HelloMsgHandler{
	return &HelloMsgHandler{
		mchan:blackHole,
	}
}

func (hellloMsgh  *HelloMsgHandler) Process() {
	for msg := range hellloMsgh.mchan {
		 fmt.Println("got a hello message", string(msg.Payload))
		}
}

func (hellloMsgh  *HelloMsgHandler) Teardown() {
	//TODO THIS is UN Allowed, because reciver cannot close the mchan
	close(hellloMsgh.mchan)
}

func (helloMsgh *HelloMsgHandler)Receive() chan<- *pb.Message{
	return helloMsgh.mchan
}

func (helloMsgh *HelloMsgHandler)Execute(msg *pb.Message) (*pb.Message,error){
	rsp  := &pb.Message{
		MessageType:pb.MsgType_HELLO,
	}
	return rsp,nil
}
