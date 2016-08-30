// author: chenquan
// date: 16-8-30
// last modified: 16-8-30 13:45
// last Modified Author: chenquan
// change log: 
//		
package peerEventHandler

import (
	"testing"
	"hyperchain/p2p/peermessage"
)

func TestResponseHandler_ProcessEvent(t *testing.T) {
	responseHandler := NewResponseHandler()
	responseHandler.ProcessEvent(&peermessage.Message{
		MessageType:peermessage.Message_RESPONSE,
		Payload:[]byte("response"),
	})
}