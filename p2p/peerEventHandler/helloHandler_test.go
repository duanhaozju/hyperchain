// author: chenquan
// date: 16-8-30
// last modified: 16-8-30 13:44
// last Modified Author: chenquan
// change log: 
//		
package peerEventHandler

import (
	"testing"
	"hyperchain/p2p/peermessage"
)

func TestHelloHandler_ProcessEvent(t *testing.T) {
	helloHandler := NewHelloHandler()
	helloHandler.ProcessEvent(&peermessage.Message{
		MessageType:peermessage.Message_HELLO,
		Payload:[]byte("hello"),
	})
}