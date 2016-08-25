package p2p

import (
	"hyperchain-alpha/p2p/peermessage"
	"log"
)
// HelloHandler hello message handler
type HelloHandler struct{

}
func (this *HelloHandler)ProcessEvent(msg *peermessage.Message)error{
	log.Println(msg.MessageType)
	return nil
}
func NewHelloHandler()*HelloHandler{
	return &HelloHandler{}
}
