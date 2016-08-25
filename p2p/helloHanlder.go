// gRPC manager
// author: Quan Chen
// date: 2016-08-25
// last modified:2016-08-25 Quan Chen
// change log:	define the hello message handler struct
// 		this is used to handle the local hello message
//
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
