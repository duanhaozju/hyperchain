// author: chenquan
// date: 16-8-25
// last modified: 16-8-29 13:23
// last Modified Author: chenquan
// change log:
//
package peerPool

import (
	"testing"
	"hyperchain/p2p/peermessage"
	"log"
	"strconv"
	node "hyperchain/p2p/node"
	peer "hyperchain/p2p/peer"
	"time"
	"hyperchain/event"
	)


// test the peer pool to put peer
func TestPeersPool_PutPeer(t *testing.T) {
	portRange := 8015
	//get the client
	//start the server
	eventMux := new(event.TypeMux)
	server := node.NewNode(portRange,true,eventMux)

	chatClient, err := peer.NewPeerByString("localhost:"+ strconv.Itoa(portRange))
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
		log.Println(msg)
		peerPool := NewPeerPool(true,false)
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
	for tick := range time.Tick(1 *time.Second){
		tickCount += 1
		log.Println(tick)
		if tickCount >2{
			break
		}
	}


	if isClosed,err := chatClient.Close(); isClosed{
		t.Log("关闭client成功")
		server.StopServer()
	}else{
		t.Errorf("关闭client错误", err)
	}

}

func TestPeersPool_GetPeer(t *testing.T) {
	portRange := 8016
	//get the client
	//start the server
	eventMux := new(event.TypeMux)
	server := node.NewNode(portRange,true,eventMux)

	chatClient, err := peer.NewPeerByString("localhost:"+ strconv.Itoa(portRange))
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
		log.Println(msg)
		peerPool := NewPeerPool(true,false)
		//here test the peer pool
		_,err := peerPool.PutPeer(chatClient.Addr, chatClient)
		if err == nil {
			//测试取出
			pAddr := peermessage.PeerAddress{
				Ip:"localhost",
				Port:int32(portRange),
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
		t.Log("关闭client成功")
		server.StopServer()
	}else{
		t.Errorf("关闭client错误", err)
	}

}

