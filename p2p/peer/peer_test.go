// author: chenquan
// date: 16-8-25
// last modified: 16-8-29 13:58
// last Modified Author: chenquan
// change log: add a comment of this file function
//

package client

import (
	"testing"
	"hyperchain/p2p/peermessage"
	node "hyperchain/p2p/node"
	"hyperchain/event"
)


func TestNewChatClient(t *testing.T) {
	//start the server
	eventMux := new(event.TypeMux)
	server := node.NewNode(8011,true,eventMux,3)

	chatClient,err := NewPeerByString("localhost:8011")
	if err != nil{
		log.Fatal("Connect failed")
		server.StopServer()
	}

	msg,err2 := chatClient.Chat(&peermessage.Message{
		MessageType:peermessage.Message_HELLO,
		Payload:[]byte("Hello"),
	})

	if err2 != nil{
		log.Fatal("Failed to send a message")
		server.StopServer()
	}else{
		log.Info(msg)
		server.StopServer()
	}
}

