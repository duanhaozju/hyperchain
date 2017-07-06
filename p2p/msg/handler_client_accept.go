package msg

import (
	pb "hyperchain/p2p/message"
	"hyperchain/admittance"
	"github.com/op/go-logging"
	"hyperchain/p2p/hts"
)

type ClientAcceptMsgHandler struct {
	serverHTS *hts.ServerHTS
	logger *logging.Logger
}

func NewClientAcceptHandler(serverhts *hts.ServerHTS,logger *logging.Logger)*ClientHelloMsgHandler{
	return &ClientAcceptMsgHandler{
		serverHTS:serverhts,
		logger:logger,
	}
}

func (h  *ClientAcceptMsgHandler) Process() {
	 h.logger.Info("client accept message not support stream message, so need not listen the stream message.")
}

func (h  *ClientAcceptMsgHandler) Teardown() {
	h.logger.Info("client accept message not support the stream message, so needn't to be close")
}

func (h *ClientAcceptMsgHandler)Receive() chan<- interface{}{
	h.logger.Info("client accpet message not support stream message")
}

func (h *ClientAcceptMsgHandler)Execute(msg *pb.Message) (*pb.Message,error){
	h.logger.Infof("got a client accept message %v",msg)
	rsp  := &pb.Message{
		MessageType:pb.MsgType_SERVERDONE,
	}
	return rsp,nil
}
