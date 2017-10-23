package handler

import (
    "hyperchain/manager/event"
    "github.com/op/go-logging"
    pb "hyperchain/common/protos"
    "github.com/golang/protobuf/proto"
    "hyperchain/core/executor"
    "hyperchain/common/client"
)

var logger *logging.Logger

func init() {
    logger = logging.MustGetLogger("handler")
}

type ExecutorHandler struct {
    executor    *executor.Executor
}

func New(executor *executor.Executor) client.Handler {
    return &ExecutorHandler{
        executor:   executor,
    }
}

func (eh *ExecutorHandler) Handle(msg *pb.IMessage) {
    switch msg.Event {
    case pb.Event_ValidationEvent:
        e := &event.ValidationEvent{}
        err := proto.Unmarshal(msg.Payload, e)
        if err != nil {
            logger.Error(err)
            return
        } else {
            logger.Debugf("handle event: %v", e)
        }
        eh.executor.Validate(*e)
    case pb.Event_CommitEvent:
        e := &event.CommitEvent{}
        err := proto.Unmarshal(msg.Payload, e)
        if err != nil {
            logger.Error(err)
            return
        } else {
            logger.Debugf("handle event: %v", e)
        }
        eh.executor.CommitBlock(*e)
    case pb.Event_VCResetEvent:
        e := &event.VCResetEvent{}
        err := proto.Unmarshal(msg.Payload, e)
        if err != nil {
            logger.Error(err)
            return
        } else {
            logger.Debugf("handle event: %v", e)
        }
        eh.executor.Rollback(*e)
    case pb.Event_ChainSyncReqEvent:
        e := &event.ChainSyncReqEvent{}
        err := proto.Unmarshal(msg.Payload, e)
        if err != nil {
            logger.Error(err)
            return
        } else {
            logger.Debugf("handle event: %v", e)
        }
        eh.executor.SyncChain(*e)
    default:
        logger.Error("Undefined event.")
    }
}
