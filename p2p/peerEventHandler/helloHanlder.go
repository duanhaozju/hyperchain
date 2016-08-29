// author: chenquan
// date: 16-8-25
// last modified: 16-8-25 20:01
// last Modified Author: chenquan
// change log:
//
package peerEventHandler

import (
	"hyperchain/p2p/peermessage"
	"log"
	"hyperchain/p2p/peerEventManager"
)
// HelloHandler hello message handler
type HelloHandler struct{
	eventManager *peerEventManager.PeerEventManager

}
// peer send a hello message should handled here
func (this *HelloHandler)ProcessEvent(msg *peermessage.Message)error{
	log.Println(msg.MessageType)
	return nil
}
func NewHelloHandler(eventManager *peerEventManager.PeerEventManager)*HelloHandler{
	return &HelloHandler{eventManager:eventManager}
}
