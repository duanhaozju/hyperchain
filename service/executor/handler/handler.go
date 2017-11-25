package handler

import (
	"github.com/gogo/protobuf/proto"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/common/service/client"
	pb "github.com/hyperchain/hyperchain/common/service/protos"
	"github.com/hyperchain/hyperchain/core/oplog/proto"
	"github.com/hyperchain/hyperchain/service/executor/controller/executor"
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
	var err error
	defer func() { eh.handleError(err) }()

	switch msg.Type {
	case pb.Type_OP_LOG:
		le := &oplog.LogEntry{}
		if err = proto.Unmarshal(msg.Payload, le); err != nil {
			return
		}
		eh.executor.Dispatch(le)
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

func (eh *ExecutorHandler) handleError(err error) {
	if err != nil {
		eh.logger.Error(err)
	}
}
