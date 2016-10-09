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
type returnRequestBatchEvent *RequestBatch

// pbftMessageEvent is sent when a consensus messages is received to be sent to pbft
type pbftMessageEvent pbftMessage

// stateUpdatedEvent  when stateUpdate is executed and return the result
type stateUpdatedEvent struct {
	seqNo uint64
}

// nullRequestEvent provides "keep-alive" null requests
type nullRequestEvent struct{}

// This structure is used for incoming PBFT bound messages
type pbftMessage struct {
	sender uint64
	msg    *Message
}

type checkpointMessage struct {
	seqNo uint64
	id    []byte
}

type stateUpdateTarget struct {
	checkpointMessage
	replicas []uint64
}