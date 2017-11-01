package handler

import (
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/common/service/client"
	pb "hyperchain/common/protos"
	"hyperchain/core/executor"
	"hyperchain/core/ledger/chain"
	"hyperchain/manager/event"
)

var logger *logging.Logger

func init() {
	logger = logging.MustGetLogger("handler")
}

type ExecutorHandler struct {
	logger   *logging.Logger
	executor *executor.Executor
}

func New(executor *executor.Executor) client.Handler {
	return &ExecutorHandler{
		executor: executor,
		logger:   common.GetLogger("global", "executor"),
	}
}

func (eh *ExecutorHandler) handleSyncMsg(client pb.Dispatcher_RegisterClient, msg *pb.IMessage) {
	e := &event.MemChainEvent{} //TODO: fix it, parse event by msg.Event
	err := proto.Unmarshal(msg.Payload, e)
	if err != nil {
		eh.logger.Criticalf("MemChainEvent unmarshal err: %v", err)
	}
	go func() {
		m := chain.GetMemChain(e.Namespace, e.Checkpoint)
		payload, err := proto.Marshal(m)
		if err != nil {
			eh.logger.Error(err)
			return
		}
		msg := &pb.IMessage{
			Type:    pb.Type_RESPONSE,
			From:    pb.FROM_EXECUTOR,
			Id:      msg.Id, // this id must be equal to the request message id.
			Ok:      true,
			Payload: payload,
		}
		client.Send(msg)
	}()
}

func (eh *ExecutorHandler) handleAsyncMsg(client pb.Dispatcher_RegisterClient, msg *pb.IMessage) {
	switch msg.Event {
	case pb.Event_ValidationEvent:
		e := &event.ValidationEvent{}
		err := proto.Unmarshal(msg.Payload, e)
		if err != nil {
			logger.Error(err)
			return
		} else {
		}
		eh.executor.Validate(*e)
	case pb.Event_CommitEvent:
		e := &event.CommitEvent{}
		err := proto.Unmarshal(msg.Payload, e)
		if err != nil {
			logger.Error(err)
			return
		} else {
		}
		eh.executor.CommitBlock(*e)
	case pb.Event_VCResetEvent:
		e := &event.VCResetEvent{}
		err := proto.Unmarshal(msg.Payload, e)
		if err != nil {
			logger.Error(err)
			return
		} else {
		}
		eh.executor.Rollback(*e)
	case pb.Event_ChainSyncReqEvent:
		e := &event.ChainSyncReqEvent{}
		err := proto.Unmarshal(msg.Payload, e)
		if err != nil {
			logger.Error(err)
			return
		} else {
		}
		eh.executor.SyncChain(*e)
	default:
		logger.Error("Undefined event.")
	}
}

func (eh *ExecutorHandler) Handle(client pb.Dispatcher_RegisterClient, msg *pb.IMessage) {
	if msg.Type == pb.Type_SYNC_REQUEST {
		eh.handleSyncMsg(client, msg)
	} else {
		eh.handleAsyncMsg(client, msg)
	}
}
