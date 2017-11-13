package admin

import (
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"hyperchain/common"
	pb "hyperchain/common/protos"
	"hyperchain/manager/event"
	"hyperchain/service/hypexec/controller"
)

var logger *logging.Logger

type AdminHandler struct {
	ecMgr controller.ExecutorController
	Ch    chan *ResponseEventWrapper
}

type ResponseEventWrapper struct { //TODO: Fix it
	rspId uint64
	are   *event.AdminResponseEvent
}

func NewAdminHandler(ecMgr controller.ExecutorController) *AdminHandler {
	logger = common.GetLogger(common.DEFAULT_NAMESPACE, "admin")
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

	//add 3 methods respectively cresponds to hypercli JVM-ops.
	//TODO: generate new service.proto file.
	case pb.Event_StartJVMEvent:
		err := ah.ecMgr.StartJVM()
		es := &event.AdminResponseEvent{}
		if err != nil {
			logger.Errorf("Start JVM manager failed, error is %s", err)
			es.Ok = false
			es.Msg = err.Error()
		} else {
			es.Ok = true
		}
		ah.Ch <- &ResponseEventWrapper{
			rspId: msg.Id,
			are:   es,
		}
	case pb.Event_StopJVMEvent:
		err := ah.ecMgr.StopJVM()
		es := &event.AdminResponseEvent{}
		if err != nil {
			logger.Errorf("Stop JVM manager failed, error is %s", err)
			es.Ok = false
			es.Msg = err.Error()
		} else {
			es.Ok = true
		}
		ah.Ch <- &ResponseEventWrapper{
			rspId: msg.Id,
			are:   es,
		}
	case pb.Event_RestartJVMEvent:
		err := ah.ecMgr.RestartJVM()
		es := &event.AdminResponseEvent{}
		if err != nil {
			logger.Errorf("Stop JVM manager failed, error is %s", err)
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
