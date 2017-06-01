//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

// returnRequestBatchEvent is sent by pbft when we are forwarded a request
//type returnRequestBatchEvent *TransactionBatch

// stateUpdatedEvent  when stateUpdate is executed and return the result
type stateUpdatedEvent struct {
	seqNo uint64
}

type checkpointMessage struct {
	seqNo uint64
	id    []byte
}

type replicaInfo struct {
	id      uint64
	height  uint64
	genesis uint64
}

type stateUpdateTarget struct {
	checkpointMessage
	replicas []replicaInfo
}

//mesage type
const (
	//1.pbft core
	CORE_BATCH_TIMER_EVENT = iota
	CORE_NULL_REQUEST_TIMER_EVENT
	CORE_FIRST_REQUEST_TIMER_EVENT
	CORE_CLEAR_DUPLICATOR_EVENT
	CORE_REMOVE_CACHE_EVENT
	CORE_STATE_UPDATE_EVENT
	CORE_VALIDATED_TXS_EVENT

	//2.view change
	VIEW_CHANGE_TIMER_EVENT
	VIEW_CHANGE_RESEND_TIMER_EVENT
	VIEW_CHANGE_QUORUM_EVENT
	VIEW_CHANGED_EVENT
	VIEW_CHANGE_VC_RESET_DONE_EVENT

	//3.recovery
	RECOVERY_RESTART_TIMER_EVENT
	RECOVERY_DONE_EVENT
	RECOVERY_NEGO_VIEW_RSP_TIMER_EVENT
	RECOVERY_NEGO_VIEW_DONE_EVENT

	//4.node mgr service
	NODE_MGR_NEW_NODE_EVENT
	NODE_MGR_ADD_NODE_EVENT
	NODE_MGR_DEL_NODE_EVENT
	NODE_MGR_AGREE_UPDATEN_QUORUM_EVENT
	NODE_MGR_UPDATEDN_EVENT
)

//service type
const (
	CORE_PBFT_SERVICE = iota
	VIEW_CHANGE_SERVICE
	RECOVERY_SERVICE
	NODE_MGR_SERVICE
	NOT_SUPPORT_SERVICE
)

//LocalEvent represent event send by local modules
type LocalEvent struct {
	Service   int
	EventType int
	Event     interface{}
}
