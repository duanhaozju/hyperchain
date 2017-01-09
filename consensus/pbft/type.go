//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

// batchTimerEvent is sent when the batch timer expires
type batchTimerEvent struct{}

// viewChangeTimerEvent is sent when the view change timer expires
type viewChangeTimerEvent struct{}

// viewChangeResendTimerEvent is sent when the view change resend timer expires
type viewChangeResendTimerEvent struct{}

// viewChangedEvent is sent when the view change timer expires
type viewChangedEvent struct{}

// returnRequestBatchEvent is sent by pbft when we are forwarded a request
type returnRequestBatch *TransactionBatch

// returnRequestBatchEvent is sent by pbft when we are forwarded a request
type returnRequestBatchEvent *TransactionBatch

// stateUpdatedEvent  when stateUpdate is executed and return the result
type stateUpdatedEvent struct {
	seqNo uint64
}

// nullRequestEvent provides "keep-alive" null requests
type nullRequestEvent struct{}

type checkpointMessage struct {
	seqNo uint64
	id    []byte
}

type stateUpdateTarget struct {
	checkpointMessage
	replicas []uint64
}

type negoViewRspTimerEvent struct{}

type negoViewDoneEvent struct{}

type recoveryRestartTimerEvent struct{}

type recoveryDoneEvent struct{}

type firstRequestTimerEvent struct{}

type removeCache struct {
	vid uint64
}
