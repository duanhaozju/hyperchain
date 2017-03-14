//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package pbft

/**
	manage the constants for pbft algoriithm
 */

const (
// constant timer names

	BATCH_TIMER = "batchTimer" //
	VALIDATE_TIMER = "validateTimer" //

	VC_RESEND_TIMER = "vcResendTimer" //timer triggering resend of a view change
	NEW_VIEW_TIMER = "newViewTimer"   //timer triggering view change by out of timeout of some requestBatch

	NULL_REQUEST_TIMER = "nullRequestTimer" //timer triggering send null request, used for heartbeat
	FIRST_REQUEST_TIMER = "firstRequestTimer" //timer set for replicas in case of primary start and shut down immediately

	NEGO_VIEW_RSP_TIMER = "negoViewRspTimer" //timer track timeout for N-f nego-view responses
	RECOVERY_RESTART_TIMER = "recoveryRestartTimer" // recoveryRestartTimer track how long a recovery is finished and fires if needed

	ADD_NODE_TIMER = "addNodeTimer" //track timeout for del node responses
	DEL_NODE_TIMER = "delNodeTimer" //track timeout for new node responses
	UPDATE_TIMER = "updateTimer"	// track timeout for N-f agree on update n

// constant PBFT configuration keys

	PBFT_NODE_NUM = "pbft.nodes"
	PBFT_BATCH_SIZE = "pbft.batchsize"
	PBFT_VC_RESEND_LIMIT = "pbft.vcresendlimit"

	PBFT_NEGOVIEW_TIMEOUT = "timeout.negoview"
	PBFT_RECOVERY_TIMEOUT = "timeout.recovery"
	PBFT_VIEWCHANGE_TIMEOUT = "timeout.viewchange"
	PBFT_RESEND_VIEWCHANGE_TIMEOUT = "timeout.resendviewchange"

	PBFT_BATCH_TIMEOUT = "timeout.batch"
	PBFT_REQUEST_TIMEOUT = "timeout.request"
	PBFT_VALIDATE_TIMEOUT = "timeout.validate"

	PBFT_NULLREQUEST_TIMEOUT = "timeout.nullrequest"
	PBFT_FIRST_REQUEST_TIMEOUT = "timeout.firstrequest"

	PBFT_ADD_NODE_TIMEOUT = "timeout.addnode"
	PBFT_DEL_NODE_TIMEOUT = "timeout.delnode"
	PBFT_UPDATE_TIMEOUT = "timeout.updateTimeout"
)

//type for timer mgr
const (
	PBFT = iota  // pbft timer manager
	BATCH        // batch timer manager
)

// type for pbft status
const (
	ON = 1
	OFF = 0
)