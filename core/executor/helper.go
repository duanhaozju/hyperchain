package executor
import (
	"hyperchain/manager/event"
	"hyperchain/manager/protos"
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
		executor.logger.Debug("inform consenus remove cache")
		msg, ok := message.(protos.RemoveCache)
		if !ok {
			return InvalidParamsErr
		}
		executor.helper.Post(event.ExecutorToConsensusEvent{
			Payload: msg,
			Type:    NOTIFY_REMOVE_CACHE,
		})
	case NOTIFY_VALIDATION_RES:
		executor.logger.Debugf("[Namespace = %s] inform consenus validation result", executor.namespace)
		msg, ok := message.(protos.ValidatedTxs)
		if !ok {
			return InvalidParamsErr
		}
		executor.helper.Post(event.ExecutorToConsensusEvent{
			Payload: msg,
			Type:    NOTIFY_VALIDATION_RES,
		})
	case NOTIFY_VC_DONE:
		executor.logger.Debug("inform consenus vc done")
		msg, ok := message.(protos.VcResetDone)
		if !ok {
			return InvalidParamsErr
		}
		executor.helper.Post(event.ExecutorToConsensusEvent{
			Payload: msg,
			Type:    NOTIFY_VC_DONE,
		})
	case NOTIFY_SYNC_DONE:
		executor.logger.Debug("inform consenus sync done")
		msg, ok := message.(protos.StateUpdatedMessage)
		if !ok {
			return InvalidParamsErr
		}
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
		executor.logger.Debug("inform p2p broadcast demand")
		if len(message) != 2 {
			return InvalidParamsErr
		}
		requiredNumber, ok := message[0].(uint64)
		if !ok {
			return InvalidParamsErr
		}
		currentNumber, ok := message[1].(uint64)
		if !ok {
			return InvalidParamsErr
		}
		required := ChainSyncRequest{
			RequiredNumber: requiredNumber,
			CurrentNumber:  currentNumber,
			PeerId:         executor.status.syncFlag.LocalId,
		}
		payload, err := proto.Marshal(&required)
		if err != nil {
			executor.logger.Errorf("sync chain request marshal message failed")
			return err
		}
		executor.helper.Post(event.ExecutorToP2PEvent{
			Payload: payload,
			Peers:   []uint64{message[2].(uint64)},
			Type:    NOTIFY_BROADCAST_DEMAND,
		})
		return nil
	case NOTIFY_UNICAST_BLOCK:
		executor.logger.Debug("inform p2p unicast block")
		if len(message) != 2 {
			return InvalidParamsErr
		}
		id, ok := message[0].(uint64)
		if !ok {
			return InvalidParamsErr
		}
		peerId, ok := message[1].(uint64)
		if !ok {
			return InvalidParamsErr
		}
		block, err := edb.GetBlockByNumber(executor.namespace, id)
		if err != nil {
			executor.logger.Errorf("no demand block number: %d", id)
			return err
		}
		payload, err := proto.Marshal(block)
		if err != nil {
			executor.logger.Error("marshal block failed")
			return err
		}
		executor.helper.Post(event.ExecutorToP2PEvent{
			Payload: payload,
			Type:    NOTIFY_UNICAST_BLOCK,
			Peers:   []uint64{peerId},
		})
		return nil
	case NOTIFY_UNICAST_INVALID:
		executor.logger.Debug("inform p2p unicast invalid tx")
		if len(message) != 1 {
			return InvalidParamsErr
		}
		r, ok := message[0].(*types.InvalidTransactionRecord)
		if !ok {
			return InvalidParamsErr
		}
		payload, err := proto.Marshal(r)
		if err != nil {
			executor.logger.Error("marshal invalid record error")
			return err
		}
		executor.helper.Post(event.ExecutorToP2PEvent{
			Payload: payload,
			Type:    NOTIFY_UNICAST_INVALID,
			Peers:   []uint64{r.Tx.Id},
		})
		return nil
	case NOTIFY_BROADCAST_SINGLE:
		executor.logger.Debug("inform p2p broadcast single demand")
		if len(message) != 1 {
			return InvalidParamsErr
		}
		id, ok := message[0].(uint64)
		if !ok {
			return InvalidParamsErr
		}
		request := ChainSyncRequest{
			RequiredNumber: id,
			CurrentNumber:  edb.GetHeightOfChain(executor.namespace),
			PeerId:         executor.status.syncFlag.LocalId,
		}
		payload, err := proto.Marshal(&request)
		if err != nil {
			executor.logger.Error("broadcast demand block, marshal message failed")
			return err
		}
		executor.helper.Post(event.ExecutorToP2PEvent{
			Payload:  payload,
			Type:     NOTIFY_BROADCAST_SINGLE,
			Peers:    executor.status.syncFlag.SyncPeers,
		})
		return nil
	case NOTIFY_SYNC_REPLICA:
		executor.logger.Debug("inform p2p sync replica")
		if len(message) != 0 {
			return InvalidParamsErr
		}
		chain, ok := message[0].(*types.Chain)
		if !ok {
			return InvalidParamsErr
		}
		payload, _ := proto.Marshal(chain)
		executor.helper.Post(event.ExecutorToP2PEvent{
			Payload: payload,
			Type:    NOTIFY_SYNC_REPLICA,
		})
		return nil
	case NOTIFY_TRANSIT_BLOCK:
		executor.logger.Debug("inform p2p to transit commited block")
		return nil

	default:
		return NoDefinedCaseErr
	}
	return nil
}

