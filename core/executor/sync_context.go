package executor

import (
	edb "hyperchain/core/ledger/db_utils"
	"hyperchain/manager/event"

	"github.com/pkg/errors"
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

/*
	chain synchronization context
	// TODO merge other flags related to `sync` here
*/
type ChainSyncContext struct {
	FullPeers []uint64 // peers list which contains all required blocks. experiential this type peer has
	// higher priority to make chain synchronization
	PartPeers []PartPeer // peers list which just has a part of required blocks. If this type peer be chosen as target
	// chain synchronization must through world state transition
	CurrentPeer    uint64 // current sync target peer id
	CurrentGenesis uint64 // target peer's genesis tag
	ResendMode     uint32 // resend mode. All include (1) block (2) world state req (3) world state piece

	UpdateGenesis      bool   // whether world state transition is necessary. If target peer chose from `partpeer` collections, this flag is `True`
	GenesisTranstioned bool   // whether world state transition has finished
	Handshaked         bool   // whether world state transition handshake has received
	ReceiveAll         bool   // whether all content has received
	WorldStatePieceId  uint64 // represent current demand world state piece id

	// WS related
	hs     *WsHandshake
	wsHome string
}

func NewChainSyncContext(namespace string, event event.ChainSyncReqEvent) *ChainSyncContext {
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
	return &ChainSyncContext{
		FullPeers:     fullPeers,
		PartPeers:     partPeers,
		UpdateGenesis: updateGenesis,
	}
}

func (ctx *ChainSyncContext) GetFullPeersId() []uint64 {
	return ctx.FullPeers
}

func (ctx *ChainSyncContext) GetPartPeersId() []uint64 {
	var ids []uint64
	for _, p := range ctx.PartPeers {
		ids = append(ids, p.Id)
	}
	return ids
}

func (ctx *ChainSyncContext) SetCurrentPeer(id uint64) {
	ctx.CurrentPeer = id
}

func (ctx *ChainSyncContext) GetCurrentPeer() uint64 {
	return ctx.CurrentPeer
}

func (ctx *ChainSyncContext) GetCurrentGenesis() (error, uint64) {
	for _, id := range ctx.FullPeers {
		if ctx.CurrentPeer == id {
			return nil, 0
		}
	}
	for _, p := range ctx.PartPeers {
		if ctx.CurrentPeer == p.Id {
			return nil, p.Genesis
		}
	}
	return errors.New("no genesis exist"), 0
}

func (ctx *ChainSyncContext) GetGenesis(pid uint64) (error, uint64) {
	for _, id := range ctx.FullPeers {
		if pid == id {
			return nil, 0
		}
	}
	for _, p := range ctx.PartPeers {
		if pid == p.Id {
			return nil, p.Genesis
		}
	}
	return errors.New("no genesis exist"), 0
}

func (ctx *ChainSyncContext) SetResendMode(mode uint32) {
	atomic.StoreUint32(&ctx.ResendMode, mode)
}

func (ctx *ChainSyncContext) GetResendMode() uint32 {
	return atomic.LoadUint32(&ctx.ResendMode)
}

func (ctx *ChainSyncContext) SetWsHome(p string) {
	ctx.wsHome = p
}

func (ctx *ChainSyncContext) GetWsHome() string {
	return ctx.wsHome
}

func (ctx *ChainSyncContext) SetTransitioned() {
	ctx.GenesisTranstioned = true
}

func (ctx *ChainSyncContext) GetTranstioned() bool {
	return ctx.GenesisTranstioned
}

func (ctx *ChainSyncContext) RecordWsHandshake(hs *WsHandshake) {
	ctx.hs = hs
	ctx.Handshaked = true
	ctx.SetResendMode(ResendMode_WorldState_Piece)
}

func (ctx *ChainSyncContext) SetWsId(id uint64) {
	atomic.StoreUint64(&ctx.WorldStatePieceId, id)
}

func (ctx *ChainSyncContext) GetWsId() uint64 {
	return atomic.LoadUint64(&ctx.WorldStatePieceId)
}
