// Copyright 2016-2017 Hyperchain Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package executor

import (
	"github.com/golang/protobuf/proto"
	er "hyperchain/core/errors"
	edb "hyperchain/core/ledger/chain"
	"hyperchain/core/types"
	"hyperchain/manager/event"
	"hyperchain/manager/protos"
	"reflect"
    "hyperchain/common/service"
    pb "hyperchain/common/protos"
    "hyperchain/p2p/msg"
)

// Communication mux implementation
type Helper struct {
	innerMux    *event.TypeMux // system internal mux
	externalMux *event.TypeMux // subscription system mux
	client      *service.ServiceClient
}

func NewHelper(innerMux *event.TypeMux, externalMux *event.TypeMux, client *service.ServiceClient) *Helper {
	return &Helper{
		innerMux:    innerMux,
		externalMux: externalMux,
		client:      client,
	}
}

func (helper *Helper) handlePost(ev interface{}) *pb.IMessage {
    msg := &pb.IMessage{
        Type: pb.Type_DISPATCH,
        From: pb.FROM_EXECUTOR,
    }
    switch ev.(type) {
    case event.ExecutorToConsensusEvent:
        msg.Event = pb.Event_ExecutorToConsensusEvent
        mv, err := proto.Marshal(ev.(*event.ExecutorToConsensusEvent))
        if err != nil {
            return nil
        }
        msg.Payload = mv
    case event.ExecutorToP2PEvent:
        msg.Event = pb.Event_ExecutorToP2PEvent
        mv, err := proto.Marshal(ev.(*event.ExecutorToP2PEvent))
        if err != nil {
            return nil
        }
        msg.Payload = mv
    }
}

// PostInner post event to inner event mux
func (helper *Helper) PostInner(ev interface{}) {
    msg := helper.handlePost(ev)
    helper.client.Send(msg)
    return
	helper.innerMux.Post(ev)
}

// PostExternal post event to outer event mux
func (helper *Helper) PostExternal(ev interface{}) {
    msg := helper.handlePost(ev)
    helper.client.Send(msg)
    return
	helper.externalMux.Post(ev)
}

// checkParams the checker of the parameters, check whether the parameters are satisfied
// both in param number and param type.
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
	case NOTIFY_VALIDATION_RES:
		executor.logger.Debugf("[Namespace = %s] inform consenus validation result", executor.namespace)
		msg, ok := message.(protos.ValidatedTxs)
		if !ok {
			return er.InvalidParamsErr
		}

		payload, err := proto.Marshal(&msg)
		if err != nil {
			return err
		}

		executor.helper.PostInner(event.ExecutorToConsensusEvent{
			Payload: payload,
			Type:    NOTIFY_VALIDATION_RES,
		})
	case NOTIFY_VC_DONE:
		executor.logger.Debug("inform consenus vc done")
		msg, ok := message.(protos.VcResetDone)
		if !ok {
			return er.InvalidParamsErr
		}

		payload, err := proto.Marshal(&msg)
		if err != nil {
			return err
		}

		executor.helper.PostInner(event.ExecutorToConsensusEvent{
			Payload: payload,
			Type:    NOTIFY_VC_DONE,
		})
	case NOTIFY_SYNC_DONE:
		executor.logger.Debug("inform consenus sync done")
		msg, ok := message.(protos.StateUpdatedMessage)
		if !ok {
			return er.InvalidParamsErr
		}

		payload, err := proto.Marshal(&msg)
		if err != nil {
			return err
		}
		executor.helper.PostInner(event.ExecutorToConsensusEvent{
			Payload: payload,
			Type:    NOTIFY_SYNC_DONE,
		})
	default:
		return er.NoDefinedCaseErr
	}
	return nil
}

// informP2P - communicate with p2p module.
func (executor *Executor) informP2P(informType int, message ...interface{}) error {
	switch informType {
	case NOTIFY_BROADCAST_DEMAND:
		// Broadcast state update demand request.
		// Include the params:
		// (1) Target block number
		// (2) Current chain height
		// (3) Current peer identification
		executor.logger.Debug("inform p2p broadcast demand")
		if !checkParams([]reflect.Kind{reflect.Uint64, reflect.Uint64, reflect.Uint64}, message...) {
			return er.InvalidParamsErr
		}
		required := ChainSyncRequest{
			RequiredNumber: message[0].(uint64),
			CurrentNumber:  message[1].(uint64),
			PeerId:         executor.context.syncCtx.localId,
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
		// Unicast block data to the fetcher.
		executor.logger.Debug("inform p2p unicast block")
		if !checkParams([]reflect.Kind{reflect.Uint64, reflect.Uint64, reflect.String}, message...) {
			return er.InvalidParamsErr
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
			Payload:   payload,
			Type:      NOTIFY_UNICAST_BLOCK,
			Peers:     []uint64{message[1].(uint64)},
			PeersHash: []string{message[2].(string)},
		})
		return nil
	case NOTIFY_UNICAST_INVALID:
		// Unicast the invalid transaction to the original peer
		executor.logger.Debug("inform p2p unicast invalid tx")
		if len(message) != 1 {
			return er.InvalidParamsErr
		}
		r, ok := message[0].(*types.InvalidTransactionRecord)
		if !ok {
			return er.InvalidParamsErr
		}
		payload, err := proto.Marshal(r)
		if err != nil {
			executor.logger.Error("marshal invalid record error")
			return err
		}
		hash := r.Tx.GetNVPHash()
		if err != nil {
			executor.logger.Errorf("get nvp hash failde. Err Mag:%v.", err.Error())
		}
		executor.helper.PostInner(event.ExecutorToP2PEvent{
			Payload:   payload,
			Type:      NOTIFY_UNICAST_INVALID,
			Peers:     []uint64{r.Tx.Id},
			PeersHash: []string{hash},
		})
		return nil
	case NOTIFY_SYNC_REPLICA:
		// Broadcast current node status to all connected peers.
		// Content included:
		// (1) Chain height
		// (2) Chain latest block number
		// (3) Chain latest block hash
		// (4) etc
		executor.logger.Debug("inform p2p sync replica")
		if len(message) != 1 {
			return er.InvalidParamsErr
		}
		chain, ok := message[0].(*types.Chain)
		if !ok {
			return er.InvalidParamsErr
		}
		payload, _ := proto.Marshal(chain)
		executor.helper.PostInner(event.ExecutorToP2PEvent{
			Payload: payload,
			Type:    NOTIFY_SYNC_REPLICA,
		})
		return nil
	case NOTIFY_REQUEST_WORLD_STATE:
		// Unicast World state fetch request
		// Those params are included:
		// (1) Relative block height of the world state
		// (2) Request peer identification
		// (3) Target peer identification
		executor.logger.Debug("inform p2p sync world state")
		if !checkParams([]reflect.Kind{reflect.Uint64}, message...) {
			return er.InvalidParamsErr
		}
		request := &WsRequest{
			Target:      message[0].(uint64),
			InitiatorId: executor.context.syncCtx.localId,
			ReceiverId:  executor.context.syncCtx.getCurrentPeer(),
		}
		payload, err := proto.Marshal(request)
		if err != nil {
			return er.MarshalFailedErr
		}
		executor.helper.PostInner(event.ExecutorToP2PEvent{
			Payload: payload,
			Type:    NOTIFY_REQUEST_WORLD_STATE,
			Peers:   []uint64{executor.context.syncCtx.getCurrentPeer()},
		})
		return nil
	case NOTIFY_SEND_WORLD_STATE_HANDSHAKE:
		// Unicast world state transmission handshake message.
		// Those params are included:
		// (1) worldstate's size
		// (2) each packet size
		// (3) totally packet number
		// (4) relative block number
		executor.logger.Debug("inform p2p send world state handshake packet")
		if len(message) != 1 {
			return er.InvalidParamsErr
		}
		hs, ok := message[0].(*WsHandshake)
		if ok == false {
			return er.InvalidParamsErr
		}
		payload, err := proto.Marshal(hs)
		if err != nil {
			return er.MarshalFailedErr
		}
		executor.helper.PostInner(event.ExecutorToP2PEvent{
			Payload: payload,
			Type:    NOTIFY_SEND_WORLD_STATE_HANDSHAKE,
			Peers:   []uint64{hs.Ctx.ReceiverId},
		})
		return nil
	case NOTIFY_SEND_WS_ACK:
		// Unicast world state transmission ack message.
		// Those params are included:
		// (1) worldstate transmission session context
		// (2) received packet id
		// (3) received status
		// (4) extra message
		executor.logger.Debug("inform p2p send ws ack")
		if len(message) != 1 {
			return er.InvalidParamsErr
		}
		ack, ok := message[0].(*WsAck)
		if ok == false {
			return er.InvalidParamsErr
		}
		payload, err := proto.Marshal(ack)
		if err != nil {
			return er.MarshalFailedErr
		}
		executor.helper.PostInner(event.ExecutorToP2PEvent{
			Payload: payload,
			Type:    NOTIFY_SEND_WS_ACK,
			Peers:   []uint64{ack.Ctx.ReceiverId},
		})
		return nil
	case NOTIFY_SEND_WORLD_STATE:
		// Unicast world state transmission ws packet.
		// Those params are included:
		// (1) worldstate transmission session context
		// (2) worldstate slice content
		executor.logger.Debug("inform p2p sync world state")
		if len(message) != 1 {
			return er.InvalidParamsErr
		}
		ws, ok := message[0].(*Ws)
		if ok == false {
			return er.InvalidParamsErr
		}
		payload, err := proto.Marshal(ws)
		if err != nil {
			return er.MarshalFailedErr
		}
		executor.helper.PostInner(event.ExecutorToP2PEvent{
			Payload: payload,
			Type:    NOTIFY_SEND_WORLD_STATE,
			Peers:   []uint64{ws.Ctx.ReceiverId},
		})
	case NOTIFY_TRANSIT_BLOCK:
		// for nvp extension
		executor.logger.Debug("inform p2p to transit commited block")
		if len(message) != 1 {
			return er.InvalidParamsErr
		}
		block, ok := message[0].([]byte)
		if !ok {
			return er.InvalidParamsErr
		}
		executor.helper.PostInner(event.ExecutorToP2PEvent{
			Payload: block,
			Type:    NOTIFY_TRANSIT_BLOCK,
		})
		return nil
	case NOTIFY_NVP_SYNC:
		executor.logger.Debug("inform p2p to sync NVP")
		if !checkParams([]reflect.Kind{reflect.Uint64, reflect.Uint64}, message...) {
			return er.InvalidParamsErr
		}
		required := ChainSyncRequest{
			RequiredNumber: message[0].(uint64),
			CurrentNumber:  message[1].(uint64),
		}
		payload, err := proto.Marshal(&required)
		if err != nil {
			executor.logger.Errorf("sync chain request marshal message failed of NVP")
			return err
		}
		executor.helper.PostInner(event.ExecutorToP2PEvent{
			Payload: payload,
			Type:    NOTIFY_NVP_SYNC,
		})
		return nil
	default:
		return er.NoDefinedCaseErr
	}
	return nil
}

// sendFilterEvent - send event to subscription system.
func (executor *Executor) sendFilterEvent(informType int, message ...interface{}) error {
	switch informType {
	case FILTER_NEW_BLOCK:
		// NewBlock event
		if len(message) != 1 {
			return er.InvalidParamsErr
		}
		blk, ok := message[0].(*types.Block)
		if ok == false {
			return er.InvalidParamsErr
		}
		executor.helper.PostExternal(event.FilterNewBlockEvent{blk})
		return nil
	case FILTER_NEW_LOG:
		// New virtual machine log event
		if len(message) != 1 {
			return er.InvalidParamsErr
		}
		logs, ok := message[0].([]*types.Log)
		if ok == false {
			return er.InvalidParamsErr
		}
		executor.helper.PostExternal(event.FilterNewLogEvent{logs})
		return nil
	case FILTER_SNAPSHOT_RESULT:
		// Snapshot operation result event
		if !checkParams([]reflect.Kind{reflect.Bool, reflect.String, reflect.String}, message...) {
			return er.InvalidParamsErr
		}
		executor.helper.PostExternal(event.FilterArchive{
			Type:     event.FilterMakeSnapshot,
			Success:  message[0].(bool),
			FilterId: message[1].(string),
			Message:  message[2].(string),
		})
		return nil
	case FILTER_DELETE_SNAPSHOT:
		// Snapshot deletion result event
		if !checkParams([]reflect.Kind{reflect.Bool, reflect.String, reflect.String}, message...) {
			return er.InvalidParamsErr
		}
		executor.helper.PostExternal(event.FilterArchive{
			Type:     event.FilterDeleteSnapshot,
			Success:  message[0].(bool),
			FilterId: message[1].(string),
			Message:  message[2].(string),
		})
		return nil
	case FILTER_ARCHIVE:
		// archive operation result event
		if !checkParams([]reflect.Kind{reflect.Bool, reflect.String, reflect.String}, message...) {
			return er.InvalidParamsErr
		}
		executor.helper.PostExternal(event.FilterArchive{
			Type:     event.FilterDoArchive,
			Success:  message[0].(bool),
			FilterId: message[1].(string),
			Message:  message[2].(string),
		})
		return nil
	default:
		return er.NoDefinedCaseErr
	}
}
