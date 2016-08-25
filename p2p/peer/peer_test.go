package client

import (
	"testing"
	"hyperchain-alpha/p2p/peermessage"
	"log"
	//"time"
	"fmt"
	//"hyperchain-alpha/peer/Server"
	"time"
	node "hyperchain-alpha/p2p/node"
)


func TestNewChatClient(t *testing.T) {
	//start the server
	server := node.NewNode(8002)
	go func() {
		for now := range time.Tick(3 * time.Second) {
			fmt.Println(now)
		}
	}()


	chatClient,err := NewPeer("localhost:8002")
	if err != nil{
		log.Fatalln("连接失败")
		server.StopServer()
	}

	msg,err2 := chatClient.Chat(&peermessage.Message{
		MessageType:peermessage.Message_HELLO,
		Payload:[]byte("Hello"),
	})

	if err2 != nil{
		log.Fatalln("发送消息失败")
		server.StopServer()
	}else{
		fmt.Println(msg)
		server.StopServer()
	}
}

func TestChatClient_Close(t *testing.T) {
	//start the server
	server := node.NewNode(8001)
	go func() {
		for now := range time.Tick(3 * time.Second) {
			fmt.Println(now)
		}
	}()


	chatClient,err := NewPeer("localhost:8001")
	if err != nil{
		log.Fatalln("连接失败")
		server.StopServer()
	}

	msg,err2 := chatClient.Chat(&peermessage.Message{
		MessageType:peermessage.Message_HELLO,
		Payload:[]byte("Hello"),
	})

	if err2 != nil{
		log.Fatalln("发送消息失败")
		server.StopServer()
	}else{
		fmt.Println(msg)
		server.StopServer()
		chatClient.Close()
	}
}
