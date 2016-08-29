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
	"hyperchain/p2p/peerPool"
)
// HelloHandler hello message handler
type BroadCastHandler struct{

}

func NewBroadCastHandler()*BroadCastHandler{
	return &BroadCastHandler{}
}

// this is the most important handler
func (this *BroadCastHandler)ProcessEvent(msg *peermessage.Message)error{
	log.Println(msg.MessageType)
	// TODO 将消息广播出去
	pPool := peerPool.NewPeerPool(false,false)
	for _,p := range pPool.GetPeers(){
		p.Chat(msg)
	}
	return nil
}

