//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

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

//rbft config keys
const (
	RBFT_BATCH_SIZE      = "consensus.rbft.batchsize"
	RBFT_POOL_SIZE       = "consensus.rbft.poolsize"
	RBFT_VC_RESEND_LIMIT = "consensus.rbft.vcresendlimit"

	//timeout keys
	RBFT_NEGOVIEW_TIMEOUT          = "consensus.rbft.timeout.negoview"
	RBFT_RECOVERY_TIMEOUT          = "consensus.rbft.timeout.recovery"
	RBFT_VIEWCHANGE_TIMEOUT        = "consensus.rbft.timeout.viewchange"
	RBFT_RESEND_VIEWCHANGE_TIMEOUT = "consensus.rbft.timeout.resendviewchange"
	RBFT_CLEAN_VIEWCHANGE_TIMEOUT  = "consensus.rbft.timeout.cleanviewchange"
	RBFT_BATCH_TIMEOUT             = "consensus.rbft.timeout.batch"
	RBFT_REQUEST_TIMEOUT           = "consensus.rbft.timeout.request"
	RBFT_VALIDATE_TIMEOUT          = "consensus.rbft.timeout.validate"
	RBFT_NULLREQUEST_TIMEOUT       = "consensus.rbft.timeout.nullrequest"
	RBFT_FIRST_REQUEST_TIMEOUT     = "consensus.rbft.timeout.firstrequest"
	RBFT_UPDATE_TIMEOUT            = "consensus.rbft.timeout.update"
)
