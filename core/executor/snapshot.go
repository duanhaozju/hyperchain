package executor

import (
	"hyperchain/manager/event"
	edb "hyperchain/core/db_utils"
	"github.com/op/go-logging"
	"sync"
	"hyperchain/hyperdb"
	"path"
	"time"
	"hyperchain/common"
	cmd "os/exec"
	"path/filepath"
)

const LatestBlockNumber uint64 = 0

const InvalidSnapshotReqErr   = "invalid snapshot request"
const MakeSnapshotFailedErr   = "make snapshot failed"
const SnapshotNotExistErr     = "snapshot doesn't exist"
const DeleteSnapshotErr       = "delete snapshot failed"

const EmptyMessage = ""



// snapshot service's entry point
func (executor *Executor) Snapshot(ev event.SnapshotEvent) {
	executor.snapshotReg.Snapshot(ev)
}

func (executor *Executor) DeleteSnapshot(ev event.DeleteSnapshotEvent) {
	executor.snapshotReg.DeleteSnapshot(ev)
}

// snapshot manager
type SnapshotRegistry struct {
	namespace  string
	rq         map[uint64]event.SnapshotEvent
	rqLock     sync.RWMutex
	logger     *logging.Logger
	newBlockC  chan uint64
	exitC      chan struct{}
	executor   *Executor
	mu         sync.Mutex
	rwc        common.ManifestRWC
}

func NewSnapshotRegistry(namespace string, logger *logging.Logger, executor *Executor) *SnapshotRegistry {
	return &SnapshotRegistry{
		namespace:  namespace,
		rq:         make(map[uint64]event.SnapshotEvent),
		logger:     logger,
		newBlockC:  make(chan uint64),
		exitC:      make(chan struct{}),
		executor:   executor,
		rwc:        common.NewManifestHandler(executor.GetManifestPath()),
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
		registry.feedback(FILTER_SNAPSHOT_RESULT, false, event.FilterId, InvalidSnapshotReqErr)
	} else {
		if registry.isExecuteImmediate(event) {
			registry.executeImmediately(event)
		} else {
			registry.addRequest(event)
		}
	}
}

func (registry *SnapshotRegistry) DeleteSnapshot(event event.DeleteSnapshotEvent) {
	// TODO add request check
	// TODO @Rongjialei fix me
	if !registry.rwc.Contain(event.FilterId) {
		registry.feedback(FILTER_DELETE_SNAPSHOT, false, event.FilterId, SnapshotNotExistErr)
	} else {
		if err := registry.deleteSnapshot(event.FilterId); err != nil {
			registry.feedback(FILTER_DELETE_SNAPSHOT, false, event.FilterId, DeleteSnapshotErr)
		}
		if err := registry.rwc.Delete(event.FilterId); err != nil {
			registry.feedback(FILTER_DELETE_SNAPSHOT, false, event.FilterId, DeleteSnapshotErr)
		}
	}
}


/*
	Internal functions
 */
func (registry *SnapshotRegistry) executeImmediately(ev event.SnapshotEvent) {
	height := edb.GetHeightOfChain(registry.namespace)
	if err := registry.makeSnapshot(ev.FilterId, height); err != nil {
		registry.logger.Noticef("snapshot at (block #%d) for filter (%s) failed", height, ev.FilterId)
		registry.feedback(FILTER_SNAPSHOT_RESULT, false, ev.FilterId, MakeSnapshotFailedErr)
	} else {
		registry.logger.Noticef("snapshot at (block #%d) for filter (%s) success", height, ev.FilterId)
		registry.feedback(FILTER_SNAPSHOT_RESULT, true, ev.FilterId, EmptyMessage)
	}
}

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
		registry.logger.Noticef("start to snapshot at (block #%d) for filter (%s)", number, ev.FilterId)
		if err := registry.makeSnapshot(ev.FilterId, number); err != nil {
			registry.logger.Noticef("snapshot at (block #%d) for filter (%s) failed", number, ev.FilterId)
			registry.feedback(FILTER_SNAPSHOT_RESULT, false, ev.FilterId, MakeSnapshotFailedErr)
		} else {
			registry.logger.Noticef("snapshot at (block #%d) for filter (%s) success", number, ev.FilterId)
			registry.feedback(FILTER_SNAPSHOT_RESULT, true, ev.FilterId, EmptyMessage)
		}
		delete(registry.rq, number)
	}
}

func (registry *SnapshotRegistry) makeSnapshot(filterId string, number uint64) error {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	registry.executor.setSuspend(IDENTIFIER_COMMIT)
	defer registry.executor.unsetSuspend(IDENTIFIER_COMMIT)
	if err := registry.duplicate(filterId); err != nil {
		return err
	}
	if err := registry.removeImpurity(filterId, number); err != nil {
		return err
	}
	if err := registry.compress(filterId); err != nil {
		return err
	}
	if err := registry.manifest(filterId, number); err != nil {
		return err
	}
	return nil
}

func (registry *SnapshotRegistry) duplicate(filterId string) error {
	conf := registry.executor.conf
	fId := registry.snapshotId(filterId)
	sPath := registry.snapshotPath(hyperdb.GetDatabaseHome(conf), fId)
	if err := registry.executor.db.Backup(sPath); err != nil {
		return err
	}
	return nil
}

func (registry *SnapshotRegistry) removeImpurity(filterId string, number uint64) error {
	conf := registry.executor.conf
	localDb, err := hyperdb.NewDatabase(conf, path.Join("snapshots", registry.snapshotId(filterId)), hyperdb.GetDatabaseType(conf))
	if err != nil {
		return err
	}
	defer localDb.Close()
	batch := localDb.NewBatch()
	var i uint64 = 0
	for ; i <= number; i += 1 {
		// TODO read block from backup is the best
		// TODO @Rongjialei fix me
		block, err := edb.GetBlockByNumber(registry.executor.namespace, i)
		if err != nil {
			registry.logger.Errorf("miss block %d ,error msg %s", i, err.Error())
			continue
		}

		for _, tx := range block.Transactions {
			if err := edb.DeleteTransaction(batch, tx.GetHash().Bytes(), false, false); err != nil {
				registry.logger.Errorf("[Namespace = %s] delete useless tx in block %d failed, error msg %s", registry.executor.namespace, i, err.Error())
			}
			if err := edb.DeleteReceipt(batch, tx.GetHash().Bytes(), false, false); err != nil {
				registry.logger.Errorf("[Namespace = %s] delete useless receipt in block %d failed, error msg %s", registry.executor.namespace, i, err.Error())
			}
		}
		// delete block
		if err := edb.DeleteBlockByNum(registry.executor.namespace, batch, i, false, false); err != nil {
			registry.logger.Errorf("[Namespace = %s] delete useless block %d failed, error msg %s", registry.executor.namespace, i, err.Error())
		}
	}
	if err := edb.DeleteAllDiscardTransaction(localDb, batch, false, false); err != nil {
		registry.logger.Errorf("[Namespace = %s] delete useless invalid records failed, error msg %s", registry.executor.namespace, err.Error())
	}
	if err := edb.RemoveChain(batch, false, false); err != nil {
		registry.logger.Errorf("[Namespace = %s] delete useless chain data , error msg %s", registry.executor.namespace, err.Error())
	}
	if err := edb.DeleteAllJournals(localDb, batch, false, false); err != nil {
		registry.logger.Errorf("[Namespace = %s] delete useless journal failed, error msg %s", registry.executor.namespace, err.Error())
	}
	if err := batch.Write(); err != nil {
		registry.logger.Errorf("[Namespace = %s] flush batch for deletion, error msg %s", registry.executor.namespace, err.Error())
	}
	registry.logger.Noticef("remove useless from snapshot %s success.", filterId)
	return nil
}

func (registry *SnapshotRegistry) compress(filterId string) error {
	path := registry.snapshotPath(hyperdb.GetDatabaseHome(registry.executor.conf), registry.snapshotId(filterId))
	localCmd := cmd.Command("tar", "-C" , filepath.Dir(path), "-czvf", path + ".tar.gz", filepath.Base(path))
	if err := localCmd.Run(); err != nil {
		return err
	}
	localCmd = cmd.Command("rm", "-rf", path)
	if err := localCmd.Run(); err != nil {
		return err
	}
	return nil
}

func (registry *SnapshotRegistry) manifest(filterId string, number uint64) error {
	blk, err := edb.GetBlockByNumber(registry.namespace, number)
	if err != nil {
		return err
	}
	d := time.Unix(time.Now().Unix(), 0).Format("2006-01-02-15:04:05")
	manifest := common.Manifest{
		Height:      number,
		FilterId:    filterId,
		MerkleRoot:  common.Bytes2Hex(blk.MerkleRoot),
		Date:        d,
	}
	if err := registry.rwc.Write(manifest); err != nil {
		return err
	}
	return nil
}

func (registry *SnapshotRegistry) deleteSnapshot(filterId string) error {
	conf := registry.executor.conf
	fId := registry.snapshotId(filterId)
	sPath := registry.snapshotPath(hyperdb.GetDatabaseHome(conf), fId)
	localCmd := cmd.Command("rm", "-rf", sPath + ".tar.gz")
	if err := localCmd.Run(); err != nil {
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

func (registry *SnapshotRegistry) isExecuteImmediate(event event.SnapshotEvent) bool {
	return event.BlockNumber == LatestBlockNumber
}

func (registry *SnapshotRegistry) notifyNewBlock(number uint64) {
	registry.newBlockC <- number
}

func (registry *SnapshotRegistry) feedback(t int, isSuccess bool, filterId string, message string) {
	registry.executor.sendFilterEvent(t, isSuccess, filterId, message)
}

func (registry *SnapshotRegistry) exit() {
	registry.exitC <- struct {}{}
}

func (registry *SnapshotRegistry) snapshotId(filterId string) string {
	return "SNAPSHOT_" + filterId
}

func (registry *SnapshotRegistry) snapshotPath(base string, filterID string) string {
	return path.Join(base, "snapshots", filterID)
}




