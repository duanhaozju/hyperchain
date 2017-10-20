package executor

import (
	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/manager/event"
	"hyperchain/core/vm"
)

type IExecutor interface {
	Validate(validationEvent event.ValidationEvent)
	CommitBlock(ev event.CommitEvent)
	RunInSandBox(tx *types.Transaction, snapshotId string) error
	Rollback(ev event.VCResetEvent)
	SyncChain(ev event.ChainSyncReqEvent)
	Snapshot(ev event.SnapshotEvent)
	DeleteSnapshot(ev event.DeleteSnapshotEvent)
	Archive(event event.ArchiveEvent)
	StoreInvalidTransaction(payload []byte)
	ReceiveReplicaInfo(payload []byte)
	ReceiveSyncBlocks(payload []byte)
	GetNVP() NVP
	ReceiveSyncRequest(payload []byte)
	ReceiveWorldStateSyncRequest(payload []byte)
	ReceiveWorldState(payload []byte)
	ReceiveWsHandshake(payload []byte)
	ReceiveWsAck(payload []byte)
	CreateInitBlock(config *common.Config) error
	FetchStateDb() vm.Database
	Start() error
	Stop() error
}

type remoteExecutor struct {}
