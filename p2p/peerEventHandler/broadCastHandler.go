// author: chenquan
// date: 16-8-25
// last modified: 16-8-25 20:01
// last Modified Author: chenquan
// change log:
//
package peerEventHandler


import (
	"hyperchain/p2p/peermessage"
	"hyperchain/p2p/peerPool"
	peer "hyperchain/p2p/peer"
	"fmt"
	"github.com/op/go-logging"
)

var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("p2p/peerEventHandler")
}
// HelloHandler hello message handler
type BroadCastHandler struct{
}

func NewBroadCastHandler()*BroadCastHandler{
	return &BroadCastHandler{}
}

// this is the most important handler
// 广播消息只会来自本地触发,外部广播信息只需要接收上报即可
func (this *BroadCastHandler)ProcessEvent(msg *peermessage.Message)error{
	log.Info(msg.MessageType)
	pPool := peerPool.NewPeerPool(false,false)
	fmt.Println("现在有节点数目:",pPool.GetAliveNodeNum())
	ps := pPool.GetPeers()
	fmt.Println("现在有节点数目:",len(ps))
	for _,peer := range pPool.GetPeers(){
		fmt.Println("广播....")
		resMsg,err := peer.Chat(msg)
		if err != nil{
			log.Error("Broadcast failed,Node",peer.Addr)
		}else{
			log.Info("resMsg:",string(resMsg.Payload))
			//this.eventManager.PostEvent(pb.Message_RESPONSE,*resMsg)
		}
	}
	return nil
}

func broadcast(msg *peermessage.Message, peer *peer.Peer){
	resMsg,err := peer.Chat(msg)
	if err != nil{
		log.Error("Broadcast failed,Node",peer.Addr)
	}else{
		log.Info("resMsg:",string(resMsg.Payload))

		//this.eventManager.PostEvent(pb.Message_RESPONSE,*resMsg)
	}
}

