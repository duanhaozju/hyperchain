package executor

import (
	"hyperchain/manager/event"
	edb "hyperchain/core/db_utils"
)

const LatestBlockNumber uint64 = 0

func (executor *Executor) Snapshot(event event.SnapshotEvent) {
	if !executor.checkSnapshotReq(event) {
		// TODO make a notification event
		// TODO wait @Sammy finish event system
	}

}

func (executor *Executor) checkSnapshotReq(event event.SnapshotEvent) bool {
	if event.BlockNumber == LatestBlockNumber {
		return true
	}
	chainHeight := edb.GetHeightOfChain(executor.namespace)
	if event.BlockNumber < chainHeight {
		return false
	}
	return true
}
