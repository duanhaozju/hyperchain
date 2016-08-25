package types

import (
	"encoding/hex"
	"strconv"
)

type Chain struct {
	LatestBlockHash string
	Height int
}

func (chain Chain) String()string{
	hash := hex.EncodeToString([]byte(chain.LatestBlockHash))
	return hash + "&" + strconv.Itoa(chain.Height)
}