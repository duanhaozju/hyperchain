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
	"time"

	com "hyperchain/common"
	edb "hyperchain/core/ledger/chain"
	"hyperchain/core/types"
	"hyperchain/manager/event"

	"github.com/op/go-logging"
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

type ArchiveManager struct {
	executor  *Executor
	registry  *SnapshotRegistry
	namespace string
	logger    *logging.Logger
	rw        com.ArchiveMetaRW
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
		if err := mgr.migrate(manifest); err != nil {
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

/*
	Internal Functions
*/
func (mgr *ArchiveManager) migrate(manifest com.Manifest) error {
	curGenesis, err := edb.GetGenesisTag(mgr.namespace)
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

	if !mgr.checkRequest(manifest, meta) {
		mgr.logger.Warningf("archive request not satisfied with requirement")
		return ArchiveRequestNotSatisfiedErr
	}

	if meta.Height >= curGenesis && meta.Height != 0 {
		mgr.logger.Warningf("archive database, chain height (#%d) larger than current genesis (#%d)", meta.Height, curGenesis)
	} else if meta.Height != curGenesis-1 {
		mgr.logger.Warningf("archive database, chain height (#%d) not continuous with current genesis (#%d)", meta.Height, curGenesis)
	}

	var txc, rectc, ic uint64
	var tb, te int64
	if err, tb, te = mgr.getTimestampRange(curGenesis, manifest.Height); err != nil {
		mgr.logger.Errorf("[Namespace = %s] get timestamp range failed, error msg %s", mgr.namespace, err.Error())
		return err
	}
	for i := curGenesis; i < manifest.Height; i += 1 {
		block, err := edb.GetBlockByNumber(mgr.namespace, i)
		if err != nil {
			mgr.logger.Errorf("miss block %d ,error msg %s", i, err.Error())
			return err
		}
		for idx, tx := range block.Transactions {
			if edb.DeleteTransactionMeta(olBatch, tx.GetHash().Bytes(), false, false); err != nil {
				mgr.logger.Errorf("[Namespace = %s] archive useless transaction meta in block %d failed, error msg %s", mgr.namespace, i, err.Error())
				return err
			}
			receipt := edb.GetRawReceipt(mgr.namespace, tx.GetHash())
			if err := edb.DeleteReceipt(olBatch, tx.GetHash().Bytes(), false, false); err != nil {
				mgr.logger.Errorf("[Namespace = %s] archive useless receipt in block %d failed, error msg %s", mgr.namespace, i, err.Error())
				return err
			}
			// persist transaction meta data
			meta := &types.TransactionMeta{
				BlockIndex: i,
				Index:      int64(idx),
			}
			if err := edb.PersistTransactionMeta(avBatch, meta, tx.GetHash(), false, false); err != nil {
				mgr.logger.Errorf("[Namespace = %s] archive txmeta in block %d to historic database failed, error msg %s", mgr.namespace, i, err.Error())
				return err
			} else {
				txc += 1
			}
			// If the transaction is invalid, there will no related receipt been saved.
			if receipt == nil {
				continue
			}
			if _, err := edb.PersistReceipt(avBatch, receipt, false, false); err != nil {
				mgr.logger.Errorf("[Namespace = %s] archive receipt in block %d to historic database failed, error msg %s", mgr.namespace, i, err.Error())
				return err
			} else {
				rectc += 1
			}
		}
		// delete block
		if err := edb.DeleteBlockByNum(mgr.namespace, olBatch, i, false, false); err != nil {
			mgr.logger.Errorf("[Namespace = %s] archive useless block %d failed, error msg %s", mgr.namespace, i, err.Error())
			return err
		}

		if _, err := edb.PersistBlock(avBatch, block, false, false); err != nil {
			mgr.logger.Errorf("[Namespace = %s] archive block %d to historic database failed. error msg %s", mgr.namespace, i, err.Error())
			return err
		}
	}
	if err := edb.DeleteJournalInRange(olBatch, curGenesis, manifest.Height, false, false); err != nil {
		mgr.logger.Errorf("[Namespace = %s] archive useless journals failed, error msg %s", mgr.namespace, err.Error())
		return err
	}
	// delete invalid records
	if ic, err = edb.DumpDiscardTransactionInRange(mgr.executor.db, olBatch, avBatch, tb, te, false, false); err != nil {
		mgr.logger.Errorf("[Namespace = %s] archive useless invalid records failed, error msg %s", mgr.namespace, err.Error())
		return err
	}

	// update chain
	if err := edb.UpdateGenesisTag(mgr.namespace, manifest.Height, olBatch, false, false); err != nil {
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

	if err := mgr.rw.Write(com.ArchiveMeta{
		Height:       manifest.Height - 1,
		TransactionN: meta.TransactionN + txc,
		ReceiptN:     meta.ReceiptN + rectc,
		InvalidTxN:   meta.InvalidTxN + ic,
		LatestUpdate: time.Unix(time.Now().Unix(), 0).Format("2006-01-02-15:04:05"),
	}); err != nil {
		mgr.logger.Errorf("[Namespace = %s] update meta failed. error msg %s", mgr.namespace, err.Error())
		return err
	}
	return nil
}

func (mgr *ArchiveManager) feedback(isSuccess bool, filterId string, message string) {
	mgr.executor.sendFilterEvent(FILTER_ARCHIVE, isSuccess, filterId, message)
}

func (mgr *ArchiveManager) getTimestampRange(begin, end uint64) (error, int64, int64) {
	bk, err := edb.GetBlockByNumber(mgr.namespace, begin)
	if err != nil {
		return err, 0, 0
	}
	ek, err := edb.GetBlockByNumber(mgr.namespace, end)
	if err != nil {
		return err, 0, 0
	}
	return nil, bk.CommitTime, ek.CommitTime
}

// check archive request is valid
// 1. specified snapshot is safe enough to do archive operation (snapshot.Height + threshold < height)
// 2. archive db is continuous with current blockchain (optional)
func (mgr *ArchiveManager) checkRequest(manifest com.Manifest, meta com.ArchiveMeta) bool {
	curHeigit := edb.GetHeightOfChain(mgr.namespace)
	genesis, err := edb.GetGenesisTag(mgr.namespace)
	if err != nil {
		return false
	}
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
}
