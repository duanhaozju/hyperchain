package handler

import (
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"hyperchain/common"
	pb "hyperchain/common/protos"
	"hyperchain/common/service/client"
	"hyperchain/core/executor"
	"hyperchain/core/ledger/chain"
	"hyperchain/manager/event"
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
			eh.logger.Error(err)
			return
		}
		eh.executor.Validate(*e)
	case pb.Event_CommitEvent:
		e := &event.CommitEvent{}
		err := proto.Unmarshal(msg.Payload, e)
		if err != nil {
			eh.logger.Error(err)
			return
		}
		eh.executor.CommitBlock(*e)
	case pb.Event_VCResetEvent:
		e := &event.VCResetEvent{}
		err := proto.Unmarshal(msg.Payload, e)
		if err != nil {
			eh.logger.Error(err)
			return
		}
		eh.executor.Rollback(*e)
	case pb.Event_ChainSyncReqEvent:
		e := &event.ChainSyncReqEvent{}
		err := proto.Unmarshal(msg.Payload, e)
		if err != nil {
			eh.logger.Error(err)
			return
		}
		eh.executor.SyncChain(*e)
	case pb.Event_StoreInvalidTransactionEvent:
		eh.executor.StoreInvalidTransaction(msg.Payload)
	case pb.Event_ReceiveReplicaInfoEvent:
		eh.executor.ReceiveReplicaInfo(msg.Payload)
	case pb.Event_ReceiveSyncBlocksEvent:
		eh.executor.ReceiveSyncBlocks(msg.Payload)
	case pb.Event_ReceiveSyncRequestEvent:
		eh.executor.ReceiveSyncRequest(msg.Payload)
	case pb.Event_ReceiveWorldStateSyncRequestEvent:
		eh.executor.ReceiveWorldStateSyncRequest(msg.Payload)
	case pb.Event_ReceiveWorldStateEvent:
		eh.executor.ReceiveWorldState(msg.Payload)
	case pb.Event_ReceiveWsHandshakeEvent:
		eh.executor.ReceiveWsHandshake(msg.Payload)
	case pb.Event_ReceiveWsAckEvent:
		eh.executor.ReceiveWsAck(msg.Payload)
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
