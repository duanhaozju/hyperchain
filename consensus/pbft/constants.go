//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package pbft

// constant timer names
const (
	BATCH_TIMER             = "batchTimer"           // timer triggering package a batch
	VALIDATE_TIMER          = "validateTimer"        // timer expected to spend in validate
	VC_RESEND_TIMER         = "vcResendTimer"        // timer triggering resend of a view change
	NEW_VIEW_TIMER          = "newViewTimer"         // timer triggering view change by out of timeout of some requestBatch
	NULL_REQUEST_TIMER      = "nullRequestTimer"     // timer triggering send null request, used for heartbeat
	FIRST_REQUEST_TIMER     = "firstRequestTimer"    // timer set for replicas in case of primary start and shut down immediately
	NEGO_VIEW_RSP_TIMER     = "negoViewRspTimer"     // timer track timeout for N-f nego-view responses
	RECOVERY_RESTART_TIMER  = "recoveryRestartTimer" // timer track how long a recovery is finished and fires if needed
	CLEAN_VIEW_CHANGE_TIMER = "cleanViewChangeTimer" // timer track how long a viewchange msg will store in memory
)

//pbft config keys
const (
	PBFT_NODE_NUM        = "consensus.pbft.nodes"
	PBFT_BATCH_SIZE      = "consensus.pbft.batchsize"
	PBFT_POOL_SIZE       = "consensus.pbft.poolsize"
	PBFT_VC_RESEND_LIMIT = "consensus.pbft.vcresendlimit"

	//timeout keys
	PBFT_NEGOVIEW_TIMEOUT          = "consensus.pbft.timeout.negoview"
	PBFT_RECOVERY_TIMEOUT          = "consensus.pbft.timeout.recovery"
	PBFT_VIEWCHANGE_TIMEOUT        = "consensus.pbft.timeout.viewchange"
	PBFT_RESEND_VIEWCHANGE_TIMEOUT = "consensus.pbft.timeout.resendviewchange"
	PBFT_CLEAN_VIEWCHANGE_TIMEOUT  = "consensus.pbft.timeout.cleanviewchange"
	PBFT_BATCH_TIMEOUT             = "consensus.pbft.timeout.batch"
	PBFT_REQUEST_TIMEOUT           = "consensus.pbft.timeout.request"
	PBFT_VALIDATE_TIMEOUT          = "consensus.pbft.timeout.validate"
	PBFT_NULLREQUEST_TIMEOUT       = "consensus.pbft.timeout.nullrequest"
	PBFT_FIRST_REQUEST_TIMEOUT     = "consensus.pbft.timeout.firstrequest"
	PBFT_UPDATE_TIMEOUT            = "consensus.pbft.timeout.update"
)

// type for pbft status
const (
	ON  = 1
	OFF = 0
)
