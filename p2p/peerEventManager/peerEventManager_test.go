package peerEventManager

import "testing"
import (
	"hyperchain-alpha/p2p/peermessage"
	"log"
	"time"
)

type hellohandler struct{

}
func (hh *hellohandler)ProcessEvent(msg *peermessage.Message)error{
	log.Println(msg.MessageType)
	return nil
}
func newHelloHandler()*hellohandler{
	return &hellohandler{}
}

func TestPeerEventManager_PostEvent(t *testing.T) {
	pem := NewPeerEventManager()
	pem.RegisterEvent(peermessage.Message_HELLO,newHelloHandler())
	pem.Start()
	tickCount := 0
	for now := range time.Tick(3* time.Second){
		tickCount += 1
		var msg  peermessage.Message
		msg.MessageType = peermessage.Message_HELLO
		msg.Payload = []byte("hello")
		pem.PostEvent(peermessage.Message_HELLO,msg)
		log.Println(now)
		if tickCount >3{
			break
		}
	}
}
