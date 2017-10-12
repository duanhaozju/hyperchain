package executor

import (
	edb "hyperchain/core/ledger/db_utils"
	"strconv"
)

func (executor *Executor) MockTest_DirtyBlocks() {
	height := edb.GetHeightOfChain(executor.namespace)
	var i uint64
	for i = 1; i <= height; i += 1 {
		fakehash := "fakehash" + strconv.Itoa(int(i))
		blk, _ := edb.GetBlockByNumber(executor.namespace, i)
		blk.BlockHash = []byte(fakehash)
		edb.PersistBlock(executor.db.NewBatch(), blk, true, true)
	}
}
