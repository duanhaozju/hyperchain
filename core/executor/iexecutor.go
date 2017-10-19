package executor

import (
	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/core/vm"
	"hyperchain/manager/event"
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

//proxy for remote executor.
type remoteExecutorProxy struct {
}

func NewRemoteExecutorProxy() IExecutor {
	return &remoteExecutorProxy{}
}

func (re *remoteExecutorProxy) Validate(ve event.ValidationEvent) {

}

func (re *remoteExecutorProxy) CommitBlock(ce event.CommitEvent) {

}

func (re *remoteExecutorProxy) RunInSandBox(tx *types.Transaction, snapshotId string) error {

	return nil
}

func (re *remoteExecutorProxy) Rollback(ev event.VCResetEvent) {

}

func (re *remoteExecutorProxy) SyncChain(ev event.ChainSyncReqEvent) {

}

func (re *remoteExecutorProxy) Snapshot(ev event.SnapshotEvent) {

}

func (re *remoteExecutorProxy) DeleteSnapshot(ev event.DeleteSnapshotEvent) {

}

func (re *remoteExecutorProxy) Archive(event event.ArchiveEvent) {

}

func (re *remoteExecutorProxy) StoreInvalidTransaction(payload []byte) {

}

func (re *remoteExecutorProxy) ReceiveReplicaInfo(payload []byte) {

}

func (re *remoteExecutorProxy) ReceiveSyncBlocks(payload []byte) {

}

func (re *remoteExecutorProxy) GetNVP() NVP {
	return nil
}

func (re *remoteExecutorProxy) ReceiveSyncRequest(payload []byte) {

}

func (re *remoteExecutorProxy) ReceiveWorldStateSyncRequest(payload []byte) {

}

func (re *remoteExecutorProxy) ReceiveWorldState(payload []byte) {

}

func (re *remoteExecutorProxy) ReceiveWsHandshake(payload []byte) {

}

func (re *remoteExecutorProxy) ReceiveWsAck(payload []byte) {}

func (re *remoteExecutorProxy) CreateInitBlock(config *common.Config) error {
	return nil
}

func (re *remoteExecutorProxy) FetchStateDb() vm.Database {
	return nil
}

func (re *remoteExecutorProxy) Start() error {
	return nil
}

func (re *remoteExecutorProxy) Stop() error {
	return nil
}
