package executor

import "hyperchain/common"

type ArchiveManager struct {
	executor *Executor
	registry *SnapshotRegistry
}

func NewArchiveManager(executor *Executor, registry *SnapshotRegistry) *ArchiveManager {
	return &ArchiveManager{
		executor:  executor,
		registry:  registry,
	}
}

/*
	External Functions
 */
func (mgr *ArchiveManager) Archive(filterId string) {
	//var err error
	//var manifest common.Manifest
	//if !mgr.registry.rwc.Contain(filterId) {
	//	// feedback
	//} else {
	//	err, manifest = mgr.registry.rwc.Read(filterId)
	//	if err != nil {
	//		//
	//	}
	//
	//}
}

/*
	Internal Functions
 */
func (mgr *ArchiveManager) migrate(manifest common.Manifest) {

}
