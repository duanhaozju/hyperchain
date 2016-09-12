package pbft

import (
	"fmt"
	"time"
	"sort"

	"hyperchain/consensus/helper"
	"hyperchain/consensus/events"
	pb "hyperchain/protos"

	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"encoding/base64"
	"github.com/golang/protobuf/proto"
)

// =============================================================================
// init
// =============================================================================

var logger *logging.Logger // package-level logger

func init() {
	logger = logging.MustGetLogger("consensus/pbft")
}

// =============================================================================
// custom interfaces and structure definitions
// =============================================================================

// Event Types

// viewChangeTimerEvent is sent when the view change timer expires
type viewChangeTimerEvent struct{}

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

type pbftCore struct {
	//internal data
	helper	helper.Stack
	batch	*batch

	// PBFT data
	activeView	bool	// view change happening
	byzantine	bool	// whether this node is intentionally acting as Byzantine; useful for debugging on the testnet
	f		int	// max. number of faults we can tolerate
	N		int	// max.number of validators in the network
	h		uint64	// low watermark
	id		uint64	// replica ID; PBFT `i`
	K		uint64	// checkpoint period
	logMultiplier	uint64	// use this value to calculate log size : k*logMultiplier
	L		uint64	// log size
	lastExec	uint64	// last request we executed
	replicaCount	int	// number of replicas; PBFT `|R|`
	seqNo		uint64	// PBFT "n", strictly monotonic increasing sequence number
	view		uint64	// current view

	chkpts		map[uint64]string		// state checkpoints; map lastExec to global hash
	pset		map[uint64]*ViewChange_PQ	// state checkpoints; map lastExec to global hash
	qset		map[qidx]*ViewChange_PQ		// state checkpoints; map lastExec to global hash

	skipInProgress    bool			// Set when we have detected a fall behind scenario until we pick a new starting point
	stateTransferring bool			// Set when state transfer is executing
	highStateTarget   *stateUpdateTarget	// Set to the highest weak checkpoint cert we have observed
	hChkpts           map[uint64]uint64	// highest checkpoint sequence number observed for each replica

	currentExec		*uint64				// currently executing request
	timerActive		bool				// is the timer running?
	requestTimeout		time.Duration            	// progress timeout for requests
	outstandingReqBatches	map[string]*RequestBatch	// track whether we are waiting for request batches to execute
	newViewTimer		events.Timer			// track the timeout for each requestBatch
	newViewTimerReason	string				// what triggered the timer
	nullRequestTimer	events.Timer			// timeout triggering a null request
	nullRequestTimeout	time.Duration			// duration for this timeout

	// implementation of PBFT `in`
	reqBatchStore	map[string]*RequestBatch	// track request batches
	certStore	map[msgID]*msgCert		// track quorum certificates for requests
	checkpointStore	map[Checkpoint]bool		// track checkpoints as set
	committedCert	map[msgID]string

	valid		bool // whether we believe the state is up to date
}

type qidx struct {
	d string
	n uint64
}

type msgID struct { // our index through certStore
	v uint64
	n uint64
}

type msgCert struct {
	digest		string
	prePrepare	*PrePrepare
	sentPrepare	bool
	prepare		[]*Prepare
	sentCommit	bool
	commit		[]*Commit
	sentExecute	bool
}

type vcidx struct {
	v  uint64
	id uint64
}

type sortableUint64Slice []uint64

func (a sortableUint64Slice) Len() int {
	return len(a)
}
func (a sortableUint64Slice) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a sortableUint64Slice) Less(i, j int) bool {
	return a[i] < a[j]
}

// =============================================================================
// constructors
// =============================================================================

func newPbftCore(id uint64, config *viper.Viper, batch *batch, etf events.TimerFactory) *pbftCore {

	var err error
	instance := &pbftCore{}
	instance.id = id
	instance.helper = batch.getHelper()
	instance.nullRequestTimer = etf.CreateTimer()
	instance.newViewTimer = etf.CreateTimer()

	instance.batch = batch

	instance.N = config.GetInt("general.N")
	instance.f = config.GetInt("general.f")

	if instance.f*3+1 > instance.N {
		panic(fmt.Sprintf("need at least %d enough replicas to tolerate %d byzantine faults, but only %d replicas configured", instance.f*3+1, instance.f, instance.N))
	}

	instance.K = uint64(config.GetInt("general.K"))

	instance.logMultiplier = uint64(config.GetInt("general.logmultiplier"))
	if instance.logMultiplier < 2 {
		panic("Log multiplier must be greater than or equal to 2")
	}

	instance.L = instance.logMultiplier * instance.K // log size
	instance.byzantine = config.GetBool("general.byzantine")
	instance.requestTimeout, err = time.ParseDuration(config.GetString("timeout.request"))

	if err != nil {
		panic(fmt.Errorf("Cannot parse request timeout: %s", err))
	}

	instance.nullRequestTimeout, err = time.ParseDuration(config.GetString("timeout.nullrequest"))

	if err != nil {
		instance.nullRequestTimeout = 0
	}

	instance.activeView = true
	instance.replicaCount = instance.N

	logger.Infof("PBFT Max number of validating peers (N) = %v", instance.N)
	logger.Infof("PBFT Max number of failing peers (f) = %v", instance.f)
	logger.Infof("PBFT byzantine flag = %v", instance.byzantine)
	logger.Infof("PBFT request timeout = %v", instance.requestTimeout)
	logger.Infof("PBFT Checkpoint period (K) = %v", instance.K)
	logger.Infof("PBFT Log multiplier = %v", instance.logMultiplier)
	logger.Infof("PBFT log size (L) = %v", instance.L)

	if instance.nullRequestTimeout > 0 {
		logger.Infof("PBFT null requests timeout = %v", instance.nullRequestTimeout)
	} else {
		logger.Infof("PBFT null requests disabled")
	}

	// init the logs
	instance.certStore = make(map[msgID]*msgCert)
	instance.reqBatchStore = make(map[string]*RequestBatch)
	instance.checkpointStore = make(map[Checkpoint]bool)
	instance.chkpts = make(map[uint64]string)
	instance.pset = make(map[uint64]*ViewChange_PQ)
	instance.qset = make(map[qidx]*ViewChange_PQ)
	instance.committedCert = make(map[msgID]string)

	// initialize state transfer
	instance.hChkpts = make(map[uint64]uint64)

	instance.chkpts[0] = "XXX GENESIS"

	instance.outstandingReqBatches = make(map[string]*RequestBatch)

	instance.restoreState()

	logger.Infof("--------PBFT finish start, nodeID: %d--------", instance.id)

	return instance
}

// close tears down resources opened by newPbftCore
func (instance *pbftCore) close() {
	instance.nullRequestTimer.Halt()
}

// allow the view-change protocol to kick-off when the timer expires
func (instance *pbftCore) ProcessEvent(e events.Event) events.Event {

	var err error
	logger.Debugf("Replica %d processing event", instance.id)

	switch et := e.(type) {
	//case *pbftMessage:
	//	return pbftMessageEvent(*et)
	case pbftMessageEvent:
		msg := et
		logger.Debugf("Replica %d received incoming message from %v", instance.id, msg.sender)
		next, err := instance.recvMsg(msg.msg, msg.sender)
		if err != nil {
			break
		}
		return next
	case *RequestBatch:
		err = instance.recvRequestBatch(et)
	case *PrePrepare:
		err = instance.recvPrePrepare(et)
	case *Prepare:
		err = instance.recvPrepare(et)
	case *Commit:
		err = instance.recvCommit(et)
	case *Checkpoint:
		return instance.recvCheckpoint(et)
	case *stateUpdatedEvent:
		instance.batch.reqStore = newRequestStore()
		err = instance.recvStateUpdatedEvent(et)
	case nullRequestEvent:
		instance.nullRequestHandler()
	default:
		logger.Warningf("Replica %d received an unknown message type %T", instance.id, et)
	}

	if err != nil {
		logger.Warning(err.Error())
	}

	return nil
}

// =============================================================================
// helper functions for PBFT
// =============================================================================

// Given a certain view n, what is the expected primary?
func (instance *pbftCore) primary(n uint64) uint64 {
	return (n % uint64(instance.replicaCount) + 1)
}

// Is the sequence number between watermarks?
func (instance *pbftCore) inW(n uint64) bool {
	return n > instance.h && n-instance.h <= instance.L
}

// Is the view right? And is the sequence number between watermarks?
func (instance *pbftCore) inWV(v uint64, n uint64) bool {
	return instance.view == v && instance.inW(n)
}

// Given a digest/view/seq, is there an entry in the certLog?
// If so, return it. If not, create it.
func (instance *pbftCore) getCert(v uint64, n uint64) (cert *msgCert) {

	idx := msgID{v, n}
	cert, ok := instance.certStore[idx]

	if ok {
		return
	}

	cert = &msgCert{}
	instance.certStore[idx] = cert

	return
}


// =============================================================================
// prepare/commit quorum checks helper
// =============================================================================

func (instance *pbftCore) preparedReplicasQuorum() int {
	return (2 * instance.f)
}

func (instance *pbftCore) committedReplicasQuorum() int {
	return (2 * instance.f + 1)
}

// intersectionQuorum returns the number of replicas that have to
// agree to guarantee that at least one correct replica is shared by
// two intersection quora
func (instance *pbftCore) intersectionQuorum() int {
	return (instance.N + instance.f + 2) / 2
}

// =============================================================================
// pre-prepare/prepare/commit check helper
// =============================================================================

func (instance *pbftCore) prePrepared(digest string, v uint64, n uint64) bool {

	_, mInLog := instance.reqBatchStore[digest]

	if digest != "" && !mInLog {
		logger.Debugf("Replica %d havan't store the reqBatch")
		return false
	}

	//if q, ok := instance.qset[qidx{digest, n}]; ok && q.View == v {
	//	return true
	//}

	cert := instance.certStore[msgID{v, n}]

	if cert != nil {
		p := cert.prePrepare
		if p != nil && p.View == v && p.SequenceNumber == n && p.BatchDigest == digest {
			return true
		}
	}

	logger.Debugf("Replica %d does not have view=%d/seqNo=%d pre-prepared",
		instance.id, v, n)

	return false
}

func (instance *pbftCore) prepared(digest string, v uint64, n uint64) bool {

	if !instance.prePrepared(digest, v, n) {
		return false
	}

	//if p, ok := instance.pset[n]; ok && p.View == v && p.BatchDigest == digest {
	//	return true
	//}

	quorum := 0
	cert := instance.certStore[msgID{v, n}]

	if cert == nil {
		return false
	}

	for _, p := range cert.prepare {
		if p.View == v && p.SequenceNumber == n && p.BatchDigest == digest {
			quorum++
		}
	}

	logger.Debugf("Replica %d prepare count for view=%d/seqNo=%d: %d",
		instance.id, v, n, quorum)

	return quorum >= instance.preparedReplicasQuorum()
}

func (instance *pbftCore) committed(digest string, v uint64, n uint64) bool {

	if !instance.prepared(digest, v, n) {
		return false
	}

	quorum := 0
	cert := instance.certStore[msgID{v, n}]

	if cert == nil {
		return false
	}

	for _, p := range cert.commit {
		if p.View == v && p.SequenceNumber == n {
			quorum++
		}
	}

	logger.Debugf("Replica %d commit count for view=%d/seqNo=%d: %d",
		instance.id, v, n, quorum)

	return quorum >= instance.committedReplicasQuorum()
}

// =============================================================================
// receive methods
// =============================================================================

func (instance *pbftCore) nullRequestHandler() {

	if !instance.activeView {
		return
	}
	// time for the primary to send a null request
	// pre-prepare with null digest
	logger.Info("Primary %d null request timer expired, sending null request", instance.id)
	instance.sendPrePrepare(nil, "")

}

func (instance *pbftCore) recvMsg(msg *Message, senderID uint64) (interface{}, error) {

	if reqBatch := msg.GetRequestBatch(); reqBatch != nil {
		return reqBatch, nil
	} else if preprep := msg.GetPrePrepare(); preprep != nil {
		if senderID != preprep.ReplicaId {
			return nil, fmt.Errorf("Sender ID included in pre-prepare message (%v) doesn't match ID corresponding to the receiving stream (%v)", preprep.ReplicaId, senderID)
		}
		return preprep, nil
	} else if prep := msg.GetPrepare(); prep != nil {
		if senderID != prep.ReplicaId {
			return nil, fmt.Errorf("Sender ID included in prepare message (%v) doesn't match ID corresponding to the receiving stream (%v)", prep.ReplicaId, senderID)
		}
		return prep, nil
	} else if commit := msg.GetCommit(); commit != nil {
		if senderID != commit.ReplicaId {
			return nil, fmt.Errorf("Sender ID included in commit message (%v) doesn't match ID corresponding to the receiving stream (%v)", commit.ReplicaId, senderID)
		}
		return commit, nil
	} else if chkpt := msg.GetCheckpoint(); chkpt != nil {
		if senderID != chkpt.ReplicaId {
			return nil, fmt.Errorf("Sender ID included in checkpoint message (%v) doesn't match ID corresponding to the receiving stream (%v)", chkpt.ReplicaId, senderID)
		}
		return chkpt, nil
	}

	return nil, fmt.Errorf("Invalid message: %v", msg)
}

func (instance *pbftCore) recvStateUpdatedEvent(et *stateUpdatedEvent) error {


	instance.stateTransferring = false
	// If state transfer did not complete successfully, or if it did not reach our low watermark, do it again
	if et.seqNo < instance.h {
		logger.Warningf("Replica %d recovered to seqNo %d but our low watermark has moved to %d", instance.id, et.seqNo, instance.h)
		if instance.highStateTarget == nil {
			logger.Debugf("Replica %d has no state targets, cannot resume state transfer yet", instance.id)
		} else if et.seqNo < instance.highStateTarget.seqNo {
			logger.Debugf("Replica %d has state target for %d, transferring", instance.id, instance.highStateTarget.seqNo)
			instance.retryStateTransfer(nil)
		} else {
			logger.Debugf("Replica %d has no state target above %d, highest is %d", instance.id, et.seqNo, instance.highStateTarget.seqNo)
		}
		return nil
	}

	logger.Infof("Replica %d application caught up via state transfer, lastExec now %d", instance.id, et.seqNo)
	// XXX create checkpoint
	instance.lastExec = et.seqNo
	instance.moveWatermarks(instance.lastExec) // The watermark movement handles moving this to a checkpoint boundary
	instance.skipInProgress = false
	instance.validateState()
	instance.executeOutstanding()

	return nil

}

func (instance *pbftCore) recvRequestBatch(reqBatch *RequestBatch) error {

	digest := hash(reqBatch)
	logger.Debugf("Replica %d received request batch %s", instance.id, digest)

	instance.reqBatchStore[digest] = reqBatch
	instance.outstandingReqBatches[digest] = reqBatch
	instance.persistRequestBatch(digest)
	if instance.activeView {
		instance.softStartTimer(instance.requestTimeout, fmt.Sprintf("new request batch %s", digest))
	}
	if instance.primary(instance.view) == instance.id && instance.activeView {
		instance.nullRequestTimer.Stop()
		instance.sendPrePrepare(reqBatch, digest)
	} else {
		logger.Debugf("Replica %d is backup, not sending pre-prepare for request batch %s", instance.id, digest)
	}

	return nil
}

func (instance *pbftCore) sendPrePrepare(reqBatch *RequestBatch, digest string) {

	logger.Debugf("Replica %d is primary, issuing pre-prepare for request batch %s", instance.id, digest)

	n := instance.seqNo + 1
	for _, cert := range instance.certStore { // check for other PRE-PREPARE for same digest, but different seqNo
		if p := cert.prePrepare; p != nil {
			if p.View == instance.view && p.SequenceNumber != n && p.BatchDigest == digest && digest != "" {
				logger.Infof("Other pre-prepare found with same digest but different seqNo: %d instead of %d", p.SequenceNumber, n)
				return
			}
		}
	}

	if !instance.inWV(instance.view, n) {
		logger.Debugf("Replica %d is primary, not sending pre-prepare for request batch %s because it is out of sequence numbers", instance.id, digest)
		return
	}

	logger.Debugf("Primary %d broadcasting pre-prepare for view=%d/seqNo=%d", instance.id, instance.view, n)
	instance.seqNo = n
	preprep := &PrePrepare{
		View:           instance.view,
		SequenceNumber: n,
		BatchDigest:    digest,
		RequestBatch:   reqBatch,
		ReplicaId:      instance.id,
	}
	cert := instance.getCert(instance.view, n)
	cert.prePrepare = preprep
	cert.digest = digest
	instance.persistQSet()
	msg := pbftMsgHelper(&Message{Payload: &Message_PrePrepare{PrePrepare: preprep}}, instance.id)
	instance.helper.InnerBroadcast(msg)

	instance.maybeSendCommit(digest, instance.view, n)
}

func (instance *pbftCore) recvPrePrepare(preprep *PrePrepare) error {

	logger.Debugf("Replica %d received pre-prepare from replica %d for view=%d/seqNo=%d, digest: ",
		instance.id, preprep.ReplicaId, preprep.View, preprep.SequenceNumber, preprep.BatchDigest)

	if !instance.activeView {
		logger.Debugf("Replica %d ignoring pre-prepare as we sre in view change", instance.id)
	}

	if instance.primary(instance.view) != preprep.ReplicaId {
		logger.Warningf("Pre-prepare from other than primary: got %d, should be %d", preprep.ReplicaId, instance.primary(instance.view))
		return nil
	}

	if !instance.inWV(preprep.View, preprep.SequenceNumber) {
		if preprep.SequenceNumber != instance.h && !instance.skipInProgress {
			logger.Warningf("Replica %d pre-prepare view different, or sequence number outside watermarks: preprep.View %d, expected.View %d, seqNo %d, low-mark %d", instance.id, preprep.View, instance.primary(instance.view), preprep.SequenceNumber, instance.h)
		} else {
			// This is perfectly normal
			logger.Debugf("Replica %d pre-prepare view different, or sequence number outside watermarks: preprep.View %d, expected.View %d, seqNo %d, low-mark %d", instance.id, preprep.View, instance.primary(instance.view), preprep.SequenceNumber, instance.h)
		}

		return nil
	}

	cert := instance.getCert(preprep.View, preprep.SequenceNumber)

	if cert.digest != "" && cert.digest != preprep.BatchDigest {
		logger.Warningf("Pre-prepare found for same view/seqNo but different digest: received %s, stored %s", preprep.BatchDigest, cert.digest)
		return nil
	}

	cert.prePrepare = preprep
	cert.digest = preprep.BatchDigest

	// Store the request batch if, for whatever reason, we haven't received it from an earlier broadcast
	if _, ok := instance.reqBatchStore[preprep.BatchDigest]; !ok && preprep.BatchDigest != "" {
		digest := hash(preprep.GetRequestBatch())
		if digest != preprep.BatchDigest {
			logger.Warningf("Pre-prepare and request digest do not match: request %s, digest %s", digest, preprep.BatchDigest)
			return nil
		}
		instance.reqBatchStore[digest] = preprep.GetRequestBatch()
		logger.Debugf("Replica %d storing request batch %s in outstanding request batch store", instance.id, digest)
		instance.outstandingReqBatches[digest] = preprep.GetRequestBatch()
		instance.persistRequestBatch(digest)
	}

	instance.softStartTimer(instance.requestTimeout, fmt.Sprintf("new pre-prepare for request batch %s", preprep.BatchDigest))
	instance.nullRequestTimer.Stop()

	if instance.primary(instance.view) != instance.id && instance.prePrepared(preprep.BatchDigest, preprep.View, preprep.SequenceNumber) && !cert.sentPrepare {
		logger.Debugf("Backup %d broadcasting prepare for view=%d/seqNo=%d", instance.id, preprep.View, preprep.SequenceNumber)
		prep := &Prepare{
			View:           preprep.View,
			SequenceNumber: preprep.SequenceNumber,
			BatchDigest:    preprep.BatchDigest,
			ReplicaId:      instance.id,
		}
		cert.sentPrepare = true
		instance.persistQSet()
		instance.recvPrepare(prep)
		msg := pbftMsgHelper(&Message{Payload: &Message_Prepare{Prepare: prep}}, instance.id)
		return instance.helper.InnerBroadcast(msg)
	}

	return nil
}

func (instance *pbftCore) recvPrepare(prep *Prepare) error {

	logger.Debugf("Replica %d received prepare from replica %d for view=%d/seqNo=%d",
		instance.id, prep.ReplicaId, prep.View, prep.SequenceNumber)

	if instance.primary(prep.View) == prep.ReplicaId {
		logger.Warningf("Replica %d received prepare from primary, ignoring", instance.id)
		return nil
	}

	if !instance.inWV(prep.View, prep.SequenceNumber) {
		if prep.SequenceNumber != instance.h && !instance.skipInProgress {
			logger.Warningf("Replica %d ignoring prepare for view=%d/seqNo=%d: not in-wv, in view %d, low water mark %d", instance.id, prep.View, prep.SequenceNumber, instance.view, instance.h)
		} else {
			// This is perfectly normal
			logger.Debugf("Replica %d ignoring prepare for view=%d/seqNo=%d: not in-wv, in view %d, low water mark %d", instance.id, prep.View, prep.SequenceNumber, instance.view, instance.h)
		}
		return nil
	}

	cert := instance.getCert(prep.View, prep.SequenceNumber)

	for _, prevPrep := range cert.prepare {
		if prevPrep.ReplicaId == prep.ReplicaId {
			logger.Warningf("Ignoring duplicate prepare from %d, --------view=%d/seqNo=%d--------", prep.ReplicaId, prep.View, prep.SequenceNumber)
			return nil
		}
	}

	cert.prepare = append(cert.prepare, prep)
	instance.persistPSet()

	return instance.maybeSendCommit(prep.BatchDigest, prep.View, prep.SequenceNumber)
}

//
func (instance *pbftCore) maybeSendCommit(digest string, v uint64, n uint64) error {

	cert := instance.getCert(v, n)

	if instance.prepared(digest, v, n) && !cert.sentCommit {
		logger.Debugf("Replica %d broadcasting commit for view=%d/seqNo=%d",
			instance.id, v, n)
		commit := &Commit{
			View:           v,
			SequenceNumber: n,
			BatchDigest:    digest,
			ReplicaId:      instance.id,
		}
		cert.sentCommit = true
		instance.recvCommit(commit)
		msg := pbftMsgHelper(&Message{Payload: &Message_Commit{Commit: commit}}, instance.id)
		return instance.helper.InnerBroadcast(msg)
	}

	return nil
}

func (instance *pbftCore) recvCommit(commit *Commit) error {

	logger.Debugf("Replica %d received commit from replica %d for view=%d/seqNo=%d",
		instance.id, commit.ReplicaId, commit.View, commit.SequenceNumber)

	if !instance.inWV(commit.View, commit.SequenceNumber) {
		if commit.SequenceNumber != instance.h && !instance.skipInProgress {
			logger.Warningf("Replica %d ignoring commit for view=%d/seqNo=%d: not in-wv, in view %d, high water mark %d", instance.id, commit.View, commit.SequenceNumber, instance.view, instance.h)
		} else {
			// This is perfectly normal
			logger.Debugf("Replica %d ignoring commit for view=%d/seqNo=%d: not in-wv, in view %d, high water mark %d", instance.id, commit.View, commit.SequenceNumber, instance.view, instance.h)
		}
		return nil
	}

	cert := instance.getCert(commit.View, commit.SequenceNumber)
	for _, prevCommit := range cert.commit {
		if prevCommit.ReplicaId == commit.ReplicaId {
			logger.Warningf("Ignoring duplicate commit from %d, --------view=%d/seqNo=%d--------", commit.ReplicaId, commit.View, commit.SequenceNumber)
			return nil
		}
	}

	cert.commit = append(cert.commit, commit)

	if instance.committed(commit.BatchDigest, commit.View, commit.SequenceNumber) && cert.sentExecute == false {
		instance.stopTimer(commit.SequenceNumber)
		delete(instance.outstandingReqBatches, commit.BatchDigest)
		idx := msgID{v: commit.View, n: commit.SequenceNumber}
		instance.committedCert[idx] = cert.digest
		instance.executeOutstanding()
	}

	return nil
}

func (instance *pbftCore) executeOutstanding() {

	if instance.currentExec != nil {
		logger.Debugf("Replica %d not attempting to executeOutstanding bacause it is currently executing %d", instance.id, instance.currentExec)
	}

	logger.Debugf("Replica %d attempting to executeOutstanding", instance.id)

	for idx := range instance.committedCert {
		if instance.executeOne(idx) {
			break
		}
	}

	//instance.startTimerIfOutstandingRequests()

}

func (instance *pbftCore) executeOne(idx msgID) bool {

	cert := instance.certStore[idx]

	if cert == nil || cert.prePrepare == nil {
		logger.Debugf("Replica %d already checkpoint for view=%d/seqNo=%d", instance.id, idx.v, idx.n)
		return false
	}

	// check if already executed
	if cert.sentExecute == true {
		logger.Debugf("Replica %d already execute for view=%d/seqNo=%d", instance.id, idx.v, idx.n)
		return false
	}

	if idx.n != instance.lastExec+1 {
		logger.Debugf("Replica %d hasn't done with last execute %d, seq=%d", instance.id, instance.lastExec, idx.n)
		return false
	}

	// skipInProgress == true, then this replica is in viewchange, not reply or execute
	if instance.skipInProgress {
		logger.Warningf("Replica %d currently picking a starting point to resume, will not execute", instance.id)
		return false
	}

	digest := cert.digest
	reqBatch := instance.reqBatchStore[digest]

	// check if committed
	if !instance.committed(digest, idx.v, idx.n) {
		return false
	}

	currentExec := idx.n
	instance.currentExec = &currentExec

	if digest == "" {
		logger.Infof("Replica %d executing null request for view=%d/seqNo=%d", instance.id, idx.v, idx.n)
		cert.sentExecute = true
		instance.execDoneSync(idx)
	} else {
		logger.Infof("--------call execute--------view=%d/seqNo=%d--------", idx.v, idx.n)
		exeBatch := exeBatchHelper(reqBatch, idx.n)
		instance.helper.Execute(exeBatch)
		cert.sentExecute = true
		instance.execDoneSync(idx)
	}

	return true
}

func (instance *pbftCore) execDoneSync(idx msgID) {

	if instance.currentExec != nil {
		logger.Debugf("Replica %d finish execution %d, trying next", instance.id, *instance.currentExec)
		instance.lastExec = *instance.currentExec
		delete(instance.committedCert, idx)
		if instance.lastExec % instance.K == 0 {
			bcInfo := getBlockchainInfo()
			height := bcInfo.Height
			if height == instance.lastExec {
				logger.Debugf("Call the checkpoint, seqNo=%d, block height=%d", instance.lastExec, height)
				//time.Sleep(3*time.Millisecond)
				instance.checkpoint(instance.lastExec, bcInfo)
			} else  {
				// reqBatch call execute but have not done with execute
				logger.Errorf("Fail to call the checkpoint, seqNo=%d, block height=%d", instance.lastExec, height)
				//instance.retryCheckpoint(instance.lastExec)
			}
		}
	} else {
		logger.Warningf("Replica %d had execDoneSync called, flagging ourselves as out of data", instance.id)
		instance.skipInProgress = true
	}

	instance.currentExec = nil
	instance.executeOutstanding()

}

func (instance *pbftCore) checkpoint(n uint64, info *pb.BlockchainInfo) {

	if n % instance.K != 0 {
		logger.Errorf("Attempted to checkpoint a sequence number (%d) which is not a multiple of the checkpoint interval (%d)", n, instance.K)
		return
	}

	id, _ := proto.Marshal(info)
	idAsString := base64.StdEncoding.EncodeToString(id)
	seqNo := n

	logger.Debugf("Replica %d preparing checkpoint for view=%d/seqNo=%d and b64 id of %s",
		instance.id, instance.view, seqNo, idAsString)

	chkpt := &Checkpoint{
		SequenceNumber: seqNo,
		ReplicaId:      instance.id,
		Id:             idAsString,
	}
	instance.chkpts[seqNo] = idAsString

	instance.persistCheckpoint(seqNo, id)
	instance.recvCheckpoint(chkpt)
	msg := pbftMsgHelper(&Message{Payload: &Message_Checkpoint{Checkpoint: chkpt}}, instance.id)
	instance.helper.InnerBroadcast(msg)
}

func (instance *pbftCore) retryCheckpoint(n uint64) {

	if n % instance.K != 0 {
		return
	}

	bcInfo := getBlockchainInfo()
	height := bcInfo.Height
	if height == n {
		logger.Debugf("Call the checkpoint, seqNo=%d, block height=%d", n, height)
		instance.checkpoint(n, bcInfo)
	} else {
		// reqBatch call execute but have not done with execute
		logger.Warningf("Fail to call the checkpoint, seqNo=%d, block height=%d", n, height)
		instance.retryCheckpoint(instance.lastExec)
	}


}

func (instance *pbftCore) recvCheckpoint(chkpt *Checkpoint) events.Event {

	logger.Infof("Replica %d received checkpoint from replica %d, seqNo %d, digest %s",
		instance.id, chkpt.ReplicaId, chkpt.SequenceNumber, chkpt.Id)

	if instance.weakCheckpointSetOutOfRange(chkpt) {
		return nil
	}

	if !instance.inW(chkpt.SequenceNumber) {
		if chkpt.SequenceNumber != instance.h && !instance.skipInProgress {
			// It is perfectly normal that we receive checkpoints for the watermark we just raised, as we raise it after 2f+1, leaving f replies left
			logger.Warningf("Checkpoint sequence number outside watermarks: seqNo %d, low-mark %d", chkpt.SequenceNumber, instance.h)
		} else {
			logger.Debugf("Checkpoint sequence number outside watermarks: seqNo %d, low-mark %d", chkpt.SequenceNumber, instance.h)
		}
		return nil
	}

	instance.checkpointStore[*chkpt] = true

	matching := 0
	for testChkpt := range instance.checkpointStore {
		if testChkpt.SequenceNumber == chkpt.SequenceNumber && testChkpt.Id == chkpt.Id {
			matching++
		}
	}
	logger.Debugf("Replica %d found %d matching checkpoints for seqNo %d, digest %s",
		instance.id, matching, chkpt.SequenceNumber, chkpt.Id)

	if matching == instance.f+1 {
		// We do have a weak cert
		instance.witnessCheckpointWeakCert(chkpt)
	}

	if matching < instance.intersectionQuorum() {
		// We do not have a quorum yet
		return nil
	}

	// It is actually just fine if we do not have this checkpoint
	// and should not trigger a state transfer
	// Imagine we are executing sequence number k-1 and we are slow for some reason
	// then everyone else finishes executing k, and we receive a checkpoint quorum
	// which we will agree with very shortly, but do not move our watermarks until
	// we have reached this checkpoint
	// Note, this is not divergent from the paper, as the paper requires that
	// the quorum certificate must contain 2f+1 messages, including its own
	chkptID, ok := instance.chkpts[chkpt.SequenceNumber]
	if !ok {
		logger.Debugf("Replica %d found checkpoint quorum for seqNo %d, digest %s, but it has not reached this checkpoint itself yet",
			instance.id, chkpt.SequenceNumber, chkpt.Id)
		if instance.skipInProgress {
			logSafetyBound := instance.h + instance.L/2
			// As an optimization, if we are more than half way out of our log and in state transfer, move our watermarks so we don't lose track of the network
			// if needed, state transfer will restart on completion to a more recent point in time
			if chkpt.SequenceNumber >= logSafetyBound {
				logger.Debugf("Replica %d is in state transfer, but, the network seems to be moving on past %d, moving our watermarks to stay with it", instance.id, logSafetyBound)
				instance.moveWatermarks(chkpt.SequenceNumber)
			}
		}
		return nil
	}

	logger.Infof("Replica %d found checkpoint quorum for seqNo %d, digest %s",
		instance.id, chkpt.SequenceNumber, chkpt.Id)

	if chkptID != chkpt.Id {
		logger.Criticalf("Replica %d generated a checkpoint of %s, but a quorum of the network agrees on %s. This is almost definitely non-deterministic chaincode.",
			instance.id, chkptID, chkpt.Id)
		instance.stateTransfer(nil)
	}

	instance.moveWatermarks(chkpt.SequenceNumber)

	return nil
}

func (instance *pbftCore) weakCheckpointSetOutOfRange(chkpt *Checkpoint) bool {
	H := instance.h + instance.L

	// Track the last observed checkpoint sequence number if it exceeds our high watermark, keyed by replica to prevent unbounded growth
	if chkpt.SequenceNumber < H {
		// For non-byzantine nodes, the checkpoint sequence number increases monotonically
		delete(instance.hChkpts, chkpt.ReplicaId)
	} else {
		// We do not track the highest one, as a byzantine node could pick an arbitrarilly high sequence number
		// and even if it recovered to be non-byzantine, we would still believe it to be far ahead
		instance.hChkpts[chkpt.ReplicaId] = chkpt.SequenceNumber

		// If f+1 other replicas have reported checkpoints that were (at one time) outside our watermarks
		// we need to check to see if we have fallen behind.
		if len(instance.hChkpts) >= instance.f+1 {
			chkptSeqNumArray := make([]uint64, len(instance.hChkpts))
			index := 0
			for replicaID, hChkpt := range instance.hChkpts {
				chkptSeqNumArray[index] = hChkpt
				index++
				if hChkpt < H {
					delete(instance.hChkpts, replicaID)
				}
			}
			sort.Sort(sortableUint64Slice(chkptSeqNumArray))

			// If f+1 nodes have issued checkpoints above our high water mark, then
			// we will never record 2f+1 checkpoints for that sequence number, we are out of date
			// (This is because all_replicas - missed - me = 3f+1 - f - 1 = 2f)
			if m := chkptSeqNumArray[len(chkptSeqNumArray)-(instance.f+1)]; m > H {
				logger.Warningf("Replica %d is out of date, f+1 nodes agree checkpoint with seqNo %d exists but our high water mark is %d", instance.id, chkpt.SequenceNumber, H)
				instance.reqBatchStore = make(map[string]*RequestBatch) // Discard all our requests, as we will never know which were executed, to be addressed in #394
				//instance.persistDelAllRequestBatches()
				instance.moveWatermarks(m)
				instance.outstandingReqBatches = make(map[string]*RequestBatch)
				instance.skipInProgress = true
				instance.invalidateState()
				instance.stopTimer(chkpt.SequenceNumber)

				// TODO, reprocess the already gathered checkpoints, this will make recovery faster, though it is presently correct

				return true
			}
		}
	}

	return false
}

func (instance *pbftCore) witnessCheckpointWeakCert(chkpt *Checkpoint) {

	// Only ever invoked for the first weak cert, so guaranteed to be f+1
	checkpointMembers := make([]uint64, instance.f+1)
	i := 0
	for testChkpt := range instance.checkpointStore {
		if testChkpt.SequenceNumber == chkpt.SequenceNumber && testChkpt.Id == chkpt.Id {
			checkpointMembers[i] = testChkpt.ReplicaId
			logger.Debugf("Replica %d adding replica %d (handle %v) to weak cert", instance.id, testChkpt.ReplicaId, checkpointMembers[i])
			i++
		}
	}

	snapshotID, err := base64.StdEncoding.DecodeString(chkpt.Id)
	if err != nil {
		err = fmt.Errorf("Replica %d received a weak checkpoint cert which could not be decoded (%s)", instance.id, chkpt.Id)
		logger.Error(err.Error())
		return
	}

	target := &stateUpdateTarget{
		checkpointMessage:	checkpointMessage{
			seqNo:	chkpt.SequenceNumber,
			id:	snapshotID,
		},
		replicas: 		checkpointMembers,
	}
	instance.updateHighStateTarget(target)

	if instance.skipInProgress {
		logger.Infof("Replica %d is catching up and witnessed a weak certificate for checkpoint %d, weak cert attested to by %d of %d (%v)",
			instance.id, chkpt.SequenceNumber, i, instance.replicaCount, checkpointMembers)
		// The view should not be set to active, this should be handled by the yet unimplemented SUSPECT, see https://github.com/hyperledger/fabric/issues/1120
		instance.retryStateTransfer(target)
	}
}

func (instance *pbftCore) moveWatermarks(n uint64) {

	// round down n to previous low watermark
	h := n / instance.K * instance.K

	for idx, cert := range instance.certStore {
		if idx.n <= h {
			logger.Debugf("Replica %d cleaning quorum certificate for view=%d/seqNo=%d",
				instance.id, idx.v, idx.n)
			instance.persistDelRequestBatch(cert.digest)
			delete(instance.reqBatchStore, cert.digest)
			delete(instance.certStore, idx)
		}
	}

	for testChkpt := range instance.checkpointStore {
		if testChkpt.SequenceNumber <= h {
			logger.Debugf("Replica %d cleaning checkpoint message from replica %d, seqNo %d, b64 snapshot id %s",
				instance.id, testChkpt.ReplicaId, testChkpt.SequenceNumber, testChkpt.Id)
			delete(instance.checkpointStore, testChkpt)
		}
	}

	for n := range instance.pset {
		if n <= h {
			delete(instance.pset, n)
			//instance.persistDelPSet(n)
		}
	}

	for idx := range instance.qset {
		if idx.n <= h {
			delete(instance.qset, idx)
			//instance.persistDelQSet(idx)
		}
	}

	for n := range instance.chkpts {
		if n < h {
			delete(instance.chkpts, n)
			instance.persistDelCheckpoint(n)
		}
	}

	instance.h = h

	logger.Infof("Replica %d updated low watermark to %d",
		instance.id, instance.h)

	instance.resubmitRequestBatches()
}

func (instance *pbftCore) updateHighStateTarget(target *stateUpdateTarget) {
	if instance.highStateTarget != nil && instance.highStateTarget.seqNo >= target.seqNo {
		logger.Infof("Replica %d not updating state target to seqNo %d, has target for seqNo %d", instance.id, target.seqNo, instance.highStateTarget.seqNo)
		return
	}

	instance.highStateTarget = target
}

func (instance *pbftCore) stateTransfer(optional *stateUpdateTarget) {

	if !instance.skipInProgress {
		logger.Debugf("Replica %d is out of sync, pending state transfer", instance.id)
		instance.skipInProgress = true
		instance.invalidateState()
	}

	instance.retryStateTransfer(optional)
}

func (instance *pbftCore) retryStateTransfer(optional *stateUpdateTarget) {

	if instance.stateTransferring {
		logger.Debugf("Replica %d is currently mid state transfer, it must wait for this state transfer to complete before initiating a new one", instance.id)
		return
	}

	target := optional
	if target == nil {
		if instance.highStateTarget == nil {
			logger.Debugf("Replica %d has no targets to attempt state transfer to, delaying", instance.id)
			return
		}
		target = instance.highStateTarget
	}

	instance.stateTransferring = true

	logger.Debugf("Replica %d is initiating state transfer to seqNo %d", instance.id, target.seqNo)

	//instance.batch.pbftManager.Queue() <- stateUpdateEvent // Todo for stateupdate
	//instance.consumer.skipTo(target.seqNo, target.id, target.replicas)

	instance.skipTo(target.seqNo, target.id, target.replicas)


}

func (instance *pbftCore) resubmitRequestBatches() {
	if instance.primary(instance.view) != instance.id {
		return
	}

	var submissionOrder []*RequestBatch

	outer:
	for d, reqBatch := range instance.outstandingReqBatches {
		for _, cert := range instance.certStore {
			if cert.digest == d {
				logger.Debugf("Replica %d already has certificate for request batch %s - not going to resubmit", instance.id, d)
				continue outer
			}
		}
		logger.Infof("Replica %d has detected request batch %s must be resubmitted", instance.id, d)
		submissionOrder = append(submissionOrder, reqBatch)
	}

	if len(submissionOrder) == 0 {
		return
	}

	for _, reqBatch := range submissionOrder {
		// This is a request batch that has not been pre-prepared yet
		// Trigger request batch processing again
		instance.recvRequestBatch(reqBatch)
	}
}

func (instance *pbftCore) startTimerIfOutstandingRequests() {
	if instance.skipInProgress || instance.currentExec != nil {
		// Do not start the view change timer if we are executing or state transferring, these take arbitrarilly long amounts of time
		return
	}

	if len(instance.outstandingReqBatches) > 0 {
		getOutstandingDigests := func() []string {
			var digests []string
			for digest := range instance.outstandingReqBatches {
				digests = append(digests, digest)
			}
			return digests
		}()
		instance.softStartTimer(instance.requestTimeout, fmt.Sprintf("outstanding request batches %v", getOutstandingDigests))
	} else if instance.nullRequestTimeout > 0 {
		timeout := instance.nullRequestTimeout
		if instance.primary(instance.view) != instance.id {
			// we're waiting for the primary to deliver a null request - give it a bit more time
			timeout += instance.requestTimeout
		}
		instance.nullRequestTimer.Reset(timeout, nullRequestEvent{})
	}
}

func (instance *pbftCore) softStartTimer(timeout time.Duration, reason string) {
	logger.Debugf("Replica %d soft starting new view timer for %s: %s", instance.id, timeout, reason)
	instance.newViewTimerReason = reason
	instance.timerActive = true
	instance.newViewTimer.SoftReset(timeout, viewChangeTimerEvent{})
}

func (instance *pbftCore) startTimer(n uint64, timeout time.Duration, reason string) {
	logger.Debugf("Replica %d starting new view timer for %s: %s", instance.id, timeout, reason)
	instance.timerActive = true
	instance.newViewTimer.Reset(timeout, viewChangeTimerEvent{})
}

func (instance *pbftCore) stopTimer(n uint64) {
	logger.Debugf("Replica %d stopping a running new view timer", instance.id)
	instance.timerActive = false
	instance.newViewTimer.Stop()
}

func (instance *pbftCore) skipTo(seqNo uint64, id []byte, replicas []uint64) {
	info := &pb.BlockchainInfo{}
	err := proto.Unmarshal(id, info)
	if err != nil {
		logger.Error(fmt.Sprintf("Error unmarshaling: %s", err))
		return
	}
	//instance.UpdateState(&checkpointMessage{seqNo, id}, info, replicas)
	instance.UpdateState(seqNo, id, replicas)
}

// invalidateState is invoked to tell us that consensus realizes the ledger is out of sync
func (instance *pbftCore) invalidateState() {
	logger.Debug("Invalidating the current state")
	instance.valid = false
}

// validateState is invoked to tell us that consensus has the ledger back in sync
func (instance *pbftCore) validateState() {
	logger.Debug("Validating the current state")
	instance.valid = true
}

// UpdateState attempts to synchronize state to a particular target, implicitly calls rollback if needed
func (instance *pbftCore) UpdateState(seqNo uint64, targetId []byte, replicaId []uint64) {
	//if instance.valid {
	//	logger.Warning("State transfer is being called for, but the state has not been invalidated")
	//}

	updateStateMsg := stateUpdateHelper(seqNo, targetId, replicaId)
	instance.helper.UpdateState(updateStateMsg) // TODO: stateUpdateEvent

}

