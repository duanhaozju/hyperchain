package msg

import "fmt"
import pb "hyperchain/p2p/message"

type HelloMsgHandler struct {
	msgChan chan *pb.Message
}

func NewHelloHandler(blackHole <- chan pb.Message)*HelloMsgHandler{
	return HelloMsgHandler{
		msgChan:make(chan *pb.Message, 100000),
	}
}

func (hellloMsgh  *HelloMsgHandler) Process() {
	for msg := range hellloMsgh.msgChan {
		 fmt.Println("got a hello message", string(msg.Payload))
		}
}

func (hellloMsgh  *HelloMsgHandler) Teardown() {
	close(hellloMsgh.msgChan)
}

func (helloMsgh *HelloMsgHandler)Recive() <- chan *pb.Message{
	return <- helloMsgh.msgChan
}

func (helloMsgh *HelloMsgHandler)Execute(msg *pb.Message) (*pb.Message,error){
	return msg,nil
}
