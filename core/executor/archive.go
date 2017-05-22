package executor

import (
	"hyperchain/manager/event"
	edb "hyperchain/core/db_utils"
	"github.com/op/go-logging"
	"sync"
)

const LatestBlockNumber uint64 = 0

const InvalidSnapshotReqErr = "invalid snapshot request"
const MakeSnapshotFailedErr = "make snapshot failed"
const EmptyMessage = ""

// snapshot service's entry point
func (executor *Executor) Snapshot(ev event.SnapshotEvent) {
	executor.snapshotReg.Snapshot(ev)
}


type SnapshotRegistry struct {
	namespace  string
	rq         map[uint64]event.SnapshotEvent
	rqLock     sync.RWMutex
	logger     *logging.Logger
	newBlockC  chan uint64
	exitC      chan struct{}
	executor   *Executor
}

func NewSnapshotRegistry(namespace string, logger *logging.Logger, executor *Executor) *SnapshotRegistry {
	return &SnapshotRegistry{
		namespace:  namespace,
		rq:         make(map[uint64]event.SnapshotEvent),
		logger:     logger,
		newBlockC:  make(chan uint64),
		exitC:      make(chan struct{}),
		executor:   executor,
	}
}

/*
	External functions
 */
func (registry *SnapshotRegistry) Start() error {
	go registry.listen()
	return nil
}

func (registry *SnapshotRegistry) Stop() error {
	registry.exit()
	return nil
}

func (registry *SnapshotRegistry) Snapshot(event event.SnapshotEvent) {
	if !registry.checkRequest(event) {
		registry.feedback(false, event, InvalidSnapshotReqErr)
	} else {
		registry.addRequest(event)
	}
}


/*
	Internal functions
 */
func (registry *SnapshotRegistry) addRequest(event event.SnapshotEvent) {
	registry.rqLock.Lock()
	defer registry.rqLock.Unlock()
	registry.logger.Noticef("receive snapshot event at (block #%d)", event.BlockNumber)
	registry.rq[event.BlockNumber] = event
}

func (registry *SnapshotRegistry) listen() {
	for {
		select {
		case <- registry.exitC:
			return
		case n := <- registry.newBlockC:
			registry.handle(n)
		}
	}
}

func (registry *SnapshotRegistry) handle(number uint64) {
	registry.rqLock.Lock()
	defer registry.rqLock.Unlock()
	if ev, existed := registry.rq[number]; existed == true {
		// TODO archive logic here
		registry.logger.Noticef("start to snapshot at (block #%d) for filter (%s)", number, ev.FilterId)
		if err := registry.makeSnapshot(ev.FilterId); err != nil {
			registry.feedback(false, ev, MakeSnapshotFailedErr)
		} else {
			registry.feedback(true, ev, EmptyMessage)
		}
		delete(registry.rq, number)
	}
}

func (registry *SnapshotRegistry) makeSnapshot(filterId string) error {
	registry.executor.setSuspend(IDENTIFIER_COMMIT)
	defer registry.executor.unsetSuspend(IDENTIFIER_COMMIT)

	fId := registry.snapshotId(filterId)
	fName := "snapshots/" + fId
	if err := registry.executor.db.Backup(fName); err != nil {
		return err
	}

	return nil
}



func (registry *SnapshotRegistry) checkRequest(event event.SnapshotEvent) bool {
	if event.BlockNumber == LatestBlockNumber {
		return true
	}
	chainHeight := edb.GetHeightOfChain(registry.namespace)
	if event.BlockNumber < chainHeight {
		return false
	}
	return true
}

func (registry *SnapshotRegistry) notifyNewBlock(number uint64) {
	registry.newBlockC <- number
}

func (registry *SnapshotRegistry) feedback(isSuccess bool, ev event.SnapshotEvent, message string) {
	registry.executor.sendFilterEvent(FILTER_SNAPSHOT_RESULT, isSuccess, ev.FilterId, message)
}

func (registry *SnapshotRegistry) exit() {
	registry.exitC <- struct {}{}
}

func (registry *SnapshotRegistry) snapshotId(filterId string) string {
	return "SNAPSHOT_" + filterId
}


