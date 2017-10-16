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
	"bytes"
	"hyperchain/core/types"
)

var RemoveLessThan = func(key interface{}, iterKey interface{}) bool {
	id := key.(uint64)
	iterId := iterKey.(uint64)
	if id >= iterId {
		return true
	}
	return false
}

func VerifyBlockIntegrity(block *types.Block) bool {
	if bytes.Compare(block.BlockHash, block.Hash().Bytes()) == 0 {
		return true
	}
	return false
}
