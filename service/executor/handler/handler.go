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
