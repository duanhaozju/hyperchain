// author: chenquan
// date: 16-8-25
// last modified: 16-8-25 20:01
// last Modified Author: chenquan
// change log:
//
package peerPool

import (
	"testing"
	"hyperchain-alpha/p2p/peermessage"
	"fmt"

	"log"

	"strconv"

	node "hyperchain-alpha/p2p/node"
	peer "hyperchain-alpha/p2p/peer"

	"time"
)

func TestPeersPool_PutPeer(t *testing.T) {
	portRange := 8002
	//get the client
	//start the server
	server := node.NewNode(portRange)

	chatClient, err := peer.NewPeer("localhost:"+ strconv.Itoa(portRange))
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
			t.Log("放入成功")
		} else {
			t.Errorf("放入缓存池错误..., %v", err)
		}
	}
	tickCount := 0
	for tick := range time.Tick(3 *time.Second){
		tickCount += 1
		log.Println(tick)
		if tickCount >5{
			break
		}
	}
	server.StopServer()
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
	server := node.NewNode(int(portRange))

	chatClient, err := peer.NewPeer("localhost:"+ strconv.Itoa(portRange))
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

