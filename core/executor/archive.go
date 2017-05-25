package executor

import (
	"hyperchain/common"
	edb "hyperchain/core/db_utils"
	"github.com/op/go-logging"
	"hyperchain/core/types"
	"hyperchain/manager/event"
)

func (executor *Executor) Archive(event event.ArchiveEvent) {
	executor.archiveMgr.Archive(event.FilterId)
}


type ArchiveManager struct {
	executor   *Executor
	registry   *SnapshotRegistry
	namespace  string
	logger     *logging.Logger
}

func NewArchiveManager(namespace string, executor *Executor, registry *SnapshotRegistry, logger *logging.Logger) *ArchiveManager {
	return &ArchiveManager{
		namespace: namespace,
		executor:  executor,
		registry:  registry,
		logger:    logger,
	}
}

/*
	External Functions
 */
func (mgr *ArchiveManager) Archive(filterId string) {
	var manifest common.Manifest
	if !mgr.registry.rwc.Contain(filterId) {
		mgr.feedback(false, filterId, SnapshotNotExistErr)
	} else {
		_, manifest = mgr.registry.rwc.Read(filterId)
		if err := mgr.migrate(manifest); err != nil {
			mgr.feedback(false, filterId, ArchiveFailedErr)
		} else {
			mgr.feedback(true, filterId, EmptyMessage)
		}
	}
}

/*
	Internal Functions
 */
func (mgr *ArchiveManager) migrate(manifest common.Manifest) error {
	err, curGenesis := edb.GetGenesisTag(mgr.namespace)
	if err != nil {
		return err
	}
	olBatch := mgr.executor.db.NewBatch()
	avBatch := mgr.executor.archiveDb.NewBatch()

	for i := curGenesis; i < manifest.Height; i += 1 {
		block, err := edb.GetBlockByNumber(mgr.namespace, i)
		if err != nil {
			mgr.logger.Errorf("miss block %d ,error msg %s", i, err.Error())
			return err
		}
		for idx, tx := range block.Transactions {
			receipt := edb.GetRawReceipt(mgr.namespace, tx.GetHash())
			if err := edb.DeleteTransaction(olBatch, tx.GetHash().Bytes(), false, false); err != nil {
				mgr.logger.Errorf("[Namespace = %s] archive useless tx in block %d failed, error msg %s", mgr.namespace, i, err.Error())
				return err
			}
			if err := edb.DeleteReceipt(olBatch, tx.GetHash().Bytes(), false, false); err != nil {
				mgr.logger.Errorf("[Namespace = %s] archive useless receipt in block %d failed, error msg %s", mgr.namespace, i, err.Error())
				return err
			}
			if err, _ := edb.PersistTransaction(avBatch, tx, false, false); err != nil {
				mgr.logger.Errorf("[Namespace = %s] archive tx in block %d to historic database failed, error msg %s", mgr.namespace, i, err.Error())
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
			}
			if err, _ := edb.PersistReceipt(avBatch, receipt, false, false); err != nil {
				mgr.logger.Errorf("[Namespace = %s] archive receipt in block %d to historic database failed, error msg %s", mgr.namespace, i, err.Error())
				return err
			}
		}
		// delete block
		if err := edb.DeleteBlockByNum(mgr.namespace, olBatch, i, false, false); err != nil {
			mgr.logger.Errorf("[Namespace = %s] archive useless block %d failed, error msg %s", mgr.namespace, i, err.Error())
			return err
		}

		if err, _ := edb.PersistBlock(avBatch, block, false, false); err != nil {
			mgr.logger.Errorf("[Namespace = %s] archive block %d to historic database failed. error msg %s", mgr.namespace, i, err.Error())
			return err
		}
	}
	if err := edb.DeleteJournalInRange(olBatch, curGenesis, manifest.Height, false, false); err != nil {
		mgr.logger.Errorf("[Namespace = %s] archive useless journals failed, error msg %s", mgr.namespace,  err.Error())
		return err
	}
	// update chain
	if err := edb.UpdateGenesisTag(mgr.namespace, manifest.Height, olBatch, false, false); err != nil {
		mgr.logger.Errorf("[Namespace = %s] update chain genesis field failed, error msg %s", mgr.namespace,  err.Error())
		return err
	}
	if err := avBatch.Write(); err != nil {
		mgr.logger.Errorf("[Namespace = %s] flush to historic database failed, error msg %s", mgr.namespace,  err.Error())
		return err
	}
	if err := olBatch.Write(); err != nil {
		mgr.logger.Errorf("[Namespace = %s] flush to online database failed, error msg %s", mgr.namespace,  err.Error())
		return err
	}
	return nil
}

func (mgr *ArchiveManager) feedback(isSuccess bool, filterId string, message string) {
	mgr.executor.sendFilterEvent(FILTER_ARCHIVE, isSuccess, filterId, message)
}
