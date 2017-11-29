//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package rbft

// mesage type
const (
	// 1.rbft core
	CORE_BATCH_TIMER_EVENT = iota
	CORE_NULL_REQUEST_TIMER_EVENT
	CORE_FIRST_REQUEST_TIMER_EVENT
	CORE_STATE_UPDATE_EVENT
	CORE_VALIDATED_TXS_EVENT
	CORE_EXECUTOR_CHECKPOINT_EVENT

	// 2.view change
	VIEW_CHANGE_TIMER_EVENT
	VIEW_CHANGE_RESEND_TIMER_EVENT
	VIEW_CHANGE_QUORUM_EVENT
	VIEW_CHANGED_EVENT
	VIEW_CHANGE_VC_RESET_DONE_EVENT

	// 3.recovery
	RECOVERY_RESTART_TIMER_EVENT
	RECOVERY_DONE_EVENT
	RECOVERY_NEGO_VIEW_RSP_TIMER_EVENT
	RECOVERY_NEGO_VIEW_DONE_EVENT

	// 4.node mgr service
	NODE_MGR_NEW_NODE_EVENT
	NODE_MGR_ADD_NODE_EVENT
	NODE_MGR_DEL_NODE_EVENT
	NODE_MGR_AGREE_UPDATE_QUORUM_EVENT
	NODE_MGR_UPDATED_EVENT
)

// service type
const (
	CORE_RBFT_SERVICE = iota
	VIEW_CHANGE_SERVICE
	RECOVERY_SERVICE
	NODE_MGR_SERVICE
	NOT_SUPPORT_SERVICE
)

// LocalEvent represents event sent by local modules
type LocalEvent struct {
	Service   int // service type range from {CORE_RBFT_SERVICE, VIEW_CHANGE_SERVICE, RECOVERY_SERVICE, NODE_MGR_SERVICE}
	EventType int
	Event     interface{}
}

// Event is a type meant to clearly convey that the return type or parameter to a function will be supplied to/from an events.Manager
type consensusEvent interface{}
