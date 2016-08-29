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
	"hyperchain/p2p/peerEventManager"
	pb "hyperchain/p2p/peermessage"
)
// HelloHandler hello message handler
type BroadCastHandler struct{
	eventManager *peerEventManager.PeerEventManager

}

func NewBroadCastHandler(eventManager *peerEventManager.PeerEventManager)*BroadCastHandler{
	return &BroadCastHandler{eventManager:eventManager}
}

// this is the most important handler
// 广播消息只会来自本地触发,外部广播信息只需要接收上报即可
func (this *BroadCastHandler)ProcessEvent(msg *peermessage.Message)error{
	log.Println(msg.MessageType)
	// TODO 将消息广播出去
	pPool := peerPool.NewPeerPool(false,false)
	for _,peer := range pPool.GetPeers(){
		resMsg,err :=peer.Chat(msg)
		if err != nil{
			log.Println("Broadcast failed,Node",peer.Addr)
		}else{
			log.Println("resMsg:",string(resMsg.Payload))
			this.eventManager.PostEvent(pb.Message_RESPONSE,*resMsg)
		}
	}


	return nil
}

