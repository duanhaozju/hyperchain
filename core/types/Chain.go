package types

import (
	"encoding/hex"
	"strconv"
)

type Chain struct {
	LastestBlockHash string
	Height int
}

func (chain Chain) String()string{
	hash := hex.EncodeToString([]byte(chain.LastestBlockHash))
	return hash + "&" + strconv.Itoa(chain.Height)
}