// author: chenquan
// date: 16-8-25
// last modified: 16-8-29 13:51
// last Modified Author: chenquan
// change log: 1. test all set of event types
//
package peerEventManager

import "testing"
import (
	"hyperchain/p2p/peermessage"
	"log"
	"time"
	"hyperchain/p2p/peerEventHandler"
)



func TestPeerEventManager_PostEvent(t *testing.T) {
	pem := NewPeerEventManager()
	pem.RegisterEvent(peermessage.Message_HELLO,peerEventHandler.NewHelloHandler())
	pem.RegisterEvent(peermessage.Message_KEEPALIVE,peerEventHandler.NewKeepAliveHandler())
	pem.RegisterEvent(peermessage.Message_CONSUS,peerEventHandler.NewBroadCastHandler())
	pem.RegisterEvent(peermessage.Message_RESPONSE,peerEventHandler.NewResponseHandler())
	pem.Start()
	tickCount := 0
	for range time.Tick(3* time.Second){
		tickCount += 1
		var msg  peermessage.Message
		log.Println("Event HEllo")
		msg.MessageType = peermessage.Message_HELLO
		msg.Payload = []byte("hello")
		pem.PostEvent(peermessage.Message_HELLO,msg)
		//log.Println(now)
		log.Println("Event Keep Alive")
		msg.MessageType = peermessage.Message_KEEPALIVE
		msg.Payload = []byte("keep alive")
		pem.PostEvent(peermessage.Message_KEEPALIVE,msg)
		//log.Println(now)

		log.Println("Event Message_RESPONSE")
		msg.MessageType = peermessage.Message_RESPONSE
		msg.Payload = []byte("Message_RESPONSE")
		pem.PostEvent(peermessage.Message_RESPONSE,msg)
		//log.Println(now)

		log.Println("Event Message_CONSUS")
		msg.MessageType = peermessage.Message_CONSUS
		msg.Payload = []byte("Message_CONSUS")
		pem.PostEvent(peermessage.Message_CONSUS,msg)
		//log.Println(now)
		if tickCount >1{
			break
		}
	}
}
