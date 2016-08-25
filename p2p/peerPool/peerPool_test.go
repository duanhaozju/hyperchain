package peerPool

import (
	"testing"
	"hyperchain-alpha/p2p/peermessage"
	"fmt"
	node "hyperchain-alpha/p2p/node"
	"log"
	peer "hyperchain-alpha/p2p/peer"
	"strconv"
	/*"github.com/ethereum/go-ethereum/node"
	"google.golang.org/grpc/peer"*/
)

func TestPeersPool_PutPeer(t *testing.T) {
	portRange := 8002
	//get the client
	//start the server
	server := node.NewChatServer(int32(portRange))

	chatClient, err := peer.NewChatClient("localhost:"+ strconv.Itoa(portRange))
	if err != nil {
		log.Fatalln("连接失败")
	}

	msg, err2 := chatClient.Chat(&peermessage.Message{
		MessageType:peermessage.Message_HELLO,
		Payload:[]byte("Hello"),
	})

	if isClosed,err := chatClient.Close(); isClosed{
		server.StopServer()
	}else{
		if err != nil {
			t.Errorf("关闭client错误", err)
		}
		server.StopServer()
	}

	if err2 != nil {
		log.Fatalln("发送消息失败")
	} else {
		fmt.Println(msg)
		peerPool := NewPeerPool(true)
		//here test the peer pool
		_,err := peerPool.PutPeer(chatClient.Addr, chatClient)
		if err == nil {
			//测试取出
			t.Log("放入成功")
		} else {
			t.Errorf("放入缓存池错误..., %v", err)
		}
	}

	//if isClosed,err := chatClient.Close(); isClosed{
	//	server.StopServer()
	//}else{
	//	if err != nil {
	//		t.Errorf("关闭client错误", err)
	//	}
	//	server.StopServer()
	//}
}

func TestPeersPool_GetPeer(t *testing.T) {
	portRange := 8001
	//get the client
	//start the server
	server := node.NewChatServer(int32(portRange))

	chatClient, err := peer.NewChatClient("localhost:"+ strconv.Itoa(portRange))
	if err != nil {
		log.Fatalln("连接失败")
	}

	msg, err2 := chatClient.Chat(&peermessage.Message{
		MessageType:peermessage.Message_HELLO,
		Payload:[]byte("Hello"),
	})

	if err2 != nil {
		log.Fatalln("发送消息失败")
	} else {
		fmt.Println(msg)
		peerPool := NewPeerPool(true)
		//here test the peer pool
		_,err := peerPool.PutPeer(chatClient.Addr, chatClient)
		if err == nil {
			//测试取出
			pAddr := peermessage.PeerAddress{
				Ip:"localhost",
				Port:int32(8001),
			}
			client := peerPool.GetPeer(pAddr)
			retMessage, err := client.Chat(&peermessage.Message{
				MessageType:peermessage.Message_HELLO,
				Payload:[]byte("Hello"),
			})
			if err != nil {
				t.Errorf("从缓存池取出错误 %v", err)
			}
			if retMessage.MessageType == peermessage.Message_RESPONSE {
				t.Log("取出成功")
			} else {
				t.Error("取出的client无法通信")

			}
		} else {
			t.Errorf("放入缓存池错误..., %v", err)

		}

	}

	if isClosed,err := chatClient.Close(); isClosed{
		server.StopServer()
	}else{
		if err != nil {
			t.Errorf("关闭client错误", err)
		}
		server.StopServer()
	}
}

