package executor

import "hyperchain/core/types"

type NVP interface {
	ReceiveBlock(*types.Block)
}

type NVPContext interface {

}

type NVPImpl struct {
	ctx    NVPContext
}

type NVPContextImpl struct {

}

func (nvp *NVPImpl) ReceiveBlock(*types.Block) {

}


