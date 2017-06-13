package network

import (
	"testing"
	"hyperchain/p2p/message"
	"fmt"
	"hyperchain/p2p/msg"
)

func TestServer_Chat(t *testing.T) {
	s := &Server{}
	helloChan := make(chan *message.Message,100000)
	keepaliveChan := make(chan *message.Message,100000)
	helloHandler := msg.NewHelloHandler(helloChan)
	keepaliveHandler := msg.NewKeepAliveHandler(keepaliveChan)
	s.RegisterSlot(message.Message_HELLO,helloHandler)
	s.RegisterSlot(message.Message_KEEPALIVE,keepaliveHandler)
	go s.StartServer(50012)

	closeChan := make(chan struct{})

	go func(){
		for msg := range helloChan{
			fmt.Println("hello" + string(msg.Payload))
		}
	}()

	go func(){
		for msg := range keepaliveChan{
			fmt.Println("keepalive" + string(msg.Payload))
		}
	}()
	<-closeChan
}
