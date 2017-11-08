package admin

import (
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	pb "hyperchain/common/protos"
	"hyperchain/manager/event"
	"hyperchain/service/hypexec/controller"
)

var logger *logging.Logger

func init() {
	logger = logging.MustGetLogger("handler")
}

type AdminHandler struct {
	ecMgr controller.ExecutorController
	Ch    chan *ResponseEventWrapper
}

type ResponseEventWrapper struct { //TODO: Fix it
	rspId uint64
	are   *event.AdminResponseEvent
}

func NewAdminHandler(ecMgr controller.ExecutorController) *AdminHandler {
	return &AdminHandler{
		ecMgr: ecMgr,
		Ch:    make(chan *ResponseEventWrapper),
	}
}

func (ah *AdminHandler) Handle(client pb.Dispatcher_RegisterClient, msg *pb.IMessage) {
	switch msg.Event {
	case pb.Event_AddNamespaceEvent:
		e := &event.AddNamespaceEvent{}
		err := proto.Unmarshal(msg.Payload, e)
		if err != nil {
			logger.Error(err)
			return
		}
		namespace := e.GetNamespace()
		err = ah.ecMgr.StartExecutorServiceByName(namespace)
		es := &event.AdminResponseEvent{}
		if err != nil {
			logger.Errorf("StartExecutorServiceByName namespce %s failed, error is %s", namespace, err)
			es.Ok = false
			es.Msg = err.Error()
		} else {
			es.Ok = true
		}
		ah.Ch <- &ResponseEventWrapper{
			rspId: msg.Id,
			are:   es,
		}
	case pb.Event_DeleteNamespaceEvent:
		e := &event.DeleteNamespaceEvent{}
		err := proto.Unmarshal(msg.Payload, e)
		if err != nil {
			logger.Error(err)
			return
		}
		namespace := e.GetNamespace()
		err = ah.ecMgr.StopExecutorServiceByName(namespace)
		es := &event.AdminResponseEvent{}
		if err != nil {
			logger.Errorf("StopExecutorServiceByName namespce%s filed, error is %s", namespace, err)
			es.Ok = false
			es.Msg = err.Error()
		} else {
			es.Ok = true
		}
		ah.Ch <- &ResponseEventWrapper{
			rspId: msg.Id,
			are:   es,
		}
	default:
		logger.Error("Undefined event.")
	}
}
