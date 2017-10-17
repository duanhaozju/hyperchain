package handler

import (
	"github.com/gogo/protobuf/proto"
	"github.com/op/go-logging"
	"hyperchain/manager/event"
	pb "hyperchain/common/protos"
	"hyperchain/common/service"
)

var logger *logging.Logger

func init() {
	logger = logging.MustGetLogger("handler")
}

type NetworkMessageHandler struct {
}

func New() service.Handler {
	return &NetworkMessageHandler{}
}

func (nmh *NetworkMessageHandler) Handle(msg *pb.Message) {
	switch msg.Event {
	case pb.Event_InformPrimaryEvent:
		event := &event.InformPrimaryEvent{}
		err := proto.Unmarshal(msg.Payload, event)
		if err != nil {
			logger.Error(err)
		} else {
			logger.Debugf("handle event: %v", event)
		}
	default:
		logger.Error("Undefined event.")
	}
}
