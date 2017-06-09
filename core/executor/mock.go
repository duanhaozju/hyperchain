package executor

import (
	edb "hyperchain/core/db_utils"
)

func (executor *Executor) MockTest_DirtyBlocks() {
	executor.logger.Critical("=========== Mock Test (dirty blocks) ============")
	height := edb.GetHeightOfChain(executor.namespace)
	var i uint64
	for i = 1; i <= height; i += 1 {
		blk, _ := edb.GetBlockByNumber(executor.namespace, i)
		blk.BlockHash = []byte("fakehash")
		edb.PersistBlock(executor.db.NewBatch(), blk, true, true)
	}
}

