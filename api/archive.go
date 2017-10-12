package api

import (
	"hyperchain/common"
	edb "hyperchain/core/db_utils"
	"hyperchain/manager"
	"hyperchain/manager/event"
	flt "hyperchain/manager/filter"
)

// This file implements the handler of Archive service API which
// can be invoked by client in JSON-RPC request.

type Archive struct {
	eh        *manager.EventHub
	namespace string
	config    *common.Config
	isPublic  bool
}

// NewPublicArchiveAPI creates and returns a new Archive instance for given namespace name.
func NewPublicArchiveAPI(namespace string, eh *manager.EventHub, config *common.Config) *Archive {
	return &Archive{
		namespace: namespace,
		eh:        eh,
		config:    config,
		isPublic:  true,
	}
}

// Snapshot makes the snapshot for given block number. It returns the snapshot id
// for the client to query.
func (admin *Archive) Snapshot(blockNumber uint64) (string, error) {
	log := common.GetLogger(admin.namespace, "api")
	handler := common.NewManifestHandler(common.GetPath(admin.namespace, getManifestPath(admin.config)))

	chainHeight := edb.GetHeightOfChain(admin.namespace)
	if blockNumber < chainHeight && blockNumber != 0 {
		return "", &common.SnapshotErr{Message: "trigger block number is less than chain height"}
	}
	if _, meta := handler.Search(chainHeight); (meta != common.Manifest{}) && blockNumber == 0 {
		return "", &common.SnapshotErr{Message: "duplicate snapshot requirement for same height"}
	}
	if _, meta := handler.Search(blockNumber); (meta != common.Manifest{}) && blockNumber != 0 {
		return "", &common.SnapshotErr{Message: "duplicate snapshot requirement for same height"}
	}

	filterId := flt.NewFilterID()
	log.Debugf("receive snapshot rpc command, params: (block number #%d), filterId: (%s)", blockNumber, filterId)
	admin.eh.GetEventObject().Post(event.SnapshotEvent{
		FilterId:    filterId,
		BlockNumber: blockNumber,
	})
	return filterId, nil
}

// QuerySnapshotExist checks if the given snapshot existed, so you can confirm that
// the last step Archive.Snapshot is successful.
func (admin *Archive) QuerySnapshotExist(filterId string) bool {
	manifestHandler := common.NewManifestHandler(common.GetPath(admin.namespace, getManifestPath(admin.config)))
	if manifestHandler.Contain(filterId) {
		return true
	} else {
		return false
	}
}

// ReadSnapshot returns the snapshot information for the given snapshot ID.
func (admin *Archive) ReadSnapshot(filterId string) (interface{}, error) {
	manifestHandler := common.NewManifestHandler(common.GetPath(admin.namespace, getManifestPath(admin.config)))
	var manifest common.Manifest
	var err error
	if err, manifest = manifestHandler.Read(filterId); err != nil {
		return nil, &common.SnapshotErr{Message: err.Error()}
	}
	return manifest, nil
}

// ListSnapshot returns all the existed snapshot information.
func (admin *Archive) ListSnapshot() (common.Manifests, error) {
	manifestHandler := common.NewManifestHandler(common.GetPath(admin.namespace, getManifestPath(admin.config)))
	if err, manifests := manifestHandler.List(); err != nil {
		return nil, &common.SnapshotErr{Message: err.Error()}
	} else {
		return manifests, nil
	}
}

// DeleteSnapshot deletes snapshot under the given snapshot ID.
func (admin *Archive) DeleteSnapshot(filterId string) (bool, error) {
	log := common.GetLogger(admin.namespace, "api")
	log.Debugf("receive delete snapshot rpc command, filterId: (%s)", filterId)
	cont := make(chan error)
	admin.eh.GetEventObject().Post(event.DeleteSnapshotEvent{
		FilterId: filterId,
		Cont:     cont,
	})
	err := <-cont
	if err != nil {
		return false, &common.SnapshotErr{Message: err.Error()}
	} else {
		return true, nil
	}
}

// CheckSnapshot will check that the snapshot is correct. If correct, returns true.
// Otherwise, returns false.
func (admin *Archive) CheckSnapshot(filterId string) (bool, error) {
	manifestHandler := common.NewManifestHandler(common.GetPath(admin.namespace, getManifestPath(admin.config)))
	var manifest common.Manifest
	var err error
	if err, manifest = manifestHandler.Read(filterId); err != nil {
		return false, &common.SnapshotErr{Message: err.Error()}
	}
	// return whole world state
	stateDb, closer, err := NewSnapshotStateDb(admin.config, manifest.FilterId, common.Hex2Bytes(manifest.MerkleRoot), manifest.Height, manifest.Namespace)
	if err != nil {
		return false, &common.SnapshotErr{Message: err.Error()}
	}
	defer closer()
	curHash, err := stateDb.RecomputeCryptoHash()
	if err != nil {
		return false, &common.SnapshotErr{Message: err.Error()}
	}
	blk, err := edb.GetBlockByNumber(admin.namespace, manifest.Height)
	if err != nil {
		return false, &common.SnapshotErr{Message: err.Error()}
	}
	if curHash != common.HexToHash(manifest.MerkleRoot) || curHash != common.BytesToHash(blk.MerkleRoot) {
		return false, nil
	}
	return true, nil
}

// Archive will archive data of the given snapshot. If successful, returns true.
func (admin *Archive) Archive(filterId string, sync bool) (bool, error) {
	log := common.GetLogger(admin.namespace, "api")
	log.Debugf("receive archive command, params: filterId: (%s)", filterId)
	cont := make(chan error)
	admin.eh.GetEventObject().Post(event.ArchiveEvent{
		FilterId: filterId,
		Cont:     cont,
		Sync:     sync,
	})
	err := <-cont
	if err != nil {
		return false, &common.SnapshotErr{Message: err.Error()}
	} else {
		return true, nil
	}
}

// QueryArchiveExist checks if the given snapshot has been archived.
func (admin *Archive) QueryArchiveExist(filterId string) (bool, error) {
	manifestHandler := common.NewManifestHandler(common.GetPath(admin.namespace, getManifestPath(admin.config)))
	var manifest common.Manifest
	var err error
	if err, manifest = manifestHandler.Read(filterId); err != nil {
		return false, &common.SnapshotErr{Message: err.Error()}
	}
	err, genesis := edb.GetGenesisTag(admin.namespace)
	if err != nil {
		return false, &common.SnapshotErr{Message: err.Error()}
	}
	return genesis == manifest.Height, nil
}
