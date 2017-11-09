package executor

import (
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/op/go-logging"
	"hyperchain/common"
	pb "hyperchain/common/protos"
	"hyperchain/common/service"
	"hyperchain/common/service/server"
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

	DeleteSnapshot(ev event.DeleteSnapEvent) error

	Archive(ev event.ArchEvent) error

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
	conf      *common.Config
}

func NewRemoteExecutorProxy(is *server.InternalServer, config *common.Config) IExecutor {
	rep := &remoteExecutorProxy{
		is:        is,
		namespace: config.GetString(common.NAMESPACE),
		conf:      config,
	}
	rep.logger = common.GetLogger(rep.namespace, "executor")
	return rep
}

func (re *remoteExecutorProxy) Start() error {
	executorHostAddr := re.conf.GetString(common.EXECUTOR_HOST_ADDR)
	if len(executorHostAddr) == 0 {
		return fmt.Errorf("No executor host addr found for this executor ")
	}
	adminSrv := re.is.ServerRegistry().AdminService(executorHostAddr)
	if adminSrv == nil {
		return fmt.Errorf("No executor admin found for %s ", executorHostAddr)
	}

	ane := event.AddNamespaceEvent{
		Namespace: re.namespace,
	}

	payload, _ := proto.Marshal(&ane)

	msg := &pb.IMessage{
		Type:    pb.Type_SYNC_REQUEST,
		Event:   pb.Event_AddNamespaceEvent,
		Payload: payload,
	}

	rsp, err := adminSrv.SyncSend(msg)

	//err := adminSrv.Send(msg)
	//if err != nil {
	//	return err
	//}
	//
	//rsp := <-adminSrv.Response()

	if err != nil {
		return err
	}

	if rsp.Type == pb.Type_RESPONSE && rsp.Ok == true {
		return nil
	} else {
		//TODO : parse error info  form rsp.Payload
		return fmt.Errorf("Start executor failed, %v ", rsp.Payload)
	}
}

func (re *remoteExecutorProxy) Validate(ve event.ValidationEvent) {
	var err error
	defer func() { re.handleError(err) }()
	msg := &pb.IMessage{
		Type:  pb.Type_EVENT,
		From:  pb.FROM_EVENTHUB,
		Event: pb.Event_ValidationEvent,
	}

	payload, err := proto.Marshal(&ve)
	if err != nil {
		return
	}

	msg.Payload = payload
	err = re.sendToExecutor(re.namespace, msg)
}

func (re *remoteExecutorProxy) CommitBlock(ce event.CommitEvent) {
	var err error
	defer func() { re.handleError(err) }()
	msg := &pb.IMessage{
		Type:  pb.Type_EVENT,
		From:  pb.FROM_EVENTHUB,
		Event: pb.Event_CommitEvent,
	}

	payload, err := proto.Marshal(&ce)
	if err != nil {
		return
	}

	msg.Payload = payload
	err = re.sendToExecutor(re.namespace, msg)
}

func (re *remoteExecutorProxy) RunInSandBox(tx *types.Transaction, snapshotId string) error {
	var err error
	defer func() { re.handleError(err) }()

	return nil
}

func (re *remoteExecutorProxy) Rollback(ev event.VCResetEvent) {
	var err error
	defer func() { re.handleError(err) }()

	msg := &pb.IMessage{
		Type:  pb.Type_EVENT,
		From:  pb.FROM_EVENTHUB,
		Event: pb.Event_VCResetEvent,
	}

	payload, err := proto.Marshal(&ev)
	if err != nil {
		return
	}

	msg.Payload = payload
	err = re.sendToExecutor(re.namespace, msg)
}

func (re *remoteExecutorProxy) SyncChain(ev event.ChainSyncReqEvent) {
	var err error
	defer func() { re.handleError(err) }()

	msg := &pb.IMessage{
		Type:  pb.Type_EVENT,
		From:  pb.FROM_EVENTHUB,
		Event: pb.Event_ChainSyncReqEvent,
	}

	payload, err := proto.Marshal(&ev)
	if err != nil {
		return
	}

	msg.Payload = payload
	err = re.sendToExecutor(re.namespace, msg)
}

func (re *remoteExecutorProxy) Snapshot(ev event.SnapshotEvent) {
	var err error
	defer func() { re.handleError(err) }()

	msg := &pb.IMessage{
		Type:  pb.Type_EVENT,
		From:  pb.FROM_EVENTHUB,
		Event: pb.Event_SnapshotEvent,
	}

	payload, err := proto.Marshal(&ev)
	if err != nil {
		return
	}

	msg.Payload = payload
	err = re.sendToExecutor(re.namespace, msg)
}

func (re *remoteExecutorProxy) DeleteSnapshot(ev event.DeleteSnapEvent) error {
	var err error
	defer func() { re.handleError(err) }()

	//TODO: check the logic.
	msg := &pb.IMessage{
		Type:  pb.Type_EVENT,
		From:  pb.FROM_EVENTHUB,
		Event: pb.Event_DeleteSnapshotEvent,
	}

	payload, err := proto.Marshal(&ev)
	if err != nil {
		return err
	}

	msg.Payload = payload
	err = re.sendToExecutor(re.namespace, msg)
	return err
}

func (re *remoteExecutorProxy) Archive(ev event.ArchEvent) error {
	var err error
	defer func() { re.handleError(err) }()

	//TODO: check the logic.
	msg := &pb.IMessage{
		Type:  pb.Type_EVENT,
		From:  pb.FROM_EVENTHUB,
		Event: pb.Event_ArchiveEvent,
	}
	payload, err := proto.Marshal(&ev)
	if err != nil {
		return err
	}

	msg.Payload = payload
	err = re.sendToExecutor(re.namespace, msg)
	return err
}

func (re *remoteExecutorProxy) StoreInvalidTransaction(payload []byte) {
	var err error
	defer func() { re.handleError(err) }()
	//wait for store successful ?
	msg := &pb.IMessage{
		Type:  pb.Type_EVENT,
		From:  pb.FROM_EVENTHUB,
		Event: pb.Event_StoreInvalidTransactionEvent,
	}
	msg.Payload = payload
	err = re.sendToExecutor(re.namespace, msg)
}

func (re *remoteExecutorProxy) ReceiveReplicaInfo(payload []byte) {
	var err error
	defer func() { re.handleError(err) }()

	msg := &pb.IMessage{
		Type:  pb.Type_EVENT,
		From:  pb.FROM_EVENTHUB,
		Event: pb.Event_ReceiveReplicaInfoEvent,
	}
	msg.Payload = payload
	err = re.sendToExecutor(re.namespace, msg)
}

func (re *remoteExecutorProxy) ReceiveSyncBlocks(payload []byte) {
	var err error
	defer func() { re.handleError(err) }()

	msg := &pb.IMessage{
		Type:  pb.Type_EVENT,
		From:  pb.FROM_EVENTHUB,
		Event: pb.Event_ReceiveSyncBlocksEvent,
	}
	msg.Payload = payload
	err = re.sendToExecutor(re.namespace, msg)
}

func (re *remoteExecutorProxy) GetNVP() NVP {
	//TODO: nvp?
	return nil
}

func (re *remoteExecutorProxy) ReceiveSyncRequest(payload []byte) {
	var err error
	defer func() { re.handleError(err) }()
	msg := &pb.IMessage{
		Type:  pb.Type_EVENT,
		From:  pb.FROM_EVENTHUB,
		Event: pb.Event_ReceiveSyncRequestEvent,
	}
	msg.Payload = payload
	err = re.sendToExecutor(re.namespace, msg)
}

func (re *remoteExecutorProxy) ReceiveWorldStateSyncRequest(payload []byte) {
	var err error
	defer func() { re.handleError(err) }()

	msg := &pb.IMessage{
		Type:  pb.Type_EVENT,
		From:  pb.FROM_EVENTHUB,
		Event: pb.Event_ReceiveWorldStateSyncRequestEvent,
	}
	msg.Payload = payload
	err = re.sendToExecutor(re.namespace, msg)
}

func (re *remoteExecutorProxy) ReceiveWorldState(payload []byte) {
	var err error
	defer func() { re.handleError(err) }()

	msg := &pb.IMessage{
		Type:  pb.Type_EVENT,
		From:  pb.FROM_EVENTHUB,
		Event: pb.Event_ReceiveWorldStateEvent,
	}
	msg.Payload = payload
	err = re.sendToExecutor(re.namespace, msg)
}

func (re *remoteExecutorProxy) ReceiveWsHandshake(payload []byte) {
	var err error
	defer func() { re.handleError(err) }()

	msg := &pb.IMessage{
		Type:  pb.Type_EVENT,
		From:  pb.FROM_EVENTHUB,
		Event: pb.Event_ReceiveWsHandshakeEvent,
	}
	msg.Payload = payload
	err = re.sendToExecutor(re.namespace, msg)
}

func (re *remoteExecutorProxy) ReceiveWsAck(payload []byte) {
	var err error
	defer func() { re.handleError(err) }()

	msg := &pb.IMessage{
		Type:  pb.Type_EVENT,
		From:  pb.FROM_EVENTHUB,
		Event: pb.Event_ReceiveWsAckEvent,
	}
	msg.Payload = payload
	err = re.sendToExecutor(re.namespace, msg)
}

func (re *remoteExecutorProxy) CreateInitBlock(config *common.Config) error {
	var err error
	defer func() { re.handleError(err) }()
	err = fmt.Errorf("Executor run in distributed mode, should not invoke CreateInitBlock this method! ")
	return nil
}

func (re *remoteExecutorProxy) FetchStateDb() vm.Database {
	var err error
	defer func() { re.handleError(err) }()
	err = fmt.Errorf("Executor run in distributed mode, should not invoke FetchStateDb this method! ")
	return nil
}

func (re *remoteExecutorProxy) Stop() error {
	var err error
	defer func() { re.handleError(err) }()
	executorHostAddr := re.conf.GetString(common.EXECUTOR_HOST_ADDR)
	if len(executorHostAddr) == 0 {
		return fmt.Errorf("No executor host addr found for this executor ")
	}
	adminSrv := re.is.ServerRegistry().AdminService(executorHostAddr)
	if adminSrv == nil {
		return fmt.Errorf("No executor admin found for %s ", executorHostAddr)
	}

	ane := event.DeleteNamespaceEvent{
		Namespace: re.namespace,
	}

	payload, _ := proto.Marshal(&ane)

	msg := &pb.IMessage{
		Type:    pb.Type_SYNC_REQUEST,
		Event:   pb.Event_DeleteNamespaceEvent,
		Payload: payload,
	}

	rsp, err := adminSrv.SyncSend(msg)
	if err != nil {
		return err
	}

	//rsp := <-adminSrv.Response()
	if rsp.Type == pb.Type_RESPONSE && rsp.Ok == true {
		return nil
	} else {
		//TODO : parse error info  form rsp.Payload
		return fmt.Errorf("Start executor failed, %v ", rsp.Payload)
	}
	return nil
}

//sendToExecutor send message to executor by namespace.
func (re *remoteExecutorProxy) sendToExecutor(namespace string, msg *pb.IMessage) error {
	var err error
	defer func() { re.handleError(err) }()

	ns := re.is.ServerRegistry().Namespace(namespace)
	if ns == nil {
		return fmt.Errorf("No services found for namespace %s ", namespace)
	}

	srv := ns.Service(fmt.Sprintf("EXECUTOR-%d", 0))
	// TODO: fix it, executor should be config in the config file
	if srv == nil {
		return fmt.Errorf("No service found for %s ", service.EXECUTOR)
	}

	if err = srv.Send(msg); err != nil {
		return err
	}else {
		return nil
	}
}

//handleError handle all kind of errors here.
func (re *remoteExecutorProxy) handleError(err error) {
	if err != nil {
		re.logger.Error(err)
	}
}
