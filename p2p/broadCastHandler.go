package p2p

import (
	"hyperchain-alpha/p2p/peermessage"
	"log"
)
// HelloHandler hello message handler
type BroadCastHandler struct{

}

func NewBroadCastHandler()*BroadCastHandler{
	return &BroadCastHandler{}
}

func (this *BroadCastHandler)ProcessEvent(msg *peermessage.Message)error{
	log.Println(msg.MessageType)
	// TODO 将消息广播出去
	return nil
}

