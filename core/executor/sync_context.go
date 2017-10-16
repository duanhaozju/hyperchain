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
	edb "hyperchain/core/ledger/chain"
	"hyperchain/manager/event"

	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"hyperchain/common"
	"sync/atomic"
)

type PartPeer struct {
	Id      uint64
	Genesis uint64
}

const (
	ResendMode_Block uint32 = iota
	ResendMode_WorldState_Hs
	ResendMode_WorldState_Piece
	ResendMode_Nope
)

type syncFlag struct {
	syncDemandBlockNum  uint64 // latest demand block number
	syncDemandBlockHash []byte // latest demand block hash
	syncTarget          uint64 // target block number in this synchronization
	localId             uint64 // local node id
	tempDownstream      uint64 // current sync request low height(all required blocks will be split to
	// several batches to fetch, this field represents the low height of a request batch)
	latestUpstream   uint64 // latest sync request high height
	latestDownstream uint64 // latest sync request low height
}

// chain synchronization context, includes some status varints, target peers qos data.
type chainSyncContext struct {
	syncFlag
	fullPeers []uint64 // peers list which contains all required blocks. experiential this type peer has
	// higher priority to make chain synchronization
	partPeers []PartPeer // peers list which just has a part of required blocks. If this type peer be chosen as target
	// chain synchronization must through world state transition
	currentPeer    uint64 // current sync target peer id
	currentGenesis uint64 // target peer's genesis tag
	resendMode     uint32 // resend mode. Includes (1) block (2) world state req (3) world state piece three modes.

	updateGenesis      bool   // whether world state transition is necessary. If target peer chosen from `partpeer` collections, this flag is `True`
	genesisTranstioned bool   // whether world state transition has finished
	handshaked         bool   // whether world state transition handshake has received
	receiveAll         bool   // whether all content has received
	worldStatePieceId  uint64 // represent current demand world state piece id
	// WS related
	hs      *WsHandshake
	wsHome  string
	qosStat *QosStat // peer selector before send sync request, adhere `BEST PEER` algorithm
}

func newChainSyncContext(namespace string, event event.ChainSyncReqEvent, config *common.Config, logger *logging.Logger) *chainSyncContext {
	var fullPeers []uint64
	var partPeers []PartPeer
	curHeight := edb.GetHeightOfChain(namespace)
	target := event.TargetHeight
	for _, r := range event.Replicas {
		if r.Genesis <= curHeight {
			fullPeers = append(fullPeers, r.Id)
		} else if r.Genesis <= target {
			partPeers = append(partPeers, PartPeer{
				Id:      r.Id,
				Genesis: r.Genesis,
			})
		}
	}
	updateGenesis := (len(fullPeers) == 0)
	ctx := &chainSyncContext{
		fullPeers:     fullPeers,
		partPeers:     partPeers,
		updateGenesis: updateGenesis,
	}
	// pre-select a best peer
	ctx.qosStat = NewQos(ctx, config, namespace, logger)
	ctx.setCurrentPeer(ctx.qosStat.SelectPeer())

	// assign initial sync target
	ctx.syncTarget = event.TargetHeight
	ctx.syncDemandBlockNum = event.TargetHeight
	ctx.syncDemandBlockHash = event.TargetBlockHash
	ctx.tempDownstream = event.TargetHeight

	// assign target peer and local identification
	ctx.localId = event.Id

	return ctx
}

// update updates demand block number, related hash and target during the sync.
func (ctx *chainSyncContext) update(num uint64, hash []byte) {
	ctx.syncDemandBlockNum = num
	ctx.syncDemandBlockHash = hash
}

// recordRequest records current sync request's high height and low height.
func (ctx *chainSyncContext) recordRequest(upstream, downstream uint64) {
	atomic.StoreUint64(&ctx.latestUpstream, upstream)
	atomic.StoreUint64(&ctx.latestDownstream, downstream)
}

// getRequest returns latest recorded request.
func (ctx *chainSyncContext) getRequest() (uint64, uint64) {
	return atomic.LoadUint64(&ctx.latestUpstream), atomic.LoadUint64(&ctx.latestDownstream)
}

// setDownstream saves latest sync request down stream.
// return 0 if hasn't been set.
func (ctx *chainSyncContext) setDownstream(num uint64) {
	ctx.tempDownstream = num
}

// getDownstream gets latest sync request down stream.
func (ctx *chainSyncContext) getDownstream() uint64 {
	return ctx.tempDownstream
}

// getFullPeersId returns the whole full peer id list.
func (ctx *chainSyncContext) getFullPeersId() []uint64 {
	return ctx.fullPeers
}

// getFullPeersId returns the whole part peer id list.
func (ctx *chainSyncContext) getPartPeersId() []uint64 {
	var ids []uint64
	for _, p := range ctx.partPeers {
		ids = append(ids, p.Id)
	}
	return ids
}

// setCurrentPeer records given peer id as the chosen one.
func (ctx *chainSyncContext) setCurrentPeer(id uint64) {
	ctx.currentPeer = id
}

// getCurrentPeer returns current chosen target peer id.
func (ctx *chainSyncContext) getCurrentPeer() uint64 {
	return ctx.currentPeer
}

// getTargerGenesis returns the genesis block number during this synchronization.
func (ctx *chainSyncContext) getTargerGenesis() (error, uint64) {
	for _, id := range ctx.fullPeers {
		if ctx.currentPeer == id {
			return nil, 0
		}
	}
	for _, p := range ctx.partPeers {
		if ctx.currentPeer == p.Id {
			return nil, p.Genesis
		}
	}
	return errors.New("no genesis exist"), 0
}

func (ctx *chainSyncContext) setResendMode(mode uint32) {
	atomic.StoreUint32(&ctx.resendMode, mode)
}

func (ctx *chainSyncContext) getResendMode() uint32 {
	return atomic.LoadUint32(&ctx.resendMode)
}

func (ctx *chainSyncContext) setWsHome(p string) {
	ctx.wsHome = p
}

func (ctx *chainSyncContext) getWsHome() string {
	return ctx.wsHome
}

func (ctx *chainSyncContext) setTransitioned() {
	ctx.genesisTranstioned = true
}

func (ctx *chainSyncContext) getTranstioned() bool {
	return ctx.genesisTranstioned
}

// recordWsHandshake saves world state transition handshake data, converts resend mode
// to resend_ws_piece.
func (ctx *chainSyncContext) recordWsHandshake(hs *WsHandshake) {
	ctx.hs = hs
	ctx.handshaked = true
	ctx.setResendMode(ResendMode_WorldState_Piece)
}

// setWsId records latest received world state piece id by given one.
func (ctx *chainSyncContext) setWsId(id uint64) {
	atomic.StoreUint64(&ctx.worldStatePieceId, id)
}

// getWsId returns latest received world state piece id.
func (ctx *chainSyncContext) getWsId() uint64 {
	return atomic.LoadUint64(&ctx.worldStatePieceId)
}
