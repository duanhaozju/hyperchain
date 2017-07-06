package msg

import (
	pb "hyperchain/p2p/message"
	"hyperchain/admittance"
	"github.com/op/go-logging"
)

type ClientHelloMsgHandler struct {
	cm *admittance.CAManager
	logger *logging.Logger
}

func NewClientHelloHandler(cm admittance.CAManager,logger *logging.Logger)*ClientHelloMsgHandler{
	return &ClientHelloMsgHandler{
		cm:cm,
		logger:logger,
	}
}

func (h  *ClientHelloMsgHandler) Process() {
	 h.logger.Info("client hello message not support stream message, so need not listen the stream message.")
}

func (h  *ClientHelloMsgHandler) Teardown() {
	h.logger.Info("client hello msg not support the stream message, so needn't to be close")
}

func (h *ClientHelloMsgHandler)Receive() chan<- interface{}{
	h.logger.Info("client hello message not support stream message")
}

func (h *ClientHelloMsgHandler)Execute(msg *pb.Message) (*pb.Message,error){
	h.logger.Infof("got a client hello message %v",msg)
	rsp  := &pb.Message{
		MessageType:pb.MsgType_SERVERHELLO,
	}
	return rsp,nil
}
