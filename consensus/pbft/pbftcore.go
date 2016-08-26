package pbft

import (
	"fmt"
	"math/rand"
	"time"

	"hyperchain-alpha/consensus/helper"
	"hyperchain-alpha/consensus/events"
	_ "github.com/hyperledger/fabric/core" // Needed for logging format init

	"github.com/op/go-logging"
	"github.com/golang/protobuf/proto"
	"github.com/spf13/viper"
)

// =============================================================================
// init
// =============================================================================

var logger *logging.Logger // package-level logger

func init() {
	logger = logging.MustGetLogger("consensus/pbft")
}

const (
	// UnreasonableTimeout is an ugly thing, we need to create timers, then stop them before they expire, so use a large timeout
	UnreasonableTimeout = 100 * time.Hour
)

// =============================================================================
// custom interfaces and structure definitions
// =============================================================================

// Event Types

// workEvent is a temporary type, to inject work
type workEvent func()

// execDoneEvent is sent when an execution completes
type execDoneEvent struct{}

// pbftMessageEvent is sent when a consensus messages is received to be sent to pbft
type pbftMessageEvent pbftMessage

// nullRequestEvent provides "keep-alive" null requests
type nullRequestEvent struct{}


// This structure is used for incoming PBFT bound messages
type pbftMessage struct {
	sender uint64
	msg    *Message
}

type pbftCore struct {
	//internal data
	consumer helper.Stack

	// PBFT data
	byzantine     bool              // whether this node is intentionally acting as Byzantine; useful for debugging on the testnet
	f             int               // max. number of faults we can tolerate
	N             int               // max.number of validators in the network
	h             uint64            // low watermark
	id            uint64            // replica ID; PBFT `i`
	K             uint64            // checkpoint period
	logMultiplier uint64            // use this value to calculate log size : k*logMultiplier
	L             uint64            // log size
	lastExec      uint64            // last request we executed
	replicaCount  int               // number of replicas; PBFT `|R|`
	seqNo         uint64            // PBFT "n", strictly monotonic increasing sequence number
	view          uint64            // current view
	pset          map[uint64]*ViewChange_PQ
	qset          map[qidx]*ViewChange_PQ

	currentExec           *uint64                  // currently executing request
	timerActive           bool                     // is the timer running?
	requestTimeout        time.Duration            // progress timeout for requests
	outstandingReqBatches map[string]*RequestBatch // track whether we are waiting for request batches to execute

	nullRequestTimer   events.Timer  // timeout triggering a null request
	nullRequestTimeout time.Duration // duration for this timeout
						       // implementation of PBFT `in`
	reqBatchStore   map[string]*RequestBatch // track request batches
	certStore       map[msgID]*msgCert       // track quorum certificates for requests
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
	digest      string
	prePrepare  *PrePrepare
	sentPrepare bool
	prepare     []*Prepare
	sentCommit  bool
	commit      []*Commit
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

func newPbftCore(id uint64, config *viper.Viper, consumer innerStack, etf events.TimerFactory) *pbftCore {
	var err error
	instance := &pbftCore{}
	instance.id = id
	instance.consumer = consumer

	instance.nullRequestTimer = etf.CreateTimer()

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

	instance.requestTimeout, err = time.ParseDuration(config.GetString("general.timeout.request"))
	if err != nil {
		panic(fmt.Errorf("Cannot parse request timeout: %s", err))
	}
	instance.nullRequestTimeout, err = time.ParseDuration(config.GetString("general.timeout.nullrequest"))
	if err != nil {
		instance.nullRequestTimeout = 0
	}

	instance.replicaCount = instance.N

	logger.Infof("PBFT type = %T", instance.consumer)
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
	instance.pset = make(map[uint64]*ViewChange_PQ)
	instance.qset = make(map[qidx]*ViewChange_PQ)

	// initialize state transfer
	instance.outstandingReqBatches = make(map[string]*RequestBatch)

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
	case *pbftMessage:
		return pbftMessageEvent(*et)
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
	case execDoneEvent:
		instance.execDoneSync()
	case nullRequestEvent:
		instance.nullRequestHandler()
	case workEvent:
		et() // Used to allow the caller to steal use of the main thread, to be removed
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
	return n % uint64(instance.replicaCount)
}

// Is the sequence number between watermarks?
func (instance *pbftCore) inW(n uint64) bool {
	return n-instance.h > 0 && n-instance.h <= instance.L
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

// =============================================================================
// pre-prepare/prepare/commit check helper
// =============================================================================

func (instance *pbftCore) prePrepared(digest string, v uint64, n uint64) bool {
	_, mInLog := instance.reqBatchStore[digest]

	if digest != "" && !mInLog {
		return false
	}

	if q, ok := instance.qset[qidx{digest, n}]; ok && q.View == v {
		return true
	}

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

	if p, ok := instance.pset[n]; ok && p.View == v && p.BatchDigest == digest {
		return true
	}

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
	}
	return nil, fmt.Errorf("Invalid message: %v", msg)
}

func (instance *pbftCore) recvRequestBatch(reqBatch *RequestBatch) error {
	digest := hash(reqBatch)
	logger.Debugf("Replica %d received request batch %s", instance.id, digest)

	instance.reqBatchStore[digest] = reqBatch
	instance.outstandingReqBatches[digest] = reqBatch
	//instance.persistRequestBatch(digest)
	if instance.primary(instance.view) == instance.id {
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

	logger.Debugf("Primary %d broadcasting pre-prepare for view=%d/seqNo=%d and digest %s", instance.id, instance.view, n, digest)
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
	instance.innerBroadcast(&Message{Payload: &Message_PrePrepare{PrePrepare: preprep}})
	instance.maybeSendCommit(digest, instance.view, n)
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
		logger.Debugf("Replica %d has detected request batch %s must be resubmitted", instance.id, d)
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

func (instance *pbftCore) recvPrePrepare(preprep *PrePrepare) error {
	logger.Debugf("Replica %d received pre-prepare from replica %d for view=%d/seqNo=%d",
		instance.id, preprep.ReplicaId, preprep.View, preprep.SequenceNumber)

	if instance.primary(instance.view) != preprep.ReplicaId {
		logger.Warningf("Pre-prepare from other than primary: got %d, should be %d", preprep.ReplicaId, instance.primary(instance.view))
		return nil
	}

	if !instance.inWV(preprep.View, preprep.SequenceNumber) {
		// This is perfectly normal
		logger.Debugf("Replica %d pre-prepare view different, or sequence number outside watermarks: preprep.View %d, expected.View %d, seqNo %d, low-mark %d", instance.id, preprep.View, instance.primary(instance.view), preprep.SequenceNumber, instance.h)

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
		//instance.persistRequestBatch(digest)
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
		instance.recvPrepare(prep)
		return instance.innerBroadcast(&Message{Payload: &Message_Prepare{Prepare: prep}})
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
		// This is perfectly normal
		logger.Debugf("Replica %d ignoring prepare for view=%d/seqNo=%d: not in-wv, in view %d, low water mark %d", instance.id, prep.View, prep.SequenceNumber, instance.view, instance.h)
		return nil
	}

	cert := instance.getCert(prep.View, prep.SequenceNumber)

	for _, prevPrep := range cert.prepare {
		if prevPrep.ReplicaId == prep.ReplicaId {
			logger.Warningf("Ignoring duplicate prepare from %d", prep.ReplicaId)
			return nil
		}
	}
	cert.prepare = append(cert.prepare, prep)

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
		return instance.innerBroadcast(&Message{&Message_Commit{commit}})
	}
	return nil
}

func (instance *pbftCore) recvCommit(commit *Commit) error {
	logger.Debugf("Replica %d received commit from replica %d for view=%d/seqNo=%d",
		instance.id, commit.ReplicaId, commit.View, commit.SequenceNumber)

	if !instance.inWV(commit.View, commit.SequenceNumber) {
		// This is perfectly normal
		logger.Debugf("Replica %d ignoring commit for view=%d/seqNo=%d: not in-wv, in view %d, high water mark %d", instance.id, commit.View, commit.SequenceNumber, instance.view, instance.h)
		return nil
	}

	cert := instance.getCert(commit.View, commit.SequenceNumber)
	for _, prevCommit := range cert.commit {
		if prevCommit.ReplicaId == commit.ReplicaId {
			logger.Warningf("Ignoring duplicate commit from %d", commit.ReplicaId)
			return nil
		}
	}
	cert.commit = append(cert.commit, commit)

	if instance.committed(commit.BatchDigest, commit.View, commit.SequenceNumber) {
		delete(instance.outstandingReqBatches, commit.BatchDigest)

		instance.executeOutstanding()
	}

	return nil
}

func (instance *pbftCore) executeOutstanding() {
	if instance.currentExec != nil {
		logger.Debugf("Replica %d not attempting to executeOutstanding because it is currently executing %d", instance.id, *instance.currentExec)
		return
	}
	logger.Debugf("Replica %d attempting to executeOutstanding", instance.id)

	for idx := range instance.certStore {
		if instance.executeOne(idx) {
			break
		}
	}

	logger.Debugf("Replica %d certstore %+v", instance.id, instance.certStore)

	instance.startTimerIfOutstandingRequests()
}

func (instance *pbftCore) executeOne(idx msgID) bool {
	cert := instance.certStore[idx]

	if idx.n != instance.lastExec+1 || cert == nil || cert.prePrepare == nil {
		return false
	}

	// we now have the right sequence number that doesn't create holes

	digest := cert.digest
	reqBatch := instance.reqBatchStore[digest]

	if !instance.committed(digest, idx.v, idx.n) {
		return false
	}

	// we have a commit certificate for this request batch
	currentExec := idx.n
	instance.currentExec = &currentExec

	// null request
	if digest == "" {
		logger.Infof("Replica %d executing/committing null request for view=%d/seqNo=%d",
			instance.id, idx.v, idx.n)
		instance.execDoneSync()
	} else {
		logger.Infof("Replica %d executing/committing request batch for view=%d/seqNo=%d and digest %s",
			instance.id, idx.v, idx.n, digest)
		// synchronously execute, it is the other side's responsibility to execute in the background if needed
		instance.consumer.execute(idx.n, reqBatch)
	}
	return true
}

func (instance *pbftCore) execDoneSync() {
	if instance.currentExec != nil {
		logger.Infof("Replica %d finished execution %d, trying next", instance.id, *instance.currentExec)
		instance.lastExec = *instance.currentExec

	} else {
		// XXX This masks a bug, this should not be called when currentExec is nil
		logger.Warningf("Replica %d had execDoneSync called, flagging ourselves as out of date", instance.id)
	}
	instance.currentExec = nil

	instance.executeOutstanding()
}

// =============================================================================
// Misc. methods go here
// =============================================================================

// Marshals a Message and hands it to the Stack. If toSelf is true,
// the message is also dispatched to the local instance's RecvMsgSync.
func (instance *pbftCore) innerBroadcast(msg *Message) error {
	msgRaw, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("Cannot marshal message %s", err)
	}

	doByzantine := false
	if instance.byzantine {
		rand1 := rand.New(rand.NewSource(time.Now().UnixNano()))
		doIt := rand1.Intn(3) // go byzantine about 1/3 of the time
		if doIt == 1 {
			doByzantine = true
		}
	}

	// testing byzantine fault.
	if doByzantine {
		rand2 := rand.New(rand.NewSource(time.Now().UnixNano()))
		ignoreidx := rand2.Intn(instance.N)
		for i := 0; i < instance.N; i++ {
			if i != ignoreidx && uint64(i) != instance.id { //Pick a random replica and do not send message
				instance.consumer.unicast(msgRaw, uint64(i))
			} else {
				logger.Debugf("PBFT byzantine: not broadcasting to replica %v", i)
			}
		}
	} else {
		instance.consumer.broadcast(msgRaw)
	}
	return nil
}

func (instance *pbftCore) startTimerIfOutstandingRequests() {

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
	instance.timerActive = true
}
