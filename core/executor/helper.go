package executor
import (
	"hyperchain/manager/event"
	"hyperchain/manager/protos"
	"github.com/golang/protobuf/proto"
	edb "hyperchain/core/db_utils"
	"hyperchain/core/types"
	"hyperchain/core/vm"
)
type Helper struct {
	innerMux        *event.TypeMux
	externalMux     *event.TypeMux
}

func NewHelper(innerMux *event.TypeMux, externalMux *event.TypeMux) *Helper {
	return &Helper{
		innerMux:       innerMux,
		externalMux:    externalMux,
	}
}

func (helper *Helper) PostInner(ev interface{}) {
	helper.innerMux.Post(ev)
}

func (helper *Helper) PostExternal(ev interface{}) {
	helper.externalMux.Post(ev)
}

// informConsensus - communicate with consensus module.
func (executor *Executor) informConsensus(informType int, message interface{}) error {
	switch informType {
	case NOTIFY_REMOVE_CACHE:
		executor.logger.Debug("inform consenus remove cache")
		msg := message.(protos.RemoveCache)
		executor.helper.PostInner(event.ExecutorToConsensusEvent{
			Payload: msg,
			Type:    NOTIFY_REMOVE_CACHE,
		})
	case NOTIFY_VALIDATION_RES:
		executor.logger.Debugf("[Namespace = %s] inform consenus validation result", executor.namespace)
		msg := message.(protos.ValidatedTxs)
		executor.helper.PostInner(event.ExecutorToConsensusEvent{
			Payload: msg,
			Type:    NOTIFY_VALIDATION_RES,
		})
	case NOTIFY_VC_DONE:
		executor.logger.Debug("inform consenus vc done")
		msg := message.(protos.VcResetDone)
		executor.helper.PostInner(event.ExecutorToConsensusEvent{
			Payload: msg,
			Type:    NOTIFY_VC_DONE,
		})
	case NOTIFY_SYNC_DONE:
		executor.logger.Debug("inform consenus sync done")
		msg := message.(protos.StateUpdatedMessage)
		executor.helper.PostInner(event.ExecutorToConsensusEvent{
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
		required := ChainSyncRequest{
			RequiredNumber: message[0].(uint64),
			CurrentNumber:  message[1].(uint64),
			PeerId:         executor.status.syncFlag.LocalId,
		}
		payload, err := proto.Marshal(&required)
		if err != nil {
			executor.logger.Errorf("sync chain request marshal message failed")
			return err
		}
		executor.helper.PostInner(event.ExecutorToP2PEvent{
			Payload: payload,
			Peers:   []uint64{message[2].(uint64)},
			Type:    NOTIFY_BROADCAST_DEMAND,
		})
		return nil
	case NOTIFY_UNICAST_BLOCK:
		executor.logger.Debug("inform p2p unicast block")
		id := message[0].(uint64)
		peerId := message[1].(uint64)
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
		executor.helper.PostInner(event.ExecutorToP2PEvent{
			Payload: payload,
			Type:    NOTIFY_UNICAST_BLOCK,
			Peers:   []uint64{peerId},
		})
		return nil
	case NOTIFY_UNICAST_INVALID:
		executor.logger.Debug("inform p2p unicast invalid tx")
		r := message[0].(*types.InvalidTransactionRecord)
		payload, err := proto.Marshal(r)
		if err != nil {
			executor.logger.Error("marshal invalid record error")
			return err
		}
		executor.helper.PostInner(event.ExecutorToP2PEvent{
			Payload: payload,
			Type:    NOTIFY_UNICAST_INVALID,
			Peers:   []uint64{r.Tx.Id},
		})
		return nil
	case NOTIFY_BROADCAST_SINGLE:
		executor.logger.Debug("inform p2p broadcast single demand")
		id := message[0].(uint64)
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
		executor.helper.PostInner(event.ExecutorToP2PEvent{
			Payload:  payload,
			Type:     NOTIFY_BROADCAST_SINGLE,
			Peers:    executor.status.syncFlag.SyncPeers,
		})
		return nil
	case NOTIFY_SYNC_REPLICA:
		executor.logger.Debug("inform p2p sync replica")
		payload, _ := proto.Marshal(message[0].(*types.Chain))
		executor.helper.PostInner(event.ExecutorToP2PEvent{
			Payload: payload,
			Type:    NOTIFY_SYNC_REPLICA,
		})
		return nil
	default:
		return NoDefinedCaseErr
	}
	return nil
}

func (executor *Executor) sendFilterEvent(informType int, message ...interface{}) error {
	switch informType {
	case FILTER_NEW_BLOCK:
		if len(message) != 1 {
			return InvalidParams
		}
		blk, ok := message[0].(*types.Block)
		if ok == false {
			return InvalidParams
		}
		executor.helper.PostExternal(event.FilterNewBlockEvent{blk})
		return nil
	case FILTER_NEW_LOG:
		if len(message) != 1 {
			return InvalidParams
		}
		logs, ok := message[0].([]*vm.Log)
		if ok == false {
			return InvalidParams
		}
		executor.helper.PostExternal(event.FilterNewLogEvent{logs})
		return nil
	case FILTER_SNAPSHOT_RESULT:
		if len(message) != 3 {
			return InvalidParams
		}
		isSuccess, ok := message[0].(bool)
		if ok == false {
			return InvalidParams
		}
		filterId, ok := message[1].(string)
		if ok == false {
			return InvalidParams
		}
		msg, ok := message[2].(string)
		if ok == false {
			return InvalidParams
		}
		executor.helper.PostExternal(event.FilterArchiveResultEvent{
			FilterId:   filterId,
			Success:    isSuccess,
			Message:    msg,
		})
		return nil
	default:
		return NoDefinedCaseErr
	}
}

