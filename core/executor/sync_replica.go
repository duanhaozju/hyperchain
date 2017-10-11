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
	"github.com/golang/protobuf/proto"
	"hyperchain/common"
	edb "hyperchain/core/ledger/db_utils"
	"hyperchain/core/types"
	"time"
)

func (executor *Executor) syncReplica() {
	if executor.conf.GetSyncReplicaEnable() {
		executor.sendReplicaInfo()
	}
}

func (executor *Executor) sendReplicaInfo() {
	interval := executor.conf.GetSyncReplicaInterval()
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-executor.context.exit:
			executor.logger.Notice("replica sync backend exit")
			return
		case <-ticker.C:
			executor.informP2P(NOTIFY_SYNC_REPLICA, edb.GetChainCopy(executor.namespace))
		}
	}
}

func (executor *Executor) ReceiveReplicaInfo(payload []byte) {
	info := &types.ReplicaInfo{}
	proto.Unmarshal(payload, info)
	if string(info.Namespace) != executor.namespace {
		return
	}
	executor.logger.Noticef("[Namespace = %s] receive replica info, ip %s, port %d, chain height %d, latest block hash %s",
		string(info.Namespace), string(info.Ip), info.Port, info.Chain.Height, common.Bytes2Hex(info.Chain.LatestBlockHash))
	executor.addToReplicaCache(string(info.Ip), info.Port, info.Chain)
}
