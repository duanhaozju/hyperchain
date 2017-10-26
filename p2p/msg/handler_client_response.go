package msg

import (
	"github.com/hyperchain/hyperchain/p2p/hts"
	pb "github.com/hyperchain/hyperchain/p2p/message"

	"github.com/op/go-logging"
)

type ClientAcceptMsgHandler struct {
	serverHTS *hts.ServerHTS
	logger    *logging.Logger
}

func NewClientAcceptHandler(serverhts *hts.ServerHTS, logger *logging.Logger) *ClientAcceptMsgHandler {
	return &ClientAcceptMsgHandler{
		serverHTS: serverhts,
		logger:    logger,
	}
}

func (h *ClientAcceptMsgHandler) Process() {
	h.logger.Info("client accept message not support stream message, so need not listen the stream message.")
}

func (h *ClientAcceptMsgHandler) Teardown() {
	h.logger.Info("client accept message not support the stream message, so needn't to be close")
}

func (h *ClientAcceptMsgHandler) Receive() chan<- interface{} {
	h.logger.Info("client accpet message not support stream message")
	return nil
}

func (h *ClientAcceptMsgHandler) Execute(msg *pb.Message) (*pb.Message, error) {
	h.logger.Debugf("got a client accept message ,msg type %s", msg.MessageType)
	rsp := &pb.Message{
		MessageType: pb.MsgType_SERVERDONE,
		Payload:     []byte("server send a server Done message(this message is from server)"),
	}
	return rsp, nil
}
