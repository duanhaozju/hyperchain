package executor

import (
	"github.com/gogo/protobuf/proto"
	"github.com/op/go-logging"
	"hyperchain/common"
	pb "hyperchain/common/protos"
	"hyperchain/common/service"
	"hyperchain/core/types"
	"hyperchain/core/vm"
	"hyperchain/manager/event"
	"hyperchain/common/service/server"
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
	namespace string
	is        *server.InternalServer
	logger    *logging.Logger
}

func NewRemoteExecutorProxy(is *server.InternalServer, config *common.Config) IExecutor {
	rep := &remoteExecutorProxy{
		is:        is,
		namespace: config.GetString(common.NAMESPACE),
	}
	rep.logger = common.GetLogger(rep.namespace, "executor")
	return rep
}

func (re *remoteExecutorProxy) Start() error {
	//TODO: wait until this namespace registered
	return nil
}

func (re *remoteExecutorProxy) Validate(ve event.ValidationEvent) {
	msg := &pb.IMessage{
		Type:  pb.Type_EVENT,
		From:  pb.FROM_EVENTHUB,
		Event: pb.Event_ValidationEvent,
	}

	payload, err := proto.Marshal(&ve)
	if err != nil {
		//TODO: handle error
	}

	msg.Payload = payload
	re.is.ServerRegistry().Namespace(re.namespace).Service(service.EXECUTOR).Send(msg)
}

func (re *remoteExecutorProxy) CommitBlock(ce event.CommitEvent) {
	msg := &pb.IMessage{
		Type:  pb.Type_EVENT,
		From:  pb.FROM_EVENTHUB,
		Event: pb.Event_CommitEvent,
	}

	payload, err := proto.Marshal(&ce)
	if err != nil {
		//TODO: handle error
	}

	msg.Payload = payload
	re.is.ServerRegistry().Namespace(re.namespace).Service(service.EXECUTOR).Send(msg)
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

func (re *remoteExecutorProxy) Stop() error {
	return nil
}