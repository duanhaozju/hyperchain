package executor

import (
	"github.com/op/go-logging"
	"hyperchain/common"
	cm "hyperchain/core/common"
	edb "hyperchain/core/db_utils"
	"hyperchain/core/hyperstate"
	"hyperchain/hyperdb"
	"hyperchain/manager/event"
	"os"
	cmd "os/exec"
	"path"
	"path/filepath"
	"sync"
	"time"
)

// snapshot service's entry point
func (executor *Executor) Snapshot(ev event.SnapshotEvent) {
	executor.snapshotReg.Snapshot(ev)
}

func (executor *Executor) DeleteSnapshot(ev event.DeleteSnapshotEvent) {
	executor.snapshotReg.DeleteSnapshot(ev)
}

// snapshot manager
type SnapshotRegistry struct {
	namespace string
	rq        map[uint64]event.SnapshotEvent
	rqLock    sync.RWMutex
	logger    *logging.Logger
	newBlockC chan uint64
	exitC     chan struct{}
	executor  *Executor
	mu        sync.Mutex
	rwc       common.ManifestRWC
}

func NewSnapshotRegistry(namespace string, logger *logging.Logger, executor *Executor) *SnapshotRegistry {
	return &SnapshotRegistry{
		namespace: namespace,
		rq:        make(map[uint64]event.SnapshotEvent),
		logger:    logger,
		newBlockC: make(chan uint64),
		exitC:     make(chan struct{}),
		executor:  executor,
		rwc:       common.NewManifestHandler(executor.GetManifestPath()),
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
	if !registry.checkSnapshotRequest(event) {
		registry.feedback(FILTER_SNAPSHOT_RESULT, false, event.FilterId, InvalidSnapshotReqErr)
	} else {
		if registry.isExecuteImmediate(event) {
			registry.executeImmediately(event)
		} else {
			registry.addRequest(event)
		}
	}
}

// DeleteSnapshot Remove the snapshot from the online database.
// Result can been received by subscription.
func (registry *SnapshotRegistry) DeleteSnapshot(event event.DeleteSnapshotEvent) {
	if !registry.checkDeletionRequest(event) {
		registry.feedback(FILTER_DELETE_SNAPSHOT, false, event.FilterId, InvalidDeletionReqErr)
	} else {
		if !registry.rwc.Contain(event.FilterId) {
			registry.feedback(FILTER_DELETE_SNAPSHOT, false, event.FilterId, SnapshotNotExistErr)
		} else {
			if err := registry.deleteSnapshot(event.FilterId); err != nil {
				registry.logger.Warningf("delete snapshot %s failed. reason %s", event.FilterId, err.Error())
				registry.feedback(FILTER_DELETE_SNAPSHOT, false, event.FilterId, DeleteSnapshotErr)
			}
			if err := registry.rwc.Delete(event.FilterId); err != nil {
				registry.feedback(FILTER_DELETE_SNAPSHOT, false, event.FilterId, DeleteSnapshotErr)
			}
		}
	}
}

func (registry *SnapshotRegistry) CompressSnapshot(filterId string) (error, int64) {
	return registry.compress(filterId)
}

func (registry *SnapshotRegistry) CompressedSnapshotPath(filterId string) string {
	return path.Join(hyperdb.GetDatabaseHome(registry.executor.conf), "snapshots", registry.snapshotId(filterId)+".tar.gz")
}

/*
	Internal functions
*/
func (registry *SnapshotRegistry) executeImmediately(ev event.SnapshotEvent) {
	height := edb.GetHeightOfChain(registry.namespace)
	if err := registry.makeSnapshot(ev.FilterId, height); err != nil {
		registry.logger.Noticef("snapshot at (block #%d) for filter (%s) failed, reason %s", height, ev.FilterId, err.Error())
		registry.feedback(FILTER_SNAPSHOT_RESULT, false, ev.FilterId, MakeSnapshotFailedErr)
	} else {
		registry.logger.Noticef("snapshot at (block #%d) for filter (%s) success", height, ev.FilterId)
		registry.feedback(FILTER_SNAPSHOT_RESULT, true, ev.FilterId, EmptyMessage)
	}
}

// addRequest save snapshot request into pending queue, wait for condition trigger.
func (registry *SnapshotRegistry) addRequest(event event.SnapshotEvent) {
	registry.rqLock.Lock()
	defer registry.rqLock.Unlock()
	if len(registry.rq) > MaxPendingSnapshotReq {
		registry.logger.Notice("too many pending snapshot request, dicard the new one")
	} else {
		registry.logger.Noticef("receive snapshot event at (block #%d)", event.BlockNumber)
		registry.rq[event.BlockNumber] = event
	}
}

func (registry *SnapshotRegistry) listen() {
	for {
		select {
		case <-registry.exitC:
			return
		case n := <-registry.newBlockC:
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
			registry.logger.Noticef("snapshot at (block #%d) for filter (%s) failed, reason: %s", number, ev.FilterId, err.Error())
			registry.feedback(FILTER_SNAPSHOT_RESULT, false, ev.FilterId, MakeSnapshotFailedErr)
		} else {
			registry.logger.Noticef("snapshot at (block #%d) for filter (%s) success", number, ev.FilterId)
			registry.feedback(FILTER_SNAPSHOT_RESULT, true, ev.FilterId, EmptyMessage)
		}
		delete(registry.rq, number)
	}
}

// makeSnapshot do snapshot job here.
// Several things to do:
// 1. Make a world state copy with the help of leveldb snapshot mechanism.
// 2. Push the pivot block to snapshot data.
// 3. Write meta to snapshot manifest file.
func (registry *SnapshotRegistry) makeSnapshot(filterId string, number uint64) error {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	if err := registry.duplicate(filterId); err != nil {
		return err
	}
	if err := registry.pushBlock(filterId, number); err != nil {
		return err
	}
	if err := registry.writeMeta(filterId, number); err != nil {
		return err
	}
	return nil
}

func (registry *SnapshotRegistry) duplicate(filterId string) error {
	sPath := path.Join("snapshots", registry.snapshotId(filterId))
	if err := registry.executor.db.MakeSnapshot(sPath, cm.RetrieveSnapshotFileds()); err != nil {
		return err
	}
	return nil
}

func (registry *SnapshotRegistry) pushBlock(filterId string, number uint64) error {
	blk, err := edb.GetBlockByNumber(registry.namespace, number)
	if err != nil {
		return err
	}
	p := path.Join("snapshots", registry.snapshotId(filterId))
	sdb, err := hyperdb.NewDatabase(registry.executor.conf, p, hyperdb.GetDatabaseType(registry.executor.conf), registry.namespace)
	if err != nil {
		return err
	}
	defer sdb.Close()
	wb := sdb.NewBatch()
	err, _ = edb.PersistBlock(wb, blk, true, true)
	if err != nil {
		return err
	}
	return nil
}

func (registry *SnapshotRegistry) writeMeta(filterId string, number uint64) error {
	blk, err := edb.GetBlockByNumber(registry.namespace, number)
	if err != nil {
		return err
	}
	hash, err := registry.CalculateStateHash(filterId, common.BytesToHash(blk.MerkleRoot), number)
	if err != nil || hash != common.BytesToHash(blk.MerkleRoot) {
		return SnapshotContentInvalidErr
	}
	d := time.Unix(time.Now().Unix(), 0).Format("2006-01-02-15:04:05")
	manifest := common.Manifest{
		Height:     number,
		BlockHash:  common.Bytes2Hex(blk.BlockHash),
		FilterId:   filterId,
		MerkleRoot: hash.Hex(),
		Date:       d,
		Namespace:  registry.namespace,
	}
	if err := registry.rwc.Write(manifest); err != nil {
		return err
	}

	p := path.Join("snapshots", registry.snapshotId(filterId))
	sdb, err := hyperdb.NewDatabase(registry.executor.conf, p, hyperdb.GetDatabaseType(registry.executor.conf), registry.namespace)
	if err != nil {
		return err
	}
	defer sdb.Close()

	if err := edb.PersistSnapshotMeta(sdb.NewBatch(), &manifest, true, true); err != nil {
		return err
	}

	return nil
}

// CalculateStateHash use bucket tree to force recompute the whole state hash fingerprint.
func (registry *SnapshotRegistry) CalculateStateHash(filterId string, compareTag common.Hash, height uint64) (common.Hash, error) {
	localDb, err := hyperdb.NewDatabase(registry.executor.conf, path.Join("snapshots", registry.snapshotId(filterId)), hyperdb.GetDatabaseType(registry.executor.conf), registry.namespace)
	defer localDb.Close()
	if err != nil {
		return common.Hash{}, err
	}
	localState, err := hyperstate.New(compareTag, localDb, nil, registry.executor.conf, height, registry.executor.namespace)
	if err != nil {
		return common.Hash{}, err
	}
	curhash, err := localState.RecomputeCryptoHash()
	registry.logger.Noticef("full quantity calculation, snapshot world state hash (%s)", curhash.Hex())
	if err != nil {
		return common.Hash{}, err
	} else {
		return curhash, nil
	}
}

// deleteSnapshot remove specific snapshot.
func (registry *SnapshotRegistry) deleteSnapshot(filterId string) error {
	conf := registry.executor.conf
	fId := registry.snapshotId(filterId)
	sPath := registry.snapshotPath(hyperdb.GetDatabaseHome(conf), fId)
	localCmd := cmd.Command("rm", "-rf", sPath)
	if err := localCmd.Run(); err != nil {
		return err
	}
	return nil
}

// compress compress snapshot data for network transport.
func (registry *SnapshotRegistry) compress(filterId string) (error, int64) {
	if !registry.rwc.Contain(filterId) {
		return SnapshotDoesntExistErr, 0
	}
	spath := path.Join(hyperdb.GetDatabaseHome(registry.executor.conf), "snapshots", registry.snapshotId(filterId))
	localCmd := cmd.Command("tar", "-C", filepath.Dir(spath), "-czvf", spath+".tar.gz", registry.snapshotId(filterId))
	if err := localCmd.Run(); err != nil {
		return err, 0
	}
	fd, err := os.OpenFile(spath+".tar.gz", os.O_RDONLY, 0644)
	if err != nil {
		return err, 0
	}
	stat, err := fd.Stat()
	if err != nil {
		return err, 0
	}
	return nil, stat.Size()
}

// checkSnapshotRequest if the pivot point of request is less than current chain height,
// this type of request is recognized as invalid one.
func (registry *SnapshotRegistry) checkSnapshotRequest(event event.SnapshotEvent) bool {
	if event.BlockNumber == LatestBlockNumber {
		return true
	}
	chainHeight := edb.GetHeightOfChain(registry.namespace)
	if event.BlockNumber < chainHeight {
		return false
	}
	return true
}

// checkDeletionRequest check whether the deletion request is satisfied.
// If the snapshot required to delete is the current genesis, the delete operation is invalid.
func (registry *SnapshotRegistry) checkDeletionRequest(event event.DeleteSnapshotEvent) bool {
	err, manifest := registry.rwc.Read(event.FilterId)
	if err != nil {
		return false
	}
	err, curGenesis := edb.GetGenesisTag(registry.namespace)
	if err != nil {
		return false
	}
	if curGenesis == manifest.Height {
		return false
	}
	return true
}

// isExecuteImmediate check is this snapshot request is instant trigger type.
func (registry *SnapshotRegistry) isExecuteImmediate(event event.SnapshotEvent) bool {
	return event.BlockNumber == LatestBlockNumber
}

// notifyNewBlock condition trigger source to wake up pending snapshot request.
func (registry *SnapshotRegistry) notifyNewBlock(number uint64) {
	registry.newBlockC <- number
}

func (registry *SnapshotRegistry) feedback(t int, isSuccess bool, filterId string, message string) {
	registry.executor.sendFilterEvent(t, isSuccess, filterId, message)
}

func (registry *SnapshotRegistry) exit() {
	registry.exitC <- struct{}{}
}

func (registry *SnapshotRegistry) snapshotId(filterId string) string {
	return "SNAPSHOT_" + filterId
}

func (registry *SnapshotRegistry) snapshotPath(base string, filterID string) string {
	return path.Join(base, "snapshots", filterID)
}
