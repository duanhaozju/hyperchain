// author: chenquan
// date: 16-8-25
// last modified: 16-8-25 20:01
// last Modified Author: chenquan
// change log:
//
package peerEventHandler

import (
	"hyperchain/p2p/peermessage"
)

type ResponseHandler struct{
}


// ProcessEvent Response Type Message handler
// response message has two conditions:
// inner system need to response the peer keep alive/hello message
func (this *ResponseHandler)ProcessEvent(msg *peermessage.Message)error{
	log.Info(msg.MessageType)
	log.Info(string(msg.Payload))
	return nil
}

// return a new ResponseHandler instance
func NewResponseHandler()*ResponseHandler{
	return &ResponseHandler{}
}
