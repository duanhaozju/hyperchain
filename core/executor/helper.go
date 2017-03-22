package executor
import (
	"hyperchain/event"
	"hyperchain/protos"
	"hyperchain/recovery"
	"github.com/golang/protobuf/proto"
	edb "hyperchain/core/db_utils"
	"hyperchain/core/types"
)
type Helper struct {
	msgQ *event.TypeMux
}

func NewHelper(msgQ *event.TypeMux) *Helper {
	return &Helper{
		msgQ: msgQ,
	}
}

func (helper *Helper) Post(ev interface{}) {
	helper.msgQ.Post(ev)
}

// informConsensus - communicate with consensus module.
func (executor *Executor) informConsensus(informType int, message interface{}) error {
	switch informType {
	case NOTIFY_REMOVE_CACHE:
		executor.logger.Debugf("[Namespace = %s] inform consenus remove cache", executor.namespace)
		msg := message.(protos.RemoveCache)
		executor.helper.Post(event.ExecutorToConsensusEvent{
			Payload: msg,
			Type:    NOTIFY_REMOVE_CACHE,
		})
	case NOTIFY_VALIDATION_RES:
		executor.logger.Debugf("[Namespace = %s] inform consenus validation result", executor.namespace)
		msg := message.(protos.ValidatedTxs)
		executor.helper.Post(event.ExecutorToConsensusEvent{
			Payload: msg,
			Type:    NOTIFY_VALIDATION_RES,
		})
	case NOTIFY_VC_DONE:
		executor.logger.Debugf("[Namespace = %s] inform consenus vc done", executor.namespace)
		msg := message.(protos.VcResetDone)
		executor.helper.Post(event.ExecutorToConsensusEvent{
			Payload: msg,
			Type:    NOTIFY_VC_DONE,
		})
	case NOTIFY_SYNC_DONE:
		executor.logger.Debugf("[Namespace = %s] inform consenus sync done", executor.namespace)
		msg := message.(protos.StateUpdatedMessage)
		executor.helper.Post(event.ExecutorToConsensusEvent{
			Payload: msg,
			Type:    NOTIFY_SYNC_DONE,
		})
	default:
		return NoDefinedCaseErr
	}
	return nil
}

// informP2P - communicate with p2p module.
func (executor *Executor) informP2P(informType int, message ...interface{}) error {
	switch informType {
	case NOTIFY_BROADCAST_DEMAND:
		executor.logger.Debugf("[Namespace = %s] inform p2p broadcast demand", executor.namespace)
		required := &recovery.CheckPointMessage{
			RequiredNumber: executor.status.syncFlag.SyncTarget,
			CurrentNumber:  edb.GetHeightOfChain(executor.namespace),
			PeerId:         executor.status.syncFlag.LocalId,
		}
		payload, err := proto.Marshal(required)
		if err != nil {
			executor.logger.Errorf("[Namespace = %s] SendSyncRequest marshal message failed", executor.namespace)
			return err
		}
		executor.helper.Post(event.ExecutorToP2PEvent{
			Payload: payload,
			Peers:   executor.status.syncFlag.SyncPeers,
			Type:    NOTIFY_BROADCAST_DEMAND,
		})
		return nil
	case NOTIFY_UNICAST_BLOCK:
		executor.logger.Debugf("[Namespace = %s] inform p2p unicast block", executor.namespace)
		id := message[0].(uint64)
		peerId := message[1].(uint64)
		block, err := edb.GetBlockByNumber(executor.namespace, id)
		if err != nil {
			executor.logger.Errorf("[Namespace = %s] no demand block number: %d", executor.namespace, id)
			return err
		}
		payload, err := proto.Marshal(block)
		if err != nil {
			executor.logger.Errorf("[Namespace = %s] marshal block failed", executor.namespace)
			return err
		}
		executor.helper.Post(event.ExecutorToP2PEvent{
			Payload: payload,
			Type:    NOTIFY_UNICAST_BLOCK,
			Peers:   []uint64{peerId},
		})
		return nil
	case NOTIFY_UNICAST_INVALID:
		executor.logger.Debugf("[Namespace = %s] inform p2p unicast invalid tx", executor.namespace)
		r := message[0].(*types.InvalidTransactionRecord)
		payload, err := proto.Marshal(r)
		if err != nil {
			executor.logger.Errorf("[Namespace = %s] marshal invalid record error", executor.namespace)
			return err
		}
		executor.helper.Post(event.ExecutorToP2PEvent{
			Payload: payload,
			Type:    NOTIFY_UNICAST_INVALID,
			Peers:   []uint64{r.Tx.Id},
		})
		return nil
	case NOTIFY_BROADCAST_SINGLE:
		executor.logger.Debugf("[Namespace = %s] inform p2p broadcast single demand", executor.namespace)
		id := message[0].(uint64)
		syncRequest := &recovery.CheckPointMessage{
			RequiredNumber: id,
			CurrentNumber:  edb.GetHeightOfChain(executor.namespace),
			PeerId:         executor.status.syncFlag.LocalId,
		}
		payload, err := proto.Marshal(syncRequest)
		if err != nil {
			executor.logger.Errorf("[Namespace = %s] broadcastDemandBlock, marshal message failed", executor.namespace)
			return err
		}
		executor.helper.Post(event.ExecutorToP2PEvent{
			Payload:  payload,
			Type:     NOTIFY_BROADCAST_SINGLE,
			Peers:    executor.status.syncFlag.SyncPeers,
		})
		return nil
	case NOTIFY_SYNC_REPLICA:
		executor.logger.Debugf("[Namespace = %s] inform p2p sync replica", executor.namespace)
		payload, _ := proto.Marshal(message[0].(*types.Chain))
		executor.helper.Post(event.ExecutorToP2PEvent{
			Payload: payload,
			Type:    NOTIFY_SYNC_REPLICA,
		})
		return nil
	default:
		return NoDefinedCaseErr
	}
	return nil
}

