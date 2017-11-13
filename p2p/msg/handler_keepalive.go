package msg

import "fmt"
import (
	pb "github.com/hyperchain/hyperchain/p2p/message"
)

type KeepaliveMsgHandler struct {
	msgChan chan *pb.Message
}

func NewKeepAliveHandler(blackHole <-chan *pb.Message) *KeepaliveMsgHandler {
	return &KeepaliveMsgHandler{
		msgChan: make(chan *pb.Message, 100000),
	}
}

func (keepAliveHandler *KeepaliveMsgHandler) Process() {
	for msg := range keepAliveHandler.msgChan {
		fmt.Println("got a keepalive message", string(msg.Payload))
	}
}

func (keepAliveHandler *KeepaliveMsgHandler) Teardown() {
	close(keepAliveHandler.msgChan)
}

func (keepAliveHandler *KeepaliveMsgHandler) Receive() <-chan *pb.Message {
	return keepAliveHandler.msgChan
}

func (keepaliveHandler *KeepaliveMsgHandler) Execute(msg *pb.Message) (*pb.Message, error) {
	return msg, nil
}
