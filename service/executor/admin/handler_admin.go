package admin

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	pb "hyperchain/common/protos"
	"hyperchain/manager/event"
	"hyperchain/service/executor/manager"
)

var logger *logging.Logger

func init() {
	logger = logging.MustGetLogger("handler")
}

type AdminHandler struct {
	ecMgr manager.ExecutorManager
	Ch    chan *event.AdminResponseEvent
}

func NewAdminHandler(ecMgr manager.ExecutorManager) *AdminHandler {
	return &AdminHandler{
		ecMgr: ecMgr,
		Ch:    make(chan *event.AdminResponseEvent),
	}
}

func (ah *AdminHandler) Handle(msg *pb.IMessage) {
	switch msg.Event {
	case pb.Event_AddNamespaceEvent:
		e := &event.AddNamespaceEvent{}
		err := proto.Unmarshal(msg.Payload, e)
		if err != nil {
			logger.Error(err)
			return
		}
		namespace := e.GetNamespace()
		err = ah.ecMgr.Start(namespace)
		es := &event.AdminResponseEvent{}
		if err != nil {
			logger.Errorf("Start namespce%s filed, error is %s", namespace, err)
			es.Ok = false
		} else {
			es.Ok = true
		}
		ah.Ch <- es
	case pb.Event_DeleteNamespaceEvent:
		e := &event.DeleteNamespaceEvent{}
		err := proto.Unmarshal(msg.Payload, e)
		if err != nil {
			logger.Error(err)
			return
		}
		namespace := e.GetNamespace()
		err = ah.ecMgr.Stop(namespace)
		es := &event.AdminResponseEvent{}
		if err != nil {
			logger.Errorf("Stop namespce%s filed, error is %s", namespace, err)
			es.Ok = false
		} else {
			es.Ok = true
		}
		ah.Ch <- es
	default:
		logger.Error("Undefined event.")
	}
}
