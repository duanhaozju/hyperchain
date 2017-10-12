package executor

import (
	"github.com/golang/protobuf/proto"
	"hyperchain/common"
	edb "hyperchain/core/ledger/db_utils"
	"hyperchain/core/types"
	"time"
)

func (executor *Executor) syncReplica() {
	if executor.GetSyncReplicaEnable() {
		executor.sendReplicaInfo()
	}
}

func (executor *Executor) sendReplicaInfo() {
	interval := executor.GetSyncReplicaInterval()
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-executor.getExit(IDENTIFIER_REPLICA_SYNC):
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
