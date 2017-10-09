package api

import (
	"hyperchain/common"
	edb "hyperchain/core/db_utils"
	"hyperchain/manager"
	"hyperchain/manager/event"
	flt "hyperchain/manager/filter"
)

type ArchivePublicAPI struct {
	eh        *manager.EventHub
	namespace string
	config    *common.Config
	isPublic  bool
}

func NewPublicArchiveAPI(namespace string, eh *manager.EventHub, config *common.Config) *ArchivePublicAPI {
	return &ArchivePublicAPI{
		namespace: namespace,
		eh:        eh,
		config:    config,
		isPublic:  true,
	}
}

func (admin *ArchivePublicAPI) Snapshot(blockNumber uint64) string {
	log := common.GetLogger(admin.namespace, "api")
	filterId := flt.NewFilterID()
	log.Debugf("receive snapshot rpc command, params: (block number #%d), filterId: (%s)", blockNumber, filterId)
	admin.eh.GetEventObject().Post(event.SnapshotEvent{
		FilterId:    filterId,
		BlockNumber: blockNumber,
	})
	return filterId
}

func (admin *ArchivePublicAPI) QuerySnapshotExist(filterId string) bool {
	manifestHandler := common.NewManifestHandler(common.GetPath(admin.namespace, getManifestPath(admin.config)))
	if manifestHandler.Contain(filterId) {
		return true
	} else {
		return false
	}
}

func (admin *ArchivePublicAPI) ReadSnapshot(filterId string) (interface{}, error) {
	manifestHandler := common.NewManifestHandler(common.GetPath(admin.namespace, getManifestPath(admin.config)))
	var manifest common.Manifest
	var err error
	if err, manifest = manifestHandler.Read(filterId); err != nil {
		return nil, &common.SnapshotErr{Message: err.Error()}
	}
	return manifest, nil
}

func (admin *ArchivePublicAPI) ListSnapshot() (common.Manifests, error) {
	manifestHandler := common.NewManifestHandler(common.GetPath(admin.namespace, getManifestPath(admin.config)))
	if err, manifests := manifestHandler.List(); err != nil {
		return nil, &common.SnapshotErr{Message: err.Error()}
	} else {
		return manifests, nil
	}
}

func (admin *ArchivePublicAPI) DeleteSnapshot(filterId string) (bool, error) {
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

func (admin *ArchivePublicAPI) CheckSnapshot(filterId string) (bool, error) {
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

func (admin *ArchivePublicAPI) Archive(filterId string, sync bool) (bool, error) {
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

func (admin *ArchivePublicAPI) QueryArchiveResult(filterId string) (bool, error) {
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
