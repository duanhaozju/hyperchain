// author: chenquan
// date: 16-8-25
// last modified: 16-8-25 20:01
// last Modified Author: chenquan
// change log:
//
package peerEventHandler

import (
	"hyperchain-alpha/p2p/peermessage"
	"log"
)

type ResponseHandler struct{

}

func (this *ResponseHandler)ProcessEvent(msg *peermessage.Message)error{
	log.Println(msg.MessageType)
	return nil
}

func NewResponseHandler()*HelloHandler{
	return &HelloHandler{}
}
