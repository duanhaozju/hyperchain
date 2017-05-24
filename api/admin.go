package api

import (
	"hyperchain/manager"
	"hyperchain/common"
	"hyperchain/manager/event"
	flt "hyperchain/manager/filter"
)

type AdminPublicAPI struct {
	eh        *manager.EventHub
	namespace string
	config    *common.Config
	isPublic  bool
}

type Manifest struct {
	Height     uint64    `json:"height"`
	FilterId   string    `json:"filterId"`
	MerkleRoot string    `json:"merkleRoot"`
	Date       string    `json:"date"`
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

func (admin *AdminPublicAPI) ReadSnapshot(filterId string) (common.Manifest, error) {
	manifestHandler := common.NewManifestHandler(getManifestPath(admin.config))
	if err, manifest := manifestHandler.Read(filterId); err != nil {
		return common.Manifest{}, &common.SnapshotErr{Message: err.Error()}
	} else {
		return manifest, nil
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
