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
	com "github.com/hyperchain/hyperchain/core/common"
	"github.com/hyperchain/hyperchain/core/ledger/chain"
	"github.com/hyperchain/hyperchain/core/types"
	"github.com/hyperchain/hyperchain/hyperdb/db"
	"github.com/hyperchain/hyperchain/manager/event"
	"github.com/op/go-logging"
	"sync"
	"time"
)

// Data archives are based on state snapshots, if you want do a archive operation,
// a valid snapshot is the precondition.
//
// e.g.
// Current chain height is 100, and current genesis block is 0.
// A snapshot s is created when chain height is 30(s is the world state backup file when chain height is 30)
// An archive operation based on s will dump blocks [0-29] and related transactions, receipts to historical database.
// Genesis block number trans to 30, and s becomes the new genesis world state.

// Archive accepts an archive request.
func (executor *Executor) Archive(event event.ArchiveEvent) {
	executor.archiveMgr.Archive(event)
}

func (executor *Executor) ArchiveRestore(event event.ArchiveRestoreEvent) {
	executor.archiveMgr.Restore(event)
}

type ArchiveManager struct {
	executor  *Executor
	registry  *SnapshotRegistry
	namespace string
	logger    *logging.Logger
	rw        com.ArchiveMetaRW
	lock      sync.Mutex
}

func NewArchiveManager(namespace string, executor *Executor, registry *SnapshotRegistry, logger *logging.Logger) *ArchiveManager {
	return &ArchiveManager{
		namespace: namespace,
		executor:  executor,
		registry:  registry,
		logger:    logger,
		rw:        com.NewArchiveMetaHandler(executor.conf.GetArchiveMetaPath()),
	}
}

/*
	External Functions
*/
func (mgr *ArchiveManager) Archive(event event.ArchiveEvent) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()
	var manifest com.Manifest
	if !event.Sync {
		close(event.Cont)
	}
	if !mgr.registry.rw.Contain(event.FilterId) {
		mgr.feedback(false, event.FilterId, SnapshotNotExistMsg)
		if event.Sync {
			event.Cont <- SnapshotDoesntExistErr
		}
	} else {
		_, manifest = mgr.registry.rw.Read(event.FilterId)
		if err := mgr.migrate(manifest, false, true); err != nil {
			mgr.logger.Noticef("archive for (filter %s) failed, detail %s", event.FilterId, err.Error())
			mgr.feedback(false, event.FilterId, ArchiveFailedMsg)
			if event.Sync {
				event.Cont <- err
			}
		} else {
			mgr.logger.Noticef("archive for (filter %s) success, genesis block changes to %d and the relative genesis world state %s",
				event.FilterId, manifest.Height, manifest.FilterId)
			mgr.feedback(true, event.FilterId, EmptyMessage)
			if event.Sync {
				close(event.Cont)
			}
		}
	}
}

func (mgr *ArchiveManager) Restore(event event.ArchiveRestoreEvent) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()
	if !event.Sync {
		close(event.Ack)
	}

	if !event.All && !mgr.executor.snapshotReg.rw.Contain(event.FilterId) {
		mgr.feedback(false, event.FilterId, SnapshotNotExistMsg)
		if event.Sync {
			event.Ack <- SnapshotDoesntExistErr
		}
	} else {
		_, manifest := mgr.registry.rw.Read(event.FilterId)
		if err := mgr.migrate(manifest, event.All, false); err != nil {
			mgr.logger.Noticef("archive for (filter %s) failed, detail %s", event.FilterId, err.Error())
			mgr.feedback(false, event.FilterId, ArchiveStoreFailedMsg)
			if event.Sync {
				event.Ack <- err
			}
		} else {
			mgr.feedback(true, event.FilterId, EmptyMessage)
			if event.Sync {
				close(event.Ack)
			}
		}
	}
}

/*
	Internal Functions
*/
// migrate moves specified blockchain data from online database to archive database if forward is true,
// otherwise moves back if forward is false
func (mgr *ArchiveManager) migrate(manifest com.Manifest, all, forward bool) error {
	curGenesis, err := chain.GetGenesisTag(mgr.namespace)
	if err != nil {
		return err
	}
	olBatch := mgr.executor.db.NewBatch()
	avBatch := mgr.executor.archiveDb.NewBatch()

	var meta com.ArchiveMeta
	if mgr.rw.Exist() {
		if err, lmeta := mgr.rw.Read(); err != nil {
			return err
		} else {
			meta = lmeta
		}
	}

	if !mgr.checkRequest(manifest, meta, all, forward) {
		mgr.logger.Warningf("archive request not satisfied with requirement")
		return ArchiveRequestNotSatisfiedErr
	}

	if meta.Height >= curGenesis && meta.Height != 0 {
		mgr.logger.Warningf("archive database, chain height (#%d) larger than current genesis (#%d)", meta.Height, curGenesis)
	} else if meta.Height != curGenesis-1 {
		mgr.logger.Warningf("archive database, chain height (#%d) not continuous with current genesis (#%d)", meta.Height, curGenesis)
	}

	var (
		start, limit   uint64
		txc, rectc, ic uint64
		tb, te         int64
	)
	if forward {
		start, limit = curGenesis, manifest.Height
	} else {
		start, limit = manifest.Height, curGenesis
		if all {
			start = 0
		}
	}

	mgr.logger.Noticef("migrate blockchain data. range [%d - %d)", start, limit)

	var (
		deleteBatch, writeBatch db.Batch
		deleteDb                db.Database
	)

	if forward {
		deleteDb = mgr.executor.db
		deleteBatch, writeBatch = olBatch, avBatch
	} else {
		deleteDb = mgr.executor.archiveDb
		deleteBatch, writeBatch = avBatch, olBatch
	}

	for i := start; i < limit; i += 1 {
		block, err := chain.GetBlockByNumberFunc(deleteDb, i)
		if err != nil {
			mgr.logger.Errorf("miss block %d, error msg %s", i, err.Error())
			return err
		}

		for idx, tx := range block.Transactions {
			if chain.DeleteTransactionMeta(deleteBatch, tx.GetHash().Bytes(), false, false); err != nil {
				mgr.logger.Errorf("[Namespace = %s] archive/restore useless transaction meta in block %d failed, error msg %s", mgr.namespace, i, err.Error())
				return err
			}
			// persist transaction meta data
			meta := &types.TransactionMeta{
				BlockIndex: i,
				Index:      int64(idx),
			}
			if err := chain.PersistTransactionMeta(writeBatch, meta, tx.GetHash(), false, false); err != nil {
				mgr.logger.Errorf("[Namespace = %s] archive/restore txmeta in block %d to historic/online database failed, error msg %s", mgr.namespace, i, err.Error())
				return err
			} else {
				txc += 1
			}

			receipt := chain.GetRawReceiptFunc(deleteDb, tx.GetHash())
			// If the transaction is invalid, there will no related receipt been saved.
			if receipt == nil {
				continue
			}

			if err := chain.DeleteReceipt(deleteBatch, tx.GetHash().Bytes(), false, false); err != nil {
				mgr.logger.Errorf("[Namespace = %s] archive/restore useless receipt in block %d failed, error msg %s", mgr.namespace, i, err.Error())
				return err
			}
			if _, err := chain.PersistReceipt(writeBatch, receipt, false, false); err != nil {
				mgr.logger.Errorf("[Namespace = %s] archive/restore receipt in block %d to historic/online database failed, error msg %s", mgr.namespace, i, err.Error())
				return err
			} else {
				rectc += 1
			}
		}
		// delete block
		if err := chain.DeleteBlockByNumberFunc(deleteDb, deleteBatch, i, false, false); err != nil {
			mgr.logger.Errorf("[Namespace = %s] archive/restore useless block %d failed, error msg %s", mgr.namespace, i, err.Error())
			return err
		}

		if _, err := chain.PersistBlock(writeBatch, block, false, false); err != nil {
			mgr.logger.Errorf("[Namespace = %s] archive/restore block %d to historic database failed. error msg %s", mgr.namespace, i, err.Error())
			return err
		}
	}
	if forward {
		// forward only
		if err := chain.DeleteJournalInRange(deleteBatch, start, limit, false, false); err != nil {
			mgr.logger.Errorf("[Namespace = %s] archive useless journals failed, error msg %s", mgr.namespace, err.Error())
			return err
		}
	}

	if err, tb, te = mgr.getTimestampRange(deleteDb, start, limit-1); err != nil {
		mgr.logger.Errorf("[Namespace = %s] get timestamp range failed, error msg %s", mgr.namespace, err.Error())
		return err
	}

	// delete invalid records
	if ic, err = chain.DumpDiscardTransactionInRange(deleteDb, deleteBatch, writeBatch, tb, te, false, false); err != nil {
		mgr.logger.Errorf("[Namespace = %s] archive/restore useless invalid records failed, error msg %s", mgr.namespace, err.Error())
		return err
	}

	var ng uint64
	if forward {
		ng = manifest.Height
	} else {
		ng = start
	}

	// update chain
	if err := chain.UpdateGenesisTag(mgr.namespace, ng, olBatch, false, false); err != nil {
		mgr.logger.Errorf("[Namespace = %s] update chain genesis field failed, error msg %s", mgr.namespace, err.Error())
		return err
	}

	if err := avBatch.Write(); err != nil {
		mgr.logger.Errorf("[Namespace = %s] flush to historic database failed, error msg %s", mgr.namespace, err.Error())
		return err
	}
	if err := olBatch.Write(); err != nil {
		mgr.logger.Errorf("[Namespace = %s] flush to online database failed, error msg %s", mgr.namespace, err.Error())
		return err
	}
	var ttotal, rtotal, itotal uint64
	if forward {
		ttotal, rtotal, itotal = meta.TransactionN+txc, meta.ReceiptN+rectc, meta.InvalidTxN+ic
	} else {
		ttotal, rtotal, itotal = meta.TransactionN-txc, meta.ReceiptN-rectc, meta.InvalidTxN-ic
	}

	var cheight uint64
	if forward || start != 0 {
		cheight = manifest.Height - 1
	} else {
		cheight = 0
	}

	if err := mgr.rw.Write(com.ArchiveMeta{
		Height:       cheight,
		TransactionN: ttotal,
		ReceiptN:     rtotal,
		InvalidTxN:   itotal,
		LatestUpdate: time.Unix(time.Now().Unix(), 0).Format("2006-01-02-15:04:05"),
	}); err != nil {
		mgr.logger.Errorf("[Namespace = %s] update meta failed. error msg %s", mgr.namespace, err.Error())
		return err
	}
	if all {
		mgr.logger.Noticef("restore all success, genesis block changes to 0")
	} else {
		mgr.logger.Noticef("restore for (filter %s) success, genesis block changes to %d",
			manifest.FilterId, start)
	}
	return nil
}

func (mgr *ArchiveManager) feedback(isSuccess bool, filterId string, message string) {
	mgr.executor.sendFilterEvent(FILTER_ARCHIVE, isSuccess, filterId, message)
}

func (mgr *ArchiveManager) getTimestampRange(db db.Database, begin, end uint64) (error, int64, int64) {
	bk, err := chain.GetBlockByNumberFunc(db, begin)
	if err != nil {
		return err, 0, 0
	}
	ek, err := chain.GetBlockByNumberFunc(db, end)
	if err != nil {
		return err, 0, 0
	}
	return nil, bk.CommitTime, ek.CommitTime
}

// check archive request is valid
// 1. specified snapshot is safe enough to do archive operation (snapshot.Height + threshold < height)
// 2. archive db is continuous with current blockchain (optional)
func (mgr *ArchiveManager) checkRequest(manifest com.Manifest, meta com.ArchiveMeta, all, forward bool) bool {
	curHeigit := chain.GetHeightOfChain(mgr.namespace)
	genesis, err := chain.GetGenesisTag(mgr.namespace)
	if err != nil {
		return false
	}
	if forward {
		if curHeigit < uint64(mgr.executor.conf.GetArchiveThreshold())+manifest.Height {
			return false
		}
		// Optional. If user set the `force consistency` config item as true,
		// which means if archived chain is not continuous with the online chain,
		// the request can be regarded as a invalid one.
		if mgr.executor.conf.IsArchiveForceConsistency() {
			if genesis != meta.Height+1 {
				return false
			}
		}
		// if genesis block is larger or equal with the snapshot height,
		// the request is regarded as a invalid one.
		if genesis >= manifest.Height {
			return false
		}
		return true
	} else {
		// historical database must has all required block,
		// check height is just the first step :)
		if meta.Height+1 != genesis {
			return false
		}
		if all {
			return true
		}
		if manifest.Height >= genesis {
			return false
		}
		return true
	}
}
