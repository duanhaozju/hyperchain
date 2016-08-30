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

func TestKeepAliveHandler_ProcessEvent(t *testing.T) {
	keepAliveHandler := NewKeepAliveHandler()
	keepAliveHandler.ProcessEvent(&peermessage.Message{
		MessageType:peermessage.Message_KEEPALIVE,
		Payload:[]byte("KeepAlive"),
	})
}
