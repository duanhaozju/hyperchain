package api

import (
	"hyperchain/manager"
	"hyperchain/common"
	"hyperchain/manager/event"
	flt "hyperchain/manager/filter"
	edb "hyperchain/core/db_utils"
)

type AdminPublicAPI struct {
	eh        *manager.EventHub
	namespace string
	config    *common.Config
	isPublic  bool
}


func NewPublicAdminAPI(namespace string, eh *manager.EventHub, config *common.Config) *AdminPublicAPI {
	return &AdminPublicAPI{
		namespace: namespace,
		eh:        eh,
		config:    config,
		isPublic:  true,
	}
}

func (admin *AdminPublicAPI) Snapshot(blockNumber uint64) string {
	log := common.GetLogger(admin.namespace, "api")
	filterId := flt.NewFilterID()
	log.Debugf("receive snapshot rpc command, params: (block number #%d), filterId: (%s)", blockNumber, filterId)
	admin.eh.GetEventObject().Post(event.SnapshotEvent{
		FilterId:    filterId,
		BlockNumber: blockNumber,
	})
	return filterId
}
// TODO snapshot callback

func (admin *AdminPublicAPI) ReadSnapshot(filterId string, verbose bool) (interface{}, error) {
	manifestHandler := common.NewManifestHandler(getManifestPath(admin.config))
	var manifest common.Manifest
	var err error
	if err, manifest = manifestHandler.Read(filterId); err != nil {
		return nil, &common.SnapshotErr{Message: err.Error()}
	}
	if !verbose {
		return manifest, nil
	} else {
		// return whole world state
		stateDb, err := NewSnapshotStateDb(admin.config, manifest.FilterId, common.Hex2Bytes(manifest.MerkleRoot), manifest.Height, manifest.Namespace)
		if err != nil {
			return nil, &common.SnapshotErr{Message: err.Error()}
		}
		// TODO use a more pretty display mode
		return string(stateDb.Dump()), nil
	}
}

func (admin *AdminPublicAPI) ListSnapshot() (common.Manifests, error) {
	manifestHandler := common.NewManifestHandler(getManifestPath(admin.config))
	if err, manifests := manifestHandler.List(); err != nil {
		return nil, &common.SnapshotErr{Message: err.Error()}
	} else {
		return manifests, nil
	}
}

func (admin *AdminPublicAPI) DeleteSnapshot(filterId string) string {
	log := common.GetLogger(admin.namespace, "api")
	log.Debugf("receive delete snapshot rpc command, filterId: (%s)", filterId)
	admin.eh.GetEventObject().Post(event.DeleteSnapshotEvent{
		FilterId:    filterId,
	})
	return filterId
}
// TODO delete snapshot callback

func (admin *AdminPublicAPI) CheckSnapshot(filterId string) (bool, error) {
	manifestHandler := common.NewManifestHandler(getManifestPath(admin.config))
	var manifest common.Manifest
	var err error
	if err, manifest = manifestHandler.Read(filterId); err != nil {
		return false, &common.SnapshotErr{Message: err.Error()}
	}
	// return whole world state
	stateDb, err := NewSnapshotStateDb(admin.config, manifest.FilterId, common.Hex2Bytes(manifest.MerkleRoot), manifest.Height, manifest.Namespace)
	if err != nil {
		return false, &common.SnapshotErr{Message: err.Error()}
	}
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

func (admin *AdminPublicAPI) Archive(filterId string) (string) {
	log := common.GetLogger(admin.namespace, "api")
	log.Debugf("receive archive command, params: filterId: (%s)", filterId)
	admin.eh.GetEventObject().Post(event.ArchiveEvent{
		FilterId:    filterId,
	})
	return filterId

}
