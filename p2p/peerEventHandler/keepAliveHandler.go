// author: chenquan
// date: 16-8-26
// last modified: 16-8-26 10:32
// last Modified Author: chenquan
// change log: 
//		
package peerEventHandler

import (
	"hyperchain-alpha/p2p/peermessage"
	"log"
)

type KeepAliveHandler struct{

}

func (this *KeepAliveHandler)ProcessEvent(msg *peermessage.Message)error{
	log.Println(msg.MessageType)
	return nil
}

func NewKeepAliveHandler()*HelloHandler{
	return &HelloHandler{}
}
