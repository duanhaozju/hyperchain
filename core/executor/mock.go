// Copyright 2016-2017 Hyperchain Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package executor

import (
	edb "hyperchain/core/ledger/chain"
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
