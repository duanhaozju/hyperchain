package api

import (
	"hyperchain/manager"
	"hyperchain/common"
	"hyperchain/manager/event"
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

func (admin *AdminPublicAPI) Snapshot(blockNumber uint64) {
	log := common.GetLogger(admin.namespace, "api")
	log.Debugf("receive snapshot rpc command, params: (block number #%d)", blockNumber)
	admin.eh.GetEventObject().Post(event.SnapshotEvent{BlockNumber: blockNumber})
}
