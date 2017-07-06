package msg

import (
	"hyperchain/admittance"
	pb "hyperchain/p2p/message"

	"github.com/op/go-logging"
)

type ServerHelloMsgHandler struct {
	cm     *admittance.CAManager
	logger *logging.Logger
}

func NewServerHelloHandler(cm admittance.CAManager, logger *logging.Logger) *ServerHelloMsgHandler {
	return &ClientHelloMsgHandler{
		cm:     cm,
		logger: logger,
	}
}

func (h *ServerHelloMsgHandler) Process() {
	h.logger.Info("server hello message not support stream message, so need not listen the stream message.")
}

func (h *ServerHelloMsgHandler) Teardown() {
	h.logger.Info("server hello msg not support the stream message, so needn't to be close")
}

func (h *ServerHelloMsgHandler) Receive() chan<- interface{} {
	h.logger.Info("server hello message not support stream message")
}

func (h *ServerHelloMsgHandler) Execute(msg *pb.Message) (*pb.Message, error) {
	h.logger.Infof("got a server hello message %v", msg)
	//May return a client ACCEPT or client REJECT message
	rsp := &pb.Message{
		MessageType: pb.MsgType_CLIENTACCEPT,
	}
	return rsp, nil
}
