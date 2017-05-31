package executor

import (
	edb "hyperchain/core/db_utils"
	"hyperchain/manager/event"

	"github.com/pkg/errors"
	"fmt"
)

type PartPeer struct {
	Id      uint64
	Genesis uint64
}

type ChainSyncContext struct {
	FullPeers       []uint64          // peers list which contains all required blocks
	PartPeers       []PartPeer        // peers list which just has a part of required blocks
	UpdateGenesis   bool              // whether transit genesis status via network data

	CurrentPeer     uint64
	CurrentGenesis  uint64
}

func NewChainSyncContext(namespace string, event event.ChainSyncReqEvent) *ChainSyncContext {
	var fullPeers []uint64
	var partPeers []PartPeer
	curHeight := edb.GetHeightOfChain(namespace)
	for _, r := range event.Replicas {
		if r.Genesis <= curHeight {
			fmt.Println("append fullpeer: ", r.Id)
			fullPeers = append(fullPeers, r.Id)
		} else {
			fmt.Println("append partpeer: ", r.Id)
			partPeers = append(partPeers, PartPeer{
				Id:        r.Id,
				Genesis:   r.Genesis,
			})
		}
	}
	return &ChainSyncContext{
		FullPeers:     fullPeers,
		PartPeers:     partPeers,
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


