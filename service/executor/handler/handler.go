package handler

import (
	"github.com/hyperchain/hyperchain/common"
	pb "github.com/hyperchain/hyperchain/common/protos"
	"github.com/hyperchain/hyperchain/common/service/client"
	"github.com/hyperchain/hyperchain/core/executor"
	"github.com/op/go-logging"
)

type ExecutorHandler struct {
	logger   *logging.Logger
	executor *executor.Executor
}

func New(namespace string, executor *executor.Executor) client.Handler {
	return &ExecutorHandler{
		executor: executor,
		logger:   common.GetLogger(namespace, "executor"),
	}
}

func (eh *ExecutorHandler) handleSyncMsg(client pb.Dispatcher_RegisterClient, msg *pb.IMessage) {
	//e := &event.MemChainEvent{} //TODO: fix it, parse event by msg.Event
	//err := proto.Unmarshal(msg.Payload, e)
	//if err != nil {
	//	eh.logger.Criticalf("MemChainEvent unmarshal err: %v", err)
	//}
	//go func() {
	//	m := chain.GetMemChain(e.Namespace, e.Checkpoint)
	//	payload, err := proto.Marshal(m)
	//	if err != nil {
	//		eh.logger.Error(err)
	//		return
	//	}
	//	msg := &pb.IMessage{
	//		Type:    pb.Type_RESPONSE,
	//		From:    pb.FROM_EXECUTOR,
	//		Id:      msg.Id, // this id must be equal to the request message id.
	//		Ok:      true,
	//		Payload: payload,
	//	}
	//	client.Send(msg)
	//}()
}

func (eh *ExecutorHandler) handleAsyncMsg(client pb.Dispatcher_RegisterClient, msg *pb.IMessage) {

	switch msg.Event {
	default:
		eh.logger.Errorf("Undefined event %v", msg)
	}
}

func (eh *ExecutorHandler) Handle(client pb.Dispatcher_RegisterClient, msg *pb.IMessage) {
	if msg.Type == pb.Type_SYNC_REQUEST {
		eh.handleSyncMsg(client, msg)
	} else {
		eh.handleAsyncMsg(client, msg)
	}
}
