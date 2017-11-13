package version1_3

import (
	"encoding/json"
	"fmt"
	"github.com/hyperchain/hyperchain/common"
)

func (self *Chain) Encode() string {
	chainView := &ChainView{
		Version:         string(self.Version),
		LatestBlockHash: common.Bytes2Hex(self.LatestBlockHash),
		ParentBlockHash: common.Bytes2Hex(self.ParentBlockHash),
		Height:          self.Height,
		Genesis:         self.Genesis,
		CurrentTxSum:    self.CurrentTxSum,
		Extra:           common.Bytes2Hex(self.Extra),
	}
	res, err := json.MarshalIndent(chainView, "", "\t")
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	return string(res)
}
