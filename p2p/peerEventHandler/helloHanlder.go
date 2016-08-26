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
// HelloHandler hello message handler
type HelloHandler struct{

}
func (this *HelloHandler)ProcessEvent(msg *peermessage.Message)error{
	log.Println(msg.MessageType)
	return nil
}
func NewHelloHandler()*HelloHandler{
	return &HelloHandler{}
}
