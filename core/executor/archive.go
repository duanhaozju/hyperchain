package executor

import (
	"hyperchain/manager/event"
	edb "hyperchain/core/db_utils"
	"github.com/op/go-logging"
	"sync"
)

const LatestBlockNumber uint64 = 0


type SnapshotRegistry struct {
	namespace  string
	rq         map[uint64]struct{}
	rqLock     sync.RWMutex
	logger     *logging.Logger
}

func NewSnapshotRegistry(namespace string, logger *logging.Logger) *SnapshotRegistry {
	return &SnapshotRegistry{
		namespace:  namespace,
		rq:         make(map[uint64]struct{}),
		logger:     logger,
	}
}


func (registry *SnapshotRegistry) addRequest(event event.SnapshotEvent) {
	registry.rqLock.Lock()
	defer registry.rqLock.Unlock()
	registry.rq[event.BlockNumber] = struct{}{}
}


func (registry *SnapshotRegistry) Snapshot(event event.SnapshotEvent) {
	if !registry.checkSnapshotReq(event) {
		// TODO make a notification event
		// TODO wait @Sammy finish event system
	} else {
		registry.addRequest(event)
	}

}

func (registry *SnapshotRegistry) checkSnapshotReq(event event.SnapshotEvent) bool {
	if event.BlockNumber == LatestBlockNumber {
		return true
	}
	chainHeight := edb.GetHeightOfChain(registry.namespace)
	if event.BlockNumber < chainHeight {
		return false
	}
	return true
}

func (registry *SnapshotRegistry) notify(isSuccess bool, message string, extra ...interface{}) {

}

