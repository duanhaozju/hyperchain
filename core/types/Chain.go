package types

import (
	"encoding/hex"
	"strconv"

	"hyperchain-alpha/common"
)

type Chain struct {
	LatestBlockHash common.Hash
	Height int
}

func (chain Chain) String()string{
	hash := hex.EncodeToString(chain.LatestBlockHash)
	return hash + "&" + strconv.Itoa(chain.Height)
}