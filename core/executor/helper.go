package executor

import (
	"github.com/golang/protobuf/proto"
	edb "hyperchain/core/db_utils"
	"hyperchain/core/types"
	"hyperchain/core/vm"
	"hyperchain/manager/event"
	"hyperchain/manager/protos"
	"reflect"
)

type Helper struct {
	innerMux    *event.TypeMux
	externalMux *event.TypeMux
}

func NewHelper(innerMux *event.TypeMux, externalMux *event.TypeMux) *Helper {
	return &Helper{
		innerMux:    innerMux,
		externalMux: externalMux,
	}
}

func (helper *Helper) PostInner(ev interface{}) {
	helper.innerMux.Post(ev)
}

func (helper *Helper) PostExternal(ev interface{}) {
	helper.externalMux.Post(ev)
}

func checkParams(expect []reflect.Kind, params ...interface{}) bool {
	if len(expect) != len(params) {
		return false
	}
	for idx, typ := range expect {
		if typ != reflect.TypeOf(params[idx]).Kind() {
			return false
		}
	}
	return true
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
		executor.helper.PostInner(event.ExecutorToConsensusEvent{
			Payload: msg,
			Type:    NOTIFY_REMOVE_CACHE,
		})
	case NOTIFY_VALIDATION_RES:
		executor.logger.Debugf("[Namespace = %s] inform consenus validation result", executor.namespace)
		msg, ok := message.(protos.ValidatedTxs)
		if !ok {
			return InvalidParamsErr
		}
		executor.helper.PostInner(event.ExecutorToConsensusEvent{
			Payload: msg,
			Type:    NOTIFY_VALIDATION_RES,
		})
	case NOTIFY_VC_DONE:
		executor.logger.Debug("inform consenus vc done")
		msg, ok := message.(protos.VcResetDone)
		if !ok {
			return InvalidParamsErr
		}
		executor.helper.PostInner(event.ExecutorToConsensusEvent{
			Payload: msg,
			Type:    NOTIFY_VC_DONE,
		})
	case NOTIFY_SYNC_DONE:
		executor.logger.Debug("inform consenus sync done")
		msg, ok := message.(protos.StateUpdatedMessage)
		if !ok {
			return InvalidParamsErr
		}
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
		if !checkParams([]reflect.Kind{reflect.Uint64, reflect.Uint64, reflect.Uint64}, message...) {
			return InvalidParamsErr
		}
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
		if !checkParams([]reflect.Kind{reflect.Uint64, reflect.Uint64}, message...) {
			return InvalidParamsErr
		}
		block, err := edb.GetBlockByNumber(executor.namespace, message[0].(uint64))
		if err != nil {
			executor.logger.Errorf("no demand block number: %d", message[0].(uint64))
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
			Peers:   []uint64{message[1].(uint64)},
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
		executor.helper.PostInner(event.ExecutorToP2PEvent{
			Payload: payload,
			Type:    NOTIFY_UNICAST_INVALID,
			Peers:   []uint64{r.Tx.Id},
		})
		return nil
	case NOTIFY_BROADCAST_SINGLE:
		executor.logger.Debug("inform p2p broadcast single demand")
		if !checkParams([]reflect.Kind{reflect.Uint64}, message...) {
			return InvalidParamsErr
		}
		request := ChainSyncRequest{
			RequiredNumber: message[0].(uint64),
			CurrentNumber:  edb.GetHeightOfChain(executor.namespace),
			PeerId:         executor.status.syncFlag.LocalId,
		}
		payload, err := proto.Marshal(&request)
		if err != nil {
			executor.logger.Error("broadcast demand block, marshal message failed")
			return err
		}
		executor.helper.PostInner(event.ExecutorToP2PEvent{
			Payload: payload,
			Type:    NOTIFY_BROADCAST_SINGLE,
			Peers:   executor.status.syncFlag.SyncPeers,
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
		executor.helper.PostInner(event.ExecutorToP2PEvent{
			Payload: payload,
			Type:    NOTIFY_SYNC_REPLICA,
		})
		return nil
	case NOTIFY_REQUEST_WORLD_STATE:
		executor.logger.Notice("inform p2p sync world state")
		if !checkParams([]reflect.Kind{reflect.Uint64}, message...) {
			return InvalidParamsErr
		}
		request := &WsRequest{
			Target:      message[0].(uint64),
			InitiatorId: executor.status.syncFlag.LocalId,
			ReceiverId:  executor.status.syncCtx.GetCurrentPeer(),
		}
		payload, err := proto.Marshal(request)
		if err != nil {
			return MarshalFailedErr
		}
		executor.helper.PostInner(event.ExecutorToP2PEvent{
			Payload: payload,
			Type:    NOTIFY_REQUEST_WORLD_STATE,
			Peers:   []uint64{executor.status.syncCtx.GetCurrentPeer()},
		})
		return nil
	case NOTIFY_SEND_WORLD_STATE_HANDSHAKE:
		executor.logger.Notice("inform p2p send world state handshake packet")
		if len(message) != 1 {
			return InvalidParamsErr
		}
		hs, ok := message[0].(*WsHandshake)
		if ok == false {
			return InvalidParamsErr
		}
		payload, err := proto.Marshal(hs)
		if err != nil {
			return MarshalFailedErr
		}
		executor.helper.PostInner(event.ExecutorToP2PEvent{
			Payload: payload,
			Type:    NOTIFY_SEND_WORLD_STATE_HANDSHAKE,
			Peers:   []uint64{hs.Ctx.ReceiverId},
		})
		return nil
	case NOTIFY_SEND_WS_ACK:
		executor.logger.Notice("inform p2p send ws ack")
		if len(message) != 1 {
			return InvalidParamsErr
		}
		ack, ok := message[0].(*WsAck)
		if ok == false {
			return InvalidParamsErr
		}
		payload, err := proto.Marshal(ack)
		if err != nil {
			return MarshalFailedErr
		}
		executor.helper.PostInner(event.ExecutorToP2PEvent{
			Payload: payload,
			Type:    NOTIFY_SEND_WS_ACK,
			Peers:   []uint64{ack.Ctx.ReceiverId},
		})
		return nil
	case NOTIFY_SEND_WORLD_STATE:
		executor.logger.Notice("inform p2p sync world state")
		if len(message) != 1 {
			return InvalidParamsErr
		}
		ws, ok := message[0].(*Ws)
		if ok == false {
			return InvalidParamsErr
		}
		payload, err := proto.Marshal(ws)
		if err != nil {
			return MarshalFailedErr
		}
		executor.helper.PostInner(event.ExecutorToP2PEvent{
			Payload: payload,
			Type:    NOTIFY_SEND_WORLD_STATE,
			Peers:   []uint64{ws.Ctx.ReceiverId},
		})
	case NOTIFY_TRANSIT_BLOCK:
		executor.logger.Debug("inform p2p to transit commited block")
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
			return InvalidParamsErr
		}
		blk, ok := message[0].(*types.Block)
		if ok == false {
			return InvalidParamsErr
		}
		executor.helper.PostExternal(event.FilterNewBlockEvent{blk})
		return nil
	case FILTER_NEW_LOG:
		if len(message) != 1 {
			return InvalidParamsErr
		}
		logs, ok := message[0].([]*vm.Log)
		if ok == false {
			return InvalidParamsErr
		}
		executor.helper.PostExternal(event.FilterNewLogEvent{logs})
		return nil
	case FILTER_SNAPSHOT_RESULT:
		if !checkParams([]reflect.Kind{reflect.Bool, reflect.String, reflect.String}, message...) {
			return InvalidParamsErr
		}
		executor.helper.PostExternal(event.FilterSnapshotEvent{
			FilterId: message[1].(string),
			Success:  message[0].(bool),
			Message:  message[2].(string),
		})
		return nil
	case FILTER_DELETE_SNAPSHOT:
		if !checkParams([]reflect.Kind{reflect.Bool, reflect.String, reflect.String}, message...) {
			return InvalidParamsErr
		}
		executor.helper.PostExternal(event.FilterDeleteSnapshotEvent{
			FilterId: message[1].(string),
			Success:  message[0].(bool),
			Message:  message[2].(string),
		})
		return nil
	case FILTER_ARCHIVE:
		if !checkParams([]reflect.Kind{reflect.Bool, reflect.String, reflect.String}, message...) {
			return InvalidParamsErr
		}
		executor.helper.PostExternal(event.FilterArchive{
			FilterId: message[1].(string),
			Success:  message[0].(bool),
			Message:  message[2].(string),
		})
		return nil
	default:
		return NoDefinedCaseErr
	}
}
