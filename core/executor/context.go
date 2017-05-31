package executor

type PartPeer struct {
	Id      uint64
	Genesis uint64
}

type ChainSyncContext struct {
	FullPeers       []uint64          // peers list which contains all required blocks
	PartPeers       []PartPeer        // peers list which just has a part of required blocks
	NewGenesis      uint64            // if `updateGenesis` is true, this field use to represent new genesis's height
	UpdateGenesis   bool              // whether transit genesis status via network data
}

func (ctx *ChainSyncContext) GetFullPeersId() []uint64 {
	return ctx.FullPeers
}

func (ctx *ChainSyncContext) GetPartPeersId() []uint64 {
	return ctx.PartPeers
}

func (ctx *ChainSyncContext) GetGenesis() uint64 {
	return ctx.NewGenesis
}
