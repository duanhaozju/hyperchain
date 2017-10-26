package version1_1

import (
	"encoding/json"
	"fmt"
	"github.com/hyperchain/hyperchain/common"
)

func (self *Chain) Encode() string {
	chainView := &ChainView{
		LatestBlockHash:  common.Bytes2Hex(self.LatestBlockHash),
		ParentBlockHash:  common.Bytes2Hex(self.ParentBlockHash),
		Height:           self.Height,
		RequiredBlockNum: self.RequiredBlockNum,
		RequireBlockHash: common.Bytes2Hex(self.RequireBlockHash),
		RecoveryNum:      self.RecoveryNum,
		CurrentTxSum:     self.CurrentTxSum,
	}
	res, err := json.MarshalIndent(chainView, "", "\t")
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	return string(res)
}
