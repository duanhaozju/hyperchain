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

type ResponseHandler struct{
	eventManager *peerEventManager.PeerEventManager

}

// response message has two conditions:
// inner system need to response the peer keep alive/hello message

func (this *ResponseHandler)ProcessEvent(msg *peermessage.Message)error{
	log.Println(msg.MessageType)
	log.Println(string(msg.Payload))
	return nil
}

func NewResponseHandler(eventManager *peerEventManager.PeerEventManager)*ResponseHandler{
	return &ResponseHandler{eventManager:eventManager}
}
