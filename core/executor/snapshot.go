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

// Snapshot manger manages all state snapshots of the local node.
// 1. What is snapshot?
// Snapshot refers to a file backup of the local world state(All account status collection)
// 2. Why we need snapshot?
// In hyperchain, data archives are based on state snapshots. So before the blockchain archive,
// a valid state snapshot is required.
package executor

import (
	"os"
	cmd "os/exec"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/hyperchain/hyperchain/common"
	cm "github.com/hyperchain/hyperchain/core/common"
	edb "github.com/hyperchain/hyperchain/core/ledger/chain"
	"github.com/hyperchain/hyperchain/core/ledger/state"
	"github.com/hyperchain/hyperchain/hyperdb"
	"github.com/hyperchain/hyperchain/manager/event"

	"github.com/op/go-logging"
)

// Snapshot creates a state snapshot in the specified block height.
// Note, the specified block height must not less than current chain height.
func (executor *Executor) Snapshot(ev event.SnapshotEvent) {
	executor.snapshotReg.Snapshot(ev)
}

// DeleteSnapshot deletes a snapshot with specified snapshot name.
// Note, if the specified snapshot is current genesis, the deletion will been regarded
// as an invalid operation.
func (executor *Executor) DeleteSnapshot(ev event.DeleteSnapshotEvent) {
	executor.snapshotReg.DeleteSnapshot(ev)
}

// InsertSnapshot inserts a snapshot to local storage. The snapshot may received from network.
func (executor *Executor) InsertSnapshot(meta cm.Manifest) {
	executor.snapshotReg.Insert(meta)
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
	rw        cm.ManifestRW
}

func NewSnapshotRegistry(namespace string, logger *logging.Logger, executor *Executor) *SnapshotRegistry {
	return &SnapshotRegistry{
		namespace: namespace,
		rq:        make(map[uint64]event.SnapshotEvent),
		logger:    logger,
		newBlockC: make(chan uint64),
		exitC:     make(chan struct{}),
		executor:  executor,
		rw:        cm.NewManifestHandler(executor.conf.GetManifestPath()),
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
		registry.feedback(FILTER_SNAPSHOT_RESULT, false, event.FilterId, InvalidSnapshotReqMsg)
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
		registry.feedback(FILTER_DELETE_SNAPSHOT, false, event.FilterId, InvalidDeletionReqMsg)
		event.Cont <- InvalidSnapshotDeletionErr
		return
	} else {
		if !registry.rw.Contain(event.FilterId) {
			registry.feedback(FILTER_DELETE_SNAPSHOT, false, event.FilterId, SnapshotNotExistMsg)
			event.Cont <- SnapshotDoesntExistErr
			return
		} else {
			if err := registry.deleteSnapshot(event.FilterId); err != nil {
				registry.logger.Warningf("delete snapshot %s failed. reason %s", event.FilterId, err.Error())
				registry.feedback(FILTER_DELETE_SNAPSHOT, false, event.FilterId, DeleteSnapshotMsg)
				event.Cont <- DeleteSnapshotFailedErr
				return
			}
			if err := registry.rw.Delete(event.FilterId); err != nil {
				registry.feedback(FILTER_DELETE_SNAPSHOT, false, event.FilterId, DeleteSnapshotMsg)
				event.Cont <- DeleteSnapshotFailedErr
				return
			}
			close(event.Cont)
			return
		}
	}
}

func (registry *SnapshotRegistry) CompressSnapshot(filterId string) (error, int64) {
	return registry.compress(filterId)
}

func (registry *SnapshotRegistry) CompressedSnapshotPath(filterId string) string {
	return path.Join(cm.GetDatabaseHome(registry.namespace, registry.executor.conf.conf), "snapshots", registry.snapshotId(filterId)+".tar.gz")
}

func (registry *SnapshotRegistry) Insert(meta cm.Manifest) {
	sdir := path.Join(cm.GetDatabaseHome(registry.namespace, registry.executor.conf.conf), "snapshots")
	if _, err := os.Stat(sdir); os.IsNotExist(err) {
		c := cmd.Command("mkdir", "-p", sdir)
		if e := c.Run(); e != nil {
			return
		}
	}
	spath := path.Join(cm.GetDatabaseHome(registry.namespace, registry.executor.conf.conf), "ws", "ws_"+meta.FilterId, "SNAPSHOT_"+meta.FilterId)
	dump := cmd.Command("mv", "-f", spath, sdir)
	if e := dump.Run(); e != nil {
		return
	}

	clear := cmd.Command("rm", "-rf", path.Join(cm.GetDatabaseHome(registry.namespace, registry.executor.conf.conf), "ws"))
	if e := clear.Run(); e != nil {
		return
	}

	if err := registry.rw.Write(meta); err != nil {
		registry.logger.Warningf("insert snapshost %s failed. reason %s", meta.FilterId, err.Error())
	}
	registry.logger.Noticef("insert snapshost %s success", meta.FilterId)
}

/*
	Internal functions
*/
func (registry *SnapshotRegistry) executeImmediately(ev event.SnapshotEvent) {
	height := edb.GetHeightOfChain(registry.namespace)
	if err := registry.makeSnapshot(ev.FilterId, height); err != nil {
		registry.logger.Noticef("snapshot at (block #%d) for filter (%s) failed, reason %s", height, ev.FilterId, err.Error())
		registry.feedback(FILTER_SNAPSHOT_RESULT, false, ev.FilterId, MakeSnapshotFailedMsg)
	} else {
		registry.logger.Noticef("snapshot at (block #%d) for filter (%s) success", height, ev.FilterId)
		registry.feedback(FILTER_SNAPSHOT_RESULT, true, ev.FilterId, EmptyMessage)
	}
}

// addRequest saves snapshot request into pending queue, wait for condition trigger.
// TODO write journal first to avoid request missing if process crash down.
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
			registry.feedback(FILTER_SNAPSHOT_RESULT, false, ev.FilterId, MakeSnapshotFailedMsg)
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
func (registry *SnapshotRegistry) makeSnapshot(filterId string, number uint64) (err error) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	// Remove snapshot file if error occur
	var begin time.Time = time.Now()

	defer func() {
		if err == nil {
			registry.logger.Noticef("make snapshot for %s elasped %v", filterId, time.Since(begin))
		}
	}()

	defer func() {
		if err != nil {
			registry.deleteSnapshot(filterId)
		}
	}()

	if err = registry.backupStatus(filterId); err != nil {
		return err
	}
	if err = registry.pushBlock(filterId, number); err != nil {
		return err
	}
	if err = registry.writeMeta(filterId, number); err != nil {
		return err
	}
	return nil
}

// backupStatus creates a world state backup file.
func (registry *SnapshotRegistry) backupStatus(filterId string) error {
	sPath := path.Join("snapshots", registry.snapshotId(filterId))
	if err := registry.executor.db.MakeSnapshot(sPath, cm.RetrieveSnapshotFileds()); err != nil {
		return err
	}
	return nil
}

// pushBlock writes the block to snapshot file.
func (registry *SnapshotRegistry) pushBlock(filterId string, number uint64) error {
	blk, err := edb.GetBlockByNumber(registry.namespace, number)
	if err != nil {
		return err
	}
	p := path.Join("snapshots", registry.snapshotId(filterId))
	sdb, err := hyperdb.NewDatabase(registry.executor.conf.conf, p, hyperdb.GetDatabaseType(registry.executor.conf.conf), registry.namespace)
	if err != nil {
		return err
	}
	defer sdb.Close()
	wb := sdb.NewBatch()
	_, err = edb.PersistBlock(wb, blk, true, true)
	if err != nil {
		return err
	}
	return nil
}

// writeMeta writes snapshot meta info.
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
	genesis, _ := edb.GetGenesisTag(registry.namespace)
	manifest := cm.Manifest{
		Height:     number,
		Genesis:    genesis,
		BlockHash:  common.Bytes2Hex(blk.BlockHash),
		FilterId:   filterId,
		MerkleRoot: hash.Hex(),
		Date:       d,
		Namespace:  registry.namespace,
	}
	if err := registry.rw.Write(manifest); err != nil {
		return err
	}

	p := path.Join("snapshots", registry.snapshotId(filterId))
	sdb, err := hyperdb.NewDatabase(registry.executor.conf.conf, p, hyperdb.GetDatabaseType(registry.executor.conf.conf), registry.namespace)
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
	localDb, err := hyperdb.NewDatabase(registry.executor.conf.conf, path.Join("snapshots", registry.snapshotId(filterId)), hyperdb.GetDatabaseType(registry.executor.conf.conf), registry.namespace)
	defer localDb.Close()
	if err != nil {
		return common.Hash{}, err
	}
	localState, err := state.New(compareTag, localDb, nil, registry.executor.conf.conf, height)
	if err != nil {
		return common.Hash{}, err
	}
	curhash, err := localState.RecomputeCryptoHash()
	registry.logger.Noticef("full quantity calculation, snapshot world state hash (%s), expect (%s)", curhash.Hex(), compareTag.Hex())
	if err != nil {
		return common.Hash{}, err
	} else {
		return curhash, nil
	}
}

// deleteSnapshot remove specific snapshot.
func (registry *SnapshotRegistry) deleteSnapshot(filterId string) error {
	sPath := registry.snapshotPath(cm.GetDatabaseHome(registry.namespace, registry.executor.conf.conf), registry.snapshotId(filterId))
	localCmd := cmd.Command("rm", "-rf", sPath)
	if err := localCmd.Run(); err != nil {
		return err
	}
	return nil
}

// compress compress snapshot data for network transport.
func (registry *SnapshotRegistry) compress(filterId string) (error, int64) {
	if !registry.rw.Contain(filterId) {
		return SnapshotDoesntExistErr, 0
	}
	spath := path.Join(cm.GetDatabaseHome(registry.namespace, registry.executor.conf.conf), "snapshots", registry.snapshotId(filterId))
	localCmd := cmd.Command("tar", "-C", filepath.Dir(spath), "-czvf", spath+".tar.gz", registry.snapshotId(filterId))
	if err := localCmd.Run(); err != nil {
		return err, 0
	}
	fd, err := os.OpenFile(spath+".tar.gz", os.O_RDONLY, 0644)
	if err != nil {
		return err, 0
	}
	defer fd.Close()
	stat, err := fd.Stat()
	if err != nil {
		return err, 0
	}
	return nil, stat.Size()
}

// checkSnapshotRequest if the pivot point of request is less than current chain height,
// this type of request is recognized as invalid one.
func (registry *SnapshotRegistry) checkSnapshotRequest(event event.SnapshotEvent) bool {
	chainHeight := edb.GetHeightOfChain(registry.namespace)
	if event.BlockNumber < chainHeight && event.BlockNumber != LatestBlockNumber {
		return false
	}
	if _, meta := registry.rw.Search(chainHeight); (meta != cm.Manifest{}) && event.BlockNumber == LatestBlockNumber {
		return false
	}
	if _, meta := registry.rw.Search(event.BlockNumber); (meta != cm.Manifest{}) && event.BlockNumber != LatestBlockNumber {
		return false
	}
	return true
}

// checkDeletionRequest check whether the deletion request is satisfied.
// If the snapshot required to delete is the current genesis, the delete operation is invalid.
func (registry *SnapshotRegistry) checkDeletionRequest(event event.DeleteSnapshotEvent) bool {
	err, manifest := registry.rw.Read(event.FilterId)
	if err != nil {
		return false
	}
	curGenesis, err := edb.GetGenesisTag(registry.namespace)
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
