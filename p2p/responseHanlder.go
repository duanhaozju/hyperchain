package p2p

import (
	"hyperchain-alpha/p2p/peermessage"
	"log"
)

type ResponseHandler struct{

}

func (this *ResponseHandler)ProcessEvent(msg *peermessage.Message)error{
	log.Println(msg.MessageType)
	return nil
}

func NewResponseHandler()*HelloHandler{
	return &HelloHandler{}
}
