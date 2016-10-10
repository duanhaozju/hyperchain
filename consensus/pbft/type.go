package pbft

import "hyperchain/core/types"

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

type consensusMessageEvent ConsensusMessage

type checkpointMessage struct {
	seqNo uint64
	id    []byte
}

type stateUpdateTarget struct {
	checkpointMessage
	replicas []uint64
}