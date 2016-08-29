// author: chenquan
// date: 16-8-26
// last modified: 16-8-26 10:32
// last Modified Author: chenquan
// change log: 
//		
package peerEventHandler

import (
	"hyperchain/p2p/peermessage"
	"log"
)

type KeepAliveHandler struct {
}
// keep live message only peer can send so should send a response message to peer

func (this *KeepAliveHandler)ProcessEvent(msg *peermessage.Message)error{
	log.Println(msg.MessageType)
	// 返回一个response消息
	return nil
}

func NewKeepAliveHandler()*KeepAliveHandler{
	return &KeepAliveHandler{}
}
