package pbft

import (
	"encoding/base64"
	"fmt"
	"sort"
	"sync"
	"time"

	"hyperchain/consensus/events"
	"hyperchain/consensus/helper"
	"hyperchain/core/types"
	"hyperchain/event"
	"hyperchain/protos"

	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

var logger *logging.Logger // package-level logger

func init() {
	logger = logging.MustGetLogger("consensus/pbft")
}

// batch is used to construct reqbatch, the middle layer between outer to pbft
type pbftProtocal struct {
	batchTimer       events.Timer
	batchTimerActive bool
	batchTimeout     time.Duration
	batchSize        int
	batchStore       []*types.Transaction //ordered message batch
	helper           helper.Stack
	batchManager     events.Manager
	pbftManager      events.Manager
	mux              sync.Mutex
	reqStore         *requestStore //received messages

	// PBFT data
	activeView     bool   // view change happening
	byzantine      bool   // whether this node is intentionally acting as Byzantine; useful for debugging on the testnet
	f              int    // max. number of faults we can tolerate
	N              int    // max.number of validators in the network
	h              uint64 // low watermark
	id             uint64 // replica ID; PBFT `i`
	K              uint64 // checkpoint period
	logMultiplier  uint64 // use this value to calculate log size : k*logMultiplier
	L              uint64 // log size
	lastExec       uint64 // last request we executed
	replicaCount   int    // number of replicas; PBFT `|R|`
	seqNo          uint64 // PBFT "n", strictly monotonic increasing sequence number
	view           uint64 // current view
	nvInitialSeqNo uint64 // initial seqNo in a new view
	valid          bool   // whether we believe the state is up to date

	chkpts map[uint64]string         // state checkpoints; map lastExec to global hash
	pset   map[uint64]*ViewChange_PQ // state checkpoints; map lastExec to global hash
	qset   map[qidx]*ViewChange_PQ   // state checkpoints; map lastExec to global hash

	skipInProgress    bool               // Set when we have detected a fall behind scenario until we pick a new starting point
	stateTransferring bool               // Set when state transfer is executing
	highStateTarget   *stateUpdateTarget // Set to the highest weak checkpoint cert we have observed
	hChkpts           map[uint64]uint64  // highest checkpoint sequence number observed for each replica

	currentExec           *uint64                      // currently executing request
	timerActive           bool                         // is the timer running?
	vcResendTimer         events.Timer                 // timer triggering resend of a view change
	vcResendTimeout       time.Duration                // timeout before resending view change
	requestTimeout        time.Duration                // progress timeout for requests
	lastNewViewTimeout    time.Duration                // last timeout we used during this view change
	outstandingReqBatches map[string]*TransactionBatch // track whether we are waiting for request batches to execute
	newViewTimeout        time.Duration                // progress timeout for new views
	newViewTimer          events.Timer                 // track the timeout for each requestBatch
	newViewTimerReason    string                       // what triggered the timer
	nullRequestTimer      events.Timer                 // timeout triggering a null request
	nullRequestTimeout    time.Duration                // duration for this timeout

	viewChangePeriod  uint64          // period between automatic view changes
	viewChangeSeqNo   uint64          // next seqNo to perform view change
	missingReqBatches map[string]bool // for all the assigned, non-checkpointed request batches we might be missing during view-change

	// implementation of PBFT `in`
	reqBatchStore   map[string]*TransactionBatch // track request batches
	certStore       map[msgID]*msgCert           // track quorum certificates for requests
	checkpointStore map[Checkpoint]bool          // track checkpoints as set
	committedCert   map[msgID]string             // track the committed cert to help excute
	chkptCertStore  map[chkptID]*chkptCert       // track quorum certificates for checkpoints
	newViewStore    map[uint64]*NewView          // track last new-view we received or sent
	viewChangeStore map[vcidx]*ViewChange        // track view-change messages

	// implement the validate transaction batch process
	vid                 	uint64				// track the validate squence number
	lastVid             	uint64                       	// track the last validate batch seqNo
	currentVid          	*uint64                      	// track the current validate batch seqNo
	validatedBatchStore 	map[string]*TransactionBatch 	// track the validated transaction rnnbatch
	cacheValidatedBatch 	map[string]*cacheBatch       	// track the cached validated batch
	validateTimer		events.Timer
	validateTimeout		time.Duration

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
	digest       string
	rawDigest    string
	prePrepare   *PrePrepare
	sentPrepare  bool
	prepare      map[Prepare]bool
	prepareCount int
	sentValidate bool
	sentCommit   bool
	commit       map[Commit]bool
	commitCount  int
	sentExecute  bool
}

type chkptID struct {
	n  uint64
	id string
}

type chkptCert struct {
	chkpts     map[Checkpoint]bool
	chkptCount int
}

type vcidx struct {
	v  uint64
	id uint64
}

type cacheBatch struct {
	batch     *TransactionBatch
	rawDigest string
	vid       uint64
}

// newBatch initializes a batch
func newPbft(id uint64, config *viper.Viper, h helper.Stack) *pbftProtocal {
	var err error
	pbft := &pbftProtocal{}

	pbft.helper = h
	pbft.id = id

	// pbftManager is used to solve pbft message
	pbft.pbftManager = events.NewManagerImpl()
	pbft.pbftManager.SetReceiver(pbft)
	pbftTimerFactory := events.NewTimerFactoryImpl(pbft.pbftManager)

	pbft.vcResendTimer = pbftTimerFactory.CreateTimer()
	pbft.nullRequestTimer = pbftTimerFactory.CreateTimer()
	pbft.newViewTimer = pbftTimerFactory.CreateTimer()
	pbft.N = config.GetInt("general.N")
	pbft.f = config.GetInt("general.f")

	if pbft.f*3+1 > pbft.N {
		panic(fmt.Sprintf("need at least %d enough replicas to tolerate %d byzantine faults, but only %d replicas configured", pbft.f*3+1, pbft.f, pbft.N))
	}

	pbft.K = uint64(config.GetInt("general.K"))

	pbft.logMultiplier = uint64(config.GetInt("general.logmultiplier"))
	if pbft.logMultiplier < 2 {
		panic("Log multiplier must be greater than or equal to 2")
	}

	pbft.L = pbft.logMultiplier * pbft.K // log size
	pbft.viewChangePeriod = uint64(config.GetInt("general.viewchangeperiod"))
	pbft.byzantine = config.GetBool("general.byzantine")

	pbft.vcResendTimeout, err = time.ParseDuration(config.GetString("timeout.resendviewchange"))
	if err != nil {
		panic(fmt.Errorf("Cannot parse resendviewchange timeout: %s", err))
	}

	pbft.newViewTimeout, err = time.ParseDuration(config.GetString("timeout.viewchange"))
	if err != nil {
		panic(fmt.Errorf("Cannot parse viewchange timeout: %s", err))
	}

	pbft.requestTimeout, err = time.ParseDuration(config.GetString("timeout.request"))
	if err != nil {
		panic(fmt.Errorf("Cannot parse request timeout: %s", err))
	}

	pbft.nullRequestTimeout, err = time.ParseDuration(config.GetString("timeout.nullrequest"))
	if err != nil {
		pbft.nullRequestTimeout = 0
	}

	pbft.activeView = true
	pbft.replicaCount = pbft.N

	logger.Infof("PBFT Max number of validating peers (N) = %v", pbft.N)
	logger.Infof("PBFT Max number of failing peers (f) = %v", pbft.f)
	logger.Infof("PBFT byzantine flag = %v", pbft.byzantine)
	logger.Infof("PBFT request timeout = %v", pbft.requestTimeout)
	logger.Infof("PBFT Checkpoint period (K) = %v", pbft.K)
	logger.Infof("PBFT Log multiplier = %v", pbft.logMultiplier)
	logger.Infof("PBFT log size (L) = %v", pbft.L)

	if pbft.nullRequestTimeout > 0 {
		logger.Infof("PBFT null requests timeout = %v", pbft.nullRequestTimeout)
	} else {
		logger.Infof("PBFT null requests disabled")
	}

	if pbft.viewChangePeriod > 0 {
		logger.Infof("PBFT view change period = %v", pbft.viewChangePeriod)
	} else {
		logger.Infof("PBFT automatic view change disabled")
	}

	// init the logs
	pbft.certStore = make(map[msgID]*msgCert)
	pbft.reqBatchStore = make(map[string]*TransactionBatch)
	pbft.checkpointStore = make(map[Checkpoint]bool)
	pbft.chkpts = make(map[uint64]string)
	pbft.pset = make(map[uint64]*ViewChange_PQ)
	pbft.qset = make(map[qidx]*ViewChange_PQ)
	pbft.committedCert = make(map[msgID]string)
	pbft.chkptCertStore = make(map[chkptID]*chkptCert)
	pbft.newViewStore = make(map[uint64]*NewView)
	pbft.viewChangeStore = make(map[vcidx]*ViewChange)
	pbft.missingReqBatches = make(map[string]bool)

	// initialize state transfer
	pbft.hChkpts = make(map[uint64]uint64)

	pbft.chkpts[0] = "XXX GENESIS"

	pbft.lastNewViewTimeout = pbft.newViewTimeout
	pbft.outstandingReqBatches = make(map[string]*TransactionBatch)
	pbft.validatedBatchStore = make(map[string]*TransactionBatch)
	pbft.cacheValidatedBatch = make(map[string]*cacheBatch)

	pbft.restoreState()

	pbft.viewChangeSeqNo = ^uint64(0) // infinity
	pbft.updateViewChangeSeqNo()

	pbft.pbftManager.Start()

	// batchManager is used to solve batch message, like *Request
	pbft.batchManager = events.NewManagerImpl()
	pbft.batchManager.SetReceiver(pbft)
	etf := events.NewTimerFactoryImpl(pbft.batchManager)
	pbft.batchManager.Start()

	// initialize the batchTimeout
	pbft.batchTimer = etf.CreateTimer()
	pbft.batchSize = config.GetInt("general.batchsize")
	pbft.batchStore = nil
	pbft.batchTimeout, err = time.ParseDuration(config.GetString("timeout.batch"))
	if err != nil {
		panic(fmt.Errorf("Cannot parse batch timeout: %s", err))
	}

	if pbft.batchTimeout >= pbft.requestTimeout {
		pbft.requestTimeout = 3 * pbft.batchTimeout / 2
		logger.Warningf("Configured request timeout must be greater than batch timeout, setting to %v", pbft.requestTimeout)
	}

	if pbft.requestTimeout >= pbft.nullRequestTimeout && pbft.nullRequestTimeout != 0 {
		pbft.nullRequestTimeout = 3 * pbft.requestTimeout / 2
		logger.Warningf("Configured null request timeout must be greater than request timeout, setting to %v", pbft.nullRequestTimeout)
	}

	pbft.validateTimer = etf.CreateTimer()
	pbft.validateTimeout, err = time.ParseDuration(config.GetString("timeout.validate"))
	if err != nil {
		panic(fmt.Errorf("Cannot parse validate timeout: %s", err))
	}

	logger.Infof("PBFT Batch size = %d", pbft.batchSize)
	logger.Infof("PBFT Batch timeout = %v", pbft.batchTimeout)
	pbft.reqStore = newRequestStore()

	logger.Noticef("--------PBFT finish start, nodeID: %d--------", pbft.id)

	return pbft
}

// Close tells us to release resources we are holding
func (pbft *pbftProtocal) Close() {
	pbft.batchTimer.Halt()
	pbft.newViewTimer.Halt()
	pbft.nullRequestTimer.Halt()
}

// RecvMsg is used by outer to send message to consensus
func (pbft *pbftProtocal) RecvMsg(e []byte) error {

	msg := &protos.Message{}
	err := proto.Unmarshal(e, msg)
	if err != nil {
		logger.Errorf("Inner RecvMsg Unmarshal error: can not unmarshal pb.Message", err)
		return err
	}

	if msg.Type == protos.Message_TRANSACTION {
		return pbft.processTransaction(msg)
	} else if msg.Type == protos.Message_CONSENSUS {
		return pbft.processConsensus(msg)
	} else if msg.Type == protos.Message_STATE_UPDATED {
		return pbft.processStateUpdated(msg)
	} else if msg.Type == protos.Message_NULL_REQUEST {
		return pbft.processNullRequest(msg)
	}

	logger.Errorf("Unknown recvMsg: %+v", msg)

	return nil
}

func (pbft *pbftProtocal) RecvValidatedResult(result event.ValidatedTxs) error {

	primary := pbft.primary(pbft.view)
	if primary == pbft.id {
		logger.Debugf("Primary %d recived validated batch for sqeNo=%d, batch is: %s", pbft.id, result.SeqNo, result.Digest)

		if !pbft.inWV(result.View, result.SeqNo) {
			logger.Debugf("Replica %d is primary, receives validated result %s that is out of sequence numbers", pbft.id, result.Digest)
			return nil
		}

		if len(result.Transactions) == 0 {
			logger.Debugf("Replica %d is primary, receives validated result %s that is empty", pbft.id, result.Digest)
			pbft.sendNullRequest()
			return nil
		}

		batch := &TransactionBatch{
			Batch:     result.Transactions,
			Timestamp: time.Now().UnixNano(),
		}
		digest := byteToString(result.Hash)
		pbft.validatedBatchStore[digest] = batch
		pbft.outstandingReqBatches[digest] = batch
		cache := &cacheBatch{
			batch:     batch,
			rawDigest: result.Digest,
			vid:       result.SeqNo,
		}
		pbft.cacheValidatedBatch[digest] = cache

		pbft.trySendPrePrepare()
	} else {
		logger.Debugf("Replica %d recived validated batch for sqeNo=%d, batch is: %s", pbft.id, result.SeqNo, result.Digest)

		if !pbft.inWV(result.View, result.SeqNo) {
			logger.Debugf("Replica %d receives validated result %s that is out of sequence numbers", pbft.id, result.Digest)
			return nil
		}

		cert := pbft.getCert(result.View, result.SeqNo)


		digest := byteToString(result.Hash)

		//logger.Notice("Replica  recived seqNo is sqeNo=%d, module digest is: %s,cert digest is: %s",result.SeqNo, result.Digest,cert.digest)


		if digest == cert.digest {
			pbft.sendCommit(digest, result.View, result.SeqNo)
		} else {
			pbft.sendViewChange()
		}
	}

	return nil
}

func (pbft *pbftProtocal) ProcessEvent(ee events.Event) events.Event {

	logger.Debugf("Replica %d start solve event", pbft.id)

	switch e := ee.(type) {

	case *types.Transaction:
		tx := e
		return pbft.processTxEvent(tx)
	case viewChangedEvent:
		pbft.processRequestsDuringViewChange()
	case batchTimerEvent:
		logger.Debugf("Replica %d batch timer expired", pbft.id)
		if pbft.activeView && (len(pbft.batchStore) > 0) {
			return pbft.sendBatch()
		}
	default:
		logger.Debugf("batch processEvent, default: %+v", e)
		return pbft.processPbftEvent(e)
	}
	return nil
}

// allow the view-change protocol to kick-off when the timer expires
func (pbft *pbftProtocal) processPbftEvent(e events.Event) events.Event {

	var err error
	logger.Debugf("Replica %d processing event", pbft.id)

	switch et := e.(type) {
	case viewChangeTimerEvent:
		logger.Infof("Replica %d view change timer expired, sending view change: %s", pbft.id, pbft.newViewTimerReason)
		pbft.timerActive = false
		pbft.sendViewChange()
	case *ConsensusMessage:
		next, err := pbft.eventToMsg(et)
		if err != nil {
			break
		}
		return next
	case *TransactionBatch:
		err = pbft.recvRequestBatch(et)
	case *PrePrepare:
		err = pbft.recvPrePrepare(et)
	case *Prepare:
		err = pbft.recvPrepare(et)
	case *Commit:
		err = pbft.recvCommit(et)
	case *Checkpoint:
		return pbft.recvCheckpoint(et)
	case *stateUpdatedEvent:
		//pbft.batch.reqStore = newRequestStore()
		err = pbft.recvStateUpdatedEvent(et)
	case *ViewChange:
		return pbft.recvViewChange(et)
	case *NewView:
		return pbft.recvNewView(et)
	case *FetchRequestBatch:
		err = pbft.recvFetchRequestBatch(et)
	case returnRequestBatchEvent:
		return pbft.recvReturnRequestBatch(et)
	case viewChangeQuorumEvent:
		logger.Debugf("Replica %d received view change quorum, processing new view", pbft.id)
		if pbft.primary(pbft.view) == pbft.id {
			return pbft.sendNewView()
		}
		return pbft.processNewView()
	case nullRequestEvent:
		pbft.nullRequestHandler()
	case viewChangeResendTimerEvent:
		if pbft.activeView {
			logger.Warningf("Replica %d had its view change resend timer expire but it's in an active view, this is benign but may indicate a bug", pbft.id)
			return nil
		}
		logger.Debugf("Replica %d view change resend timer expired before view change quorum was reached, resending", pbft.id)
		pbft.view-- // sending the view change increments this
		return pbft.sendViewChange()
	default:
		logger.Warningf("Replica %d received an unknown message type %T", pbft.id, et)
	}

	if err != nil {
		logger.Warning(err.Error())
	}

	return nil
}

// =============================================================================
// batch related methods
// =============================================================================

// process the trasaction message
func (pbft *pbftProtocal) processTransaction(msg *protos.Message) error {

	// Parse the transaction payload to transaction
	tx := &types.Transaction{}
	err := proto.Unmarshal(msg.Payload, tx)
	if err != nil {
		logger.Errorf("processTransaction Unmarshal error: can not unmarshal protos.Message", err)
		return err
	}

	// Post a requestEvent
	go pbft.postRequestEvent(tx)

	return nil
}

// process the consensus message
func (pbft *pbftProtocal) processConsensus(msg *protos.Message) error {

	consensus := &ConsensusMessage{}
	err := proto.Unmarshal(msg.Payload, consensus)
	if err != nil {
		logger.Errorf("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err)
		return err
	}

	if consensus.Type == ConsensusMessage_TRANSACTION {
		tx := &types.Transaction{}
		err := proto.Unmarshal(consensus.Payload, tx)
		if err != nil {
			logger.Errorf("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err)
			return err
		}
		go pbft.postRequestEvent(tx)
		return nil
	} else {
		pbft.postPbftEvent(consensus)
		return nil
	}

}

// process the state update message
func (pbft *pbftProtocal) processStateUpdated(msg *protos.Message) error {

	stateUpdatedMsg := &protos.StateUpdatedMessage{}
	err := proto.Unmarshal(msg.Payload, stateUpdatedMsg)

	if err != nil {
		logger.Errorf("processStateUpdate, unmarshal error: can not unmarshal UpdateStateMessage", err)
		return err
	}

	e := &stateUpdatedEvent{
		seqNo: stateUpdatedMsg.SeqNo,
	}
	pbft.postPbftEvent(e)
	return nil
}

// processNullRequest process when a null request come
func (pbft *pbftProtocal) processNullRequest(msg *protos.Message) error {
	pbft.nullReqTimerReset()
	return nil
}

func (pbft *pbftProtocal) processTxEvent(tx *types.Transaction) error {

	primary := pbft.primary(pbft.view)
	if !pbft.activeView {
		pbft.reqStore.storeOutstanding(tx)
	} else if primary != pbft.id {
		//Broadcast request to primary
		payload, err := proto.Marshal(tx)
		if err != nil {
			logger.Errorf("CConsensusMessage_TRANSACTION Marshal Error", err)
			return nil
		}
		consensusMsg := &ConsensusMessage{
			Type:    ConsensusMessage_TRANSACTION,
			Payload: payload,
		}
		pbMsg := consensusMsgHelper(consensusMsg, pbft.id)
		pbft.helper.InnerUnicast(pbMsg, primary)
	} else {
		return pbft.leaderProcReq(tx)
	}

	return nil
}

func (pbft *pbftProtocal) processRequestsDuringViewChange() error {
	primary := pbft.primary(pbft.view)
	if pbft.activeView && primary != pbft.id {
		for pbft.reqStore.outstandingRequests.Len() != 0 {

			temp := pbft.reqStore.outstandingRequests.order.Front().Value

			reqc, ok := interface{}(temp).(requestContainer)
			if !ok {
				logger.Error("type assert error:", temp)
				return nil
			}
			req := reqc.req
			if req != nil {
				payload, err := proto.Marshal(req)
				if err != nil {
					logger.Errorf("ConsensusMessage_TRANSACTION Marshal Error", err)
					return nil
				}
				consensusMsg := &ConsensusMessage{
					Type:    ConsensusMessage_TRANSACTION,
					Payload: payload,
				}
				pbMsg := consensusMsgHelper(consensusMsg, pbft.id)
				pbft.helper.InnerUnicast(pbMsg, primary)
				pbft.reqStore.remove(req)
			}
		}
	}
	return nil
}

func (pbft *pbftProtocal) leaderProcReq(tx *types.Transaction) error {

	logger.Debugf("Batch primary %d queueing new request", pbft.id)
	pbft.batchStore = append(pbft.batchStore, tx)

	if !pbft.batchTimerActive {
		pbft.startBatchTimer()
	}

	if len(pbft.batchStore) >= pbft.batchSize {
		return pbft.sendBatch()
	}

	return nil
}

func (pbft *pbftProtocal) sendBatch() error {

	pbft.stopBatchTimer()

	if len(pbft.batchStore) == 0 {
		logger.Error("Told to send an empty batch store for ordering, ignoring")
		return nil
	}

	reqBatch := &TransactionBatch{
		Batch:     pbft.batchStore,
		Timestamp: time.Now().UnixNano(),
	}
	pbft.batchStore = nil
	logger.Infof("Creating batch with %d requests", len(reqBatch.Batch))

	pbft.pbftManager.Queue() <- reqBatch

	return nil
}

// =============================================================================
// receive methods
// =============================================================================

func (pbft *pbftProtocal) nullRequestHandler() {

	if !pbft.activeView {
		return
	}

	if pbft.primary(pbft.view) != pbft.id {
		// backup expected a null request, but primary never sent one
		logger.Infof("Replica %d null request timer expired, sending view change", pbft.id)

		pbft.sendViewChange()
	} else {
		// time for the primary to send a null request
		// pre-prepare with null digest
		//todo test
		//if pbft.logstatic.Blockbool{
		//	pbft.logstatic.RecordCount("block_SendNullRequest_nullRequestHandler","currentView")
		//	return
		//}
		logger.Infof("Primary %d null request timer expired, sending null request", pbft.id)
		pbft.sendNullRequest()
	}
}

func (pbft *pbftProtocal) eventToMsg(msg *ConsensusMessage) (interface{}, error) {

	switch msg.Type {
	case ConsensusMessage_TRANSACTION:
		tx := &types.Transaction{}
		err := proto.Unmarshal(msg.Payload, tx)
		if err != nil {
			logger.Error("Unmarshal error, can not unmarshal ConsensusMessage_TRANSACTION:", err)
			return nil, err
		} else {
			return nil, fmt.Errorf("Unresolved ConsensusMessage_Transaction: %+v", tx)
		}
	case ConsensusMessage_TRANSATION_BATCH:
		txBatch := &TransactionBatch{}
		err := proto.Unmarshal(msg.Payload, txBatch)
		if err != nil {
			logger.Error("Unmarshal error, can not unmarshal ConsensusMessage_TRANSATION_BATCH:", err)
			return nil, err
		}
		return txBatch, nil
	case ConsensusMessage_PRE_PREPARE:
		preprep := &PrePrepare{}
		err := proto.Unmarshal(msg.Payload, preprep)
		if err != nil {
			logger.Error("Unmarshal error, can not unmarshal ConsensusMessage_PRE_PREPARE:", err)
			return nil, err
		}
		return preprep, nil
	case ConsensusMessage_PREPARE:
		prep := &Prepare{}
		err := proto.Unmarshal(msg.Payload, prep)
		if err != nil {
			logger.Error("Unmarshal error, can not unmarshal ConsensusMessage_PREPARE:", err)
			return nil, err
		}
		return prep, nil
	case ConsensusMessage_COMMIT:
		commit := &Commit{}
		err := proto.Unmarshal(msg.Payload, commit)
		if err != nil {
			logger.Error("Unmarshal error, can not unmarshal ConsensusMessage_COMMIT:", err)
			return nil, err
		}
		return commit, nil
	case ConsensusMessage_CHECKPOINT:
		chkpt := &Checkpoint{}
		err := proto.Unmarshal(msg.Payload, chkpt)
		if err != nil {
			logger.Error("Unmarshal error, can not unmarshal ConsensusMessage_CHECKPOINT:", err)
			return nil, err
		}
		return chkpt, nil
	case ConsensusMessage_VIEW_CHANGE:
		vc := &ViewChange{}
		err := proto.Unmarshal(msg.Payload, vc)
		if err != nil {
			logger.Error("Unmarshal error, can not unmarshal ConsensusMessage_VIEW_CHANGE:", err)
			return nil, err
		}
		return vc, nil
	case ConsensusMessage_NEW_VIEW:
		nv := &NewView{}
		err := proto.Unmarshal(msg.Payload, nv)
		if err != nil {
			logger.Error("Unmarshal error, can not unmarshal ConsensusMessage_NEW_VIEW:", err)
			return nil, err
		}
		return nv, nil
	case ConsensusMessage_FRTCH_REQUEST_BATCH:
		frb := &FetchRequestBatch{}
		err := proto.Unmarshal(msg.Payload, frb)
		if err != nil {
			logger.Error("Unmarshal error, can not unmarshal ConsensusMessage_FRTCH_REQUEST_BATCH:", err)
			return nil, err
		}
		return frb, nil
	case ConsensusMessage_RETURN_REQUEST_BATCH:
		rrb := &TransactionBatch{}
		err := proto.Unmarshal(msg.Payload, rrb)
		if err != nil {
			logger.Error("Unmarshal error, can not unmarshal ConsensusMessage_RETURN_REQUEST_BATCH:", err)
			return nil, err
		}
		return returnRequestBatchEvent(rrb), nil
	default:
		return nil, fmt.Errorf("Invalid message: %v", msg)
	}

}

func (pbft *pbftProtocal) recvStateUpdatedEvent(et *stateUpdatedEvent) error {

	pbft.stateTransferring = false
	// If state transfer did not complete successfully, or if it did not reach our low watermark, do it again
	if et.seqNo < pbft.h {
		logger.Warningf("Replica %d recovered to seqNo %d but our low watermark has moved to %d", pbft.id, et.seqNo, pbft.h)
		if pbft.highStateTarget == nil {
			logger.Debugf("Replica %d has no state targets, cannot resume state transfer yet", pbft.id)
		} else if et.seqNo < pbft.highStateTarget.seqNo {
			logger.Debugf("Replica %d has state target for %d, transferring", pbft.id, pbft.highStateTarget.seqNo)
			pbft.retryStateTransfer(nil)
		} else {
			logger.Debugf("Replica %d has no state target above %d, highest is %d", pbft.id, et.seqNo, pbft.highStateTarget.seqNo)
		}
		return nil
	}

	logger.Infof("Replica %d application caught up via state transfer, lastExec now %d", pbft.id, et.seqNo)
	// XXX create checkpoint
	pbft.lastExec = et.seqNo
	pbft.moveWatermarks(pbft.lastExec) // The watermark movement handles moving this to a checkpoint boundary
	pbft.skipInProgress = false
	pbft.validateState()
	pbft.executeOutstanding()

	return nil
}

func (pbft *pbftProtocal) recvRequestBatch(reqBatch *TransactionBatch) error {

	digest := hash(reqBatch)
	logger.Debugf("Replica %d received request batch %s", pbft.id, digest)

	pbft.reqBatchStore[digest] = reqBatch
	//pbft.persistRequestBatch(digest)
	if pbft.activeView {
		pbft.softStartTimer(pbft.requestTimeout, fmt.Sprintf("new request batch %s", digest))
	}
	if pbft.primary(pbft.view) == pbft.id && pbft.activeView {
		pbft.nullRequestTimer.Stop()
		pbft.validateBatch(reqBatch, digest, 0, 0)
	} else {
		logger.Debugf("Replica %d is backup, not sending pre-prepare for request batch %s", pbft.id, digest)
	}

	return nil
}

// sendNullRequest is for primary peer to send null when nullRequestTimer booms
func (pbft *pbftProtocal) sendNullRequest() {
	nullRequest := nullRequestMsgHelper(pbft.id)
	pbft.helper.InnerBroadcast(nullRequest)
	pbft.nullReqTimerReset()
}

func (pbft *pbftProtocal) validateBatch(txBatch *TransactionBatch, digest string, vid uint64, view uint64) {

	primary := pbft.primary(pbft.view)
	if primary == pbft.id {
		logger.Debugf("Primary %d try to  validate for batch: %+v", pbft.id, digest)

		n := pbft.vid + 1
		if !pbft.inWV(pbft.view, n) {
			logger.Debugf("Replica %d is primary, not validating for transaction batch %s because it is out of sequence numbers", pbft.id, digest)
			return
		}

		pbft.vid = n
		pbft.helper.ValidateBatch(txBatch.Batch, n, pbft.view, digest, true)
	} else {
		logger.Debugf("Replica %d try to  validate for batch: %+v", pbft.id, digest)

		if !pbft.inWV(pbft.view, vid) {
			logger.Debugf("Replica %d not validating for transaction batch %s because it is out of sequence numbers", pbft.id, digest)
			return
		}
		pbft.helper.ValidateBatch(txBatch.Batch, vid, view, digest, false)
	}

}

func (pbft *pbftProtocal) trySendPrePrepare() {


	if pbft.currentVid != nil {
		logger.Debugf("Replica %d not attempting to send pre-prepare bacause it is currently send %d, retry.", pbft.id, pbft.currentVid)
	}

	logger.Debugf("Replica %d attempting to call sendPrePrepare", pbft.id)

	for digest := range pbft.cacheValidatedBatch {
		if pbft.callSendPrePrepare(digest) {
			break
		}
	}
}

func (pbft *pbftProtocal) callSendPrePrepare(digest string) bool {

	cache := pbft.cacheValidatedBatch[digest]

	if cache == nil {
		logger.Debugf("Primary %d already call sendPrePrepare for batch: %d", pbft.id, digest)
		return false
	}

	if cache.vid != pbft.lastVid+1 {
		logger.Debugf("Primary %d hasn't done with last send pre-prepare, vid=%d", pbft.id, pbft.lastVid)
		return false
	}

	currentVid := cache.vid
	pbft.currentVid = &currentVid
	pbft.sendPrePrepare(cache.batch, digest, cache.rawDigest)

	return true
}

func (pbft *pbftProtocal) sendPrePrepare(reqBatch *TransactionBatch, digest string, rawDigest string) {

	logger.Debugf("Replica %d is primary, issuing pre-prepare for request batch %s", pbft.id, digest)

	n := pbft.seqNo + 1

	for _, cert := range pbft.certStore { // check for other PRE-PREPARE for same digest, but different seqNo
		if p := cert.prePrepare; p != nil {
			if p.View == pbft.view && p.SequenceNumber != n && p.BatchDigest == digest && digest != "" {
				logger.Infof("Other pre-prepare found with same digest but different seqNo: %d instead of %d", p.SequenceNumber, n)
				return
			}
		}
	}

	if !pbft.inWV(pbft.view, n) {
		logger.Debugf("Replica %d is primary, not sending pre-prepare for request batch %s because it is out of sequence numbers", pbft.id, digest)
		return
	}

	logger.Debugf("Primary %d broadcasting pre-prepare for view=%d/seqNo=%d", pbft.id, pbft.view, n)
	pbft.seqNo = n
	preprep := &PrePrepare{
		View:             pbft.view,
		SequenceNumber:   n,
		BatchDigest:      digest,
		TransactionBatch: reqBatch,
		ReplicaId:        pbft.id,
	}
	cert := pbft.getCert(pbft.view, n)
	cert.prePrepare = preprep
	cert.digest = digest
	cert.rawDigest = rawDigest
	cert.sentValidate = true
	delete(pbft.cacheValidatedBatch, digest)
	//pbft.persistQSet()
	payload, err := proto.Marshal(preprep)
	logger.Debug("call---send pre-pare Replica %d received pre-prepare from replica %d for view=%d/seqNo=%d, digest: ",
		pbft.id, preprep.ReplicaId, preprep.View, preprep.SequenceNumber, preprep.BatchDigest)

	if err != nil {
		logger.Errorf("ConsensusMessage_PRE_PREPARE Marshal Error", err)
		return
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_PRE_PREPARE,
		Payload: payload,
	}
	msg := consensusMsgHelper(consensusMsg, pbft.id)
	pbft.helper.InnerBroadcast(msg)

	pbft.lastVid = *pbft.currentVid
	pbft.currentVid = nil

	pbft.maybeSendCommit(digest, pbft.view, n)
}

func (pbft *pbftProtocal) recvPrePrepare(preprep *PrePrepare) error {

	//
	//logger.Notice("receive  pre-prepare first seq is:",preprep.SequenceNumber)

	logger.Debug("Replica %d received pre-prepare from replica %d for view=%d/seqNo=%d, digest: ",
		pbft.id, preprep.ReplicaId, preprep.View, preprep.SequenceNumber, preprep.BatchDigest)

	if !pbft.activeView {
		logger.Debugf("Replica %d ignoring pre-prepare as we sre in view change", pbft.id)
		return nil
	}

	if pbft.primary(pbft.view) != preprep.ReplicaId {
		logger.Warningf("Pre-prepare from other than primary: got %d, should be %d", preprep.ReplicaId, pbft.primary(pbft.view))
		return nil
	}

	if !pbft.inWV(preprep.View, preprep.SequenceNumber) {
		if preprep.SequenceNumber != pbft.h && !pbft.skipInProgress {
			logger.Warningf("Replica %d pre-prepare view different, or sequence number outside watermarks: preprep.View %d, expected.View %d, seqNo %d, low-mark %d", pbft.id, preprep.View, pbft.primary(pbft.view), preprep.SequenceNumber, pbft.h)
		} else {
			// This is perfectly normal
			logger.Debugf("Replica %d pre-prepare view different, or sequence number outside watermarks: preprep.View %d, expected.View %d, seqNo %d, low-mark %d", pbft.id, preprep.View, pbft.primary(pbft.view), preprep.SequenceNumber, pbft.h)
		}

		return nil
	}

	cert := pbft.getCert(preprep.View, preprep.SequenceNumber)

	if cert.digest != "" && cert.digest != preprep.BatchDigest {
		logger.Warningf("Pre-prepare found for same view/seqNo but different digest: received %s, stored %s", preprep.BatchDigest, cert.digest)
		return nil
	}

	cert.prePrepare = preprep
	cert.digest = preprep.BatchDigest

	// Store the request batch if, for whatever reason, we haven't received it from an earlier broadcast
	if _, ok := pbft.validatedBatchStore[preprep.BatchDigest]; !ok && preprep.BatchDigest != "" {
		//digest := hash(preprep.GetTransactionBatch())
		//if digest != preprep.BatchDigest {
		//	logger.Warningf("Pre-prepare and request digest do not match: request %s, digest %s", digest, preprep.BatchDigest)
		//	return nil
		//}
		digest := preprep.BatchDigest
		//pbft.reqBatchStore[digest] = preprep.GetTransactionBatch()
		pbft.validatedBatchStore[digest] = preprep.GetTransactionBatch()
		logger.Debugf("Replica %d storing request batch %s in outstanding request batch store", pbft.id, digest)
		pbft.outstandingReqBatches[digest] = preprep.GetTransactionBatch()
		pbft.persistRequestBatch(digest)
	}

	pbft.softStartTimer(pbft.requestTimeout, fmt.Sprintf("new pre-prepare for request batch %s", preprep.BatchDigest))
	pbft.nullRequestTimer.Stop()
	logger.Debug("receive  pre-prepare first seq is:",preprep.SequenceNumber)
	if pbft.primary(pbft.view) != pbft.id && pbft.prePrepared(preprep.BatchDigest, preprep.View, preprep.SequenceNumber) && !cert.sentPrepare {
		logger.Debugf("Backup %d broadcasting prepare for view=%d/seqNo=%d", pbft.id, preprep.View, preprep.SequenceNumber)
		prep := &Prepare{
			View:           preprep.View,
			SequenceNumber: preprep.SequenceNumber,
			BatchDigest:    preprep.BatchDigest,
			ReplicaId:      pbft.id,
		}
		cert.sentPrepare = true
		//pbft.persistQSet()
		pbft.recvPrepare(prep)
		payload, err := proto.Marshal(prep)
		if err != nil {
			logger.Errorf("ConsensusMessage_PREPARE Marshal Error", err)
			return nil
		}
		consensusMsg := &ConsensusMessage{
			Type:    ConsensusMessage_PREPARE,
			Payload: payload,
		}
		msg := consensusMsgHelper(consensusMsg, pbft.id)
		logger.Debug("after pre-prepare seq is:",prep.SequenceNumber)
		logger.Debug("after pre-prepare seq is:",prep.BatchDigest)

		return pbft.helper.InnerBroadcast(msg)
	}

	return nil
}

func (pbft *pbftProtocal) recvPrepare(prep *Prepare) error {

	logger.Noticef("Replica %d received prepare from replica %d for view=%d/seqNo=%d",
		pbft.id, prep.ReplicaId, prep.View, prep.SequenceNumber)

	if pbft.primary(prep.View) == prep.ReplicaId {
		logger.Warningf("Replica %d received prepare from primary, ignoring", pbft.id)
		return nil
	}

	if !pbft.inWV(prep.View, prep.SequenceNumber) {
		if prep.SequenceNumber != pbft.h && !pbft.skipInProgress {
			logger.Warningf("Replica %d ignoring prepare for view=%d/seqNo=%d: not in-wv, in view %d, low water mark %d", pbft.id, prep.View, prep.SequenceNumber, pbft.view, pbft.h)
		} else {
			// This is perfectly normal
			logger.Debugf("Replica %d ignoring prepare for view=%d/seqNo=%d: not in-wv, in view %d, low water mark %d", pbft.id, prep.View, prep.SequenceNumber, pbft.view, pbft.h)
		}

		return nil
	}

	cert := pbft.getCert(prep.View, prep.SequenceNumber)

	ok := cert.prepare[*prep]

	if ok {
		logger.Warningf("Ignoring duplicate prepare from %d, --------view=%d/seqNo=%d--------", prep.ReplicaId, prep.View, prep.SequenceNumber)
		return nil
	}

	cert.prepare[*prep] = true
	cert.prepareCount++

	logger.Notice("-----primary send primary")
	return pbft.maybeSendCommit(prep.BatchDigest, prep.View, prep.SequenceNumber)
}

//
func (pbft *pbftProtocal) maybeSendCommit(digest string, v uint64, n uint64) error {

	cert := pbft.getCert(v, n)

	if cert == nil {
		logger.Errorf("Replica %d can't get the cert for the view=%d/seqNo=%d", pbft.id, v, n)
		return nil
	}

	if !pbft.prepared(digest, v, n) {
		return nil
	}

	if pbft.primary(pbft.id) == pbft.id {
		return pbft.sendCommit(digest, v, n)
	} else {
		if !cert.sentValidate {
			pbft.validateBatch(cert.prePrepare.TransactionBatch, digest, n, v)
			cert.sentValidate = true
		}

		return nil
	}

}

func (pbft *pbftProtocal) sendCommit(digest string, v uint64, n uint64) error {
	logger.Notice("-----primary2 send primary")

	cert := pbft.getCert(v, n)

	if cert == nil {
		logger.Errorf("Replica %d can't get the cert for the view=%d/seqNo=%d", pbft.id, v, n)
		return nil
	}

	if !cert.sentCommit {
		logger.Noticef("Replica %d broadcasting commit for view=%d/seqNo=%d",
			pbft.id, v, n)
		commit := &Commit{
			View:           v,
			SequenceNumber: n,
			BatchDigest:    digest,
			ReplicaId:      pbft.id,
		}
		cert.sentCommit = true
		pbft.recvCommit(commit)
		payload, err := proto.Marshal(commit)
		if err != nil {
			logger.Errorf("ConsensusMessage_COMMIT Marshal Error", err)
			return nil
		}
		consensusMsg := &ConsensusMessage{
			Type:    ConsensusMessage_COMMIT,
			Payload: payload,
		}
		msg := consensusMsgHelper(consensusMsg, pbft.id)
		return pbft.helper.InnerBroadcast(msg)
	}

	return nil
}

func (pbft *pbftProtocal) recvCommit(commit *Commit) error {

	logger.Noticef("Replica %d received commit from replica %d for view=%d/seqNo=%d",
		pbft.id, commit.ReplicaId, commit.View, commit.SequenceNumber)

	if !pbft.inWV(commit.View, commit.SequenceNumber) {
		if commit.SequenceNumber != pbft.h && !pbft.skipInProgress {
			logger.Warningf("Replica %d ignoring commit for view=%d/seqNo=%d: not in-wv, in view %d, high water mark %d", pbft.id, commit.View, commit.SequenceNumber, pbft.view, pbft.h)
		} else {
			// This is perfectly normal
			logger.Debugf("Replica %d ignoring commit for view=%d/seqNo=%d: not in-wv, in view %d, high water mark %d", pbft.id, commit.View, commit.SequenceNumber, pbft.view, pbft.h)
		}
		return nil
	}

	cert := pbft.getCert(commit.View, commit.SequenceNumber)

	ok := cert.commit[*commit]

	if ok {
		logger.Warningf("Ignoring duplicate commit from %d, --------view=%d/seqNo=%d--------", commit.ReplicaId, commit.View, commit.SequenceNumber)
		return nil
	}

	cert.commit[*commit] = true
	cert.commitCount++

	if pbft.committed(commit.BatchDigest, commit.View, commit.SequenceNumber) && cert.sentExecute == false {
		pbft.stopTimer(commit.SequenceNumber)
		//todo  lastNewViewTimeout
		pbft.lastNewViewTimeout = pbft.newViewTimeout
		delete(pbft.outstandingReqBatches, commit.BatchDigest)
		idx := msgID{v: commit.View, n: commit.SequenceNumber}
		pbft.committedCert[idx] = cert.digest
		pbft.executeOutstanding()
		if commit.SequenceNumber == pbft.viewChangeSeqNo {
			logger.Infof("Replica %d cycling view for seqNo=%d", pbft.id, commit.SequenceNumber)
			pbft.sendViewChange()
		}
	}

	return nil
}

func (pbft *pbftProtocal) executeOutstanding() {

	if pbft.currentExec != nil {
		logger.Debugf("Replica %d not attempting to executeOutstanding bacause it is currently executing %d", pbft.id, pbft.currentExec)
	}

	logger.Debugf("Replica %d attempting to executeOutstanding", pbft.id)

	for idx := range pbft.committedCert {
		if pbft.executeOne(idx) {
			break
		}
	}

	pbft.startTimerIfOutstandingRequests()

}

func (pbft *pbftProtocal) executeOne(idx msgID) bool {

	cert := pbft.certStore[idx]

	if cert == nil || cert.prePrepare == nil {
		logger.Debugf("Replica %d already checkpoint for view=%d/seqNo=%d", pbft.id, idx.v, idx.n)
		return false
	}

	// check if already executed
	if cert.sentExecute == true {
		logger.Debugf("Replica %d already execute for view=%d/seqNo=%d", pbft.id, idx.v, idx.n)
		return false
	}

	if idx.n != pbft.lastExec+1 {
		logger.Debugf("Replica %d hasn't done with last execute %d, seq=%d", pbft.id, pbft.lastExec, idx.n)
		return false
	}

	// skipInProgress == true, then this replica is in viewchange, not reply or execute
	if pbft.skipInProgress {
		logger.Warningf("Replica %d currently picking a starting point to resume, will not execute", pbft.id)
		return false
	}

	digest := cert.digest

	// check if committed
	if !pbft.committed(digest, idx.v, idx.n) {
		return false
	}

	currentExec := idx.n
	pbft.currentExec = &currentExec

	if digest == "" {
		logger.Infof("Replica %d executing null request for view=%d/seqNo=%d", pbft.id, idx.v, idx.n)
		cert.sentExecute = true
		pbft.execDoneSync(idx)
	} else {
		/*
			<<<<<<< HEAD
					logger.Noticef("--------Call execute--------view=%d/seqNo=%d--------", idx.v, idx.n)
					exeBatch := exeBatchHelper(reqBatch, idx.n)
					pbft.helper.Execute(exeBatch)
			=======
		*/
		logger.Infof("--------call execute--------view=%d/seqNo=%d--------", idx.v, idx.n)
		pbft.helper.Execute(idx.n, true, cert.prePrepare.TransactionBatch.Timestamp)
		cert.sentExecute = true
		pbft.execDoneSync(idx)
	}

	return true
}

func (pbft *pbftProtocal) execDoneSync(idx msgID) {

	if pbft.currentExec != nil {
		logger.Debugf("Replica %d finish execution %d, trying next", pbft.id, *pbft.currentExec)
		pbft.lastExec = *pbft.currentExec
		delete(pbft.committedCert, idx)
		if pbft.lastExec%pbft.K == 0 {
			bcInfo := getBlockchainInfo()
			height := bcInfo.Height
			if height == pbft.lastExec {
				logger.Debugf("Call the checkpoint, seqNo=%d, block height=%d", pbft.lastExec, height)
				//time.Sleep(3*time.Millisecond)
				pbft.checkpoint(pbft.lastExec, bcInfo)
			} else {
				// reqBatch call execute but have not done with execute
				logger.Errorf("Fail to call the checkpoint, seqNo=%d, block height=%d", pbft.lastExec, height)
				//pbft.retryCheckpoint(pbft.lastExec)
			}
		}
	} else {
		logger.Warningf("Replica %d had execDoneSync called, flagging ourselves as out of data", pbft.id)
		pbft.skipInProgress = true
	}

	pbft.currentExec = nil
	pbft.executeOutstanding()

}

func (pbft *pbftProtocal) checkpoint(n uint64, info *protos.BlockchainInfo) {

	if n%pbft.K != 0 {
		logger.Errorf("Attempted to checkpoint a sequence number (%d) which is not a multiple of the checkpoint interval (%d)", n, pbft.K)
		return
	}

	id, _ := proto.Marshal(info)
	idAsString := byteToString(id)
	seqNo := n

	logger.Debugf("Replica %d preparing checkpoint for view=%d/seqNo=%d and b64 id of %s",
		pbft.id, pbft.view, seqNo, idAsString)

	chkpt := &Checkpoint{
		SequenceNumber: seqNo,
		ReplicaId:      pbft.id,
		Id:             idAsString,
	}
	pbft.chkpts[seqNo] = idAsString

	pbft.persistCheckpoint(seqNo, id)
	pbft.recvCheckpoint(chkpt)
	payload, err := proto.Marshal(chkpt)
	if err != nil {
		logger.Errorf("ConsensusMessage_CHECKPOINT Marshal Error", err)
		return
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_CHECKPOINT,
		Payload: payload,
	}
	msg := consensusMsgHelper(consensusMsg, pbft.id)
	pbft.helper.InnerBroadcast(msg)
}

func (pbft *pbftProtocal) recvCheckpoint(chkpt *Checkpoint) events.Event {

	logger.Infof("Replica %d received checkpoint from replica %d, seqNo %d, digest %s",
		pbft.id, chkpt.ReplicaId, chkpt.SequenceNumber, chkpt.Id)

	if pbft.weakCheckpointSetOutOfRange(chkpt) {
		return nil
	}

	if !pbft.inW(chkpt.SequenceNumber) {
		if chkpt.SequenceNumber != pbft.h && !pbft.skipInProgress {
			// It is perfectly normal that we receive checkpoints for the watermark we just raised, as we raise it after 2f+1, leaving f replies left
			logger.Warningf("Checkpoint sequence number outside watermarks: seqNo %d, low-mark %d", chkpt.SequenceNumber, pbft.h)
		} else {
			logger.Debugf("Checkpoint sequence number outside watermarks: seqNo %d, low-mark %d", chkpt.SequenceNumber, pbft.h)
		}
		return nil
	}

	cert := pbft.getChkptCert(chkpt.SequenceNumber, chkpt.Id)
	ok := cert.chkpts[*chkpt]

	if ok {
		logger.Warningf("Ignoring duplicate checkpoint from %d, --------seqNo=%d--------", chkpt.ReplicaId, chkpt.SequenceNumber)
		return nil
	}

	cert.chkpts[*chkpt] = true
	cert.chkptCount++
	pbft.checkpointStore[*chkpt] = true

	logger.Debugf("Replica %d found %d matching checkpoints for seqNo %d, digest %s",
		pbft.id, cert.chkptCount, chkpt.SequenceNumber, chkpt.Id)

	if cert.chkptCount == pbft.f+1 {
		// We do have a weak cert
		pbft.witnessCheckpointWeakCert(chkpt)
	}

	if cert.chkptCount < pbft.intersectionQuorum() {
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

	chkptID, ok := pbft.chkpts[chkpt.SequenceNumber]
	if !ok {
		logger.Debugf("Replica %d found checkpoint quorum for seqNo %d, digest %s, but it has not reached this checkpoint itself yet",
			pbft.id, chkpt.SequenceNumber, chkpt.Id)
		if pbft.skipInProgress {
			logSafetyBound := pbft.h + pbft.L/2
			// As an optimization, if we are more than half way out of our log and in state transfer, move our watermarks so we don't lose track of the network
			// if needed, state transfer will restart on completion to a more recent point in time
			if chkpt.SequenceNumber >= logSafetyBound {
				logger.Debugf("Replica %d is in state transfer, but, the network seems to be moving on past %d, moving our watermarks to stay with it", pbft.id, logSafetyBound)
				pbft.moveWatermarks(chkpt.SequenceNumber)
			}
		}
		return nil
	}

	logger.Infof("Replica %d found checkpoint quorum for seqNo %d, digest %s",
		pbft.id, chkpt.SequenceNumber, chkpt.Id)

	if chkptID != chkpt.Id {
		logger.Criticalf("Replica %d generated a checkpoint of %s, but a quorum of the network agrees on %s. This is almost definitely non-deterministic chaincode.",
			pbft.id, chkptID, chkpt.Id)
		pbft.stateTransfer(nil)
	}

	pbft.moveWatermarks(chkpt.SequenceNumber)

	return nil
}

// used in view-change to fetch missing assigned, non-checkpointed requests
func (pbft *pbftProtocal) fetchRequestBatches() (err error) {

	for digest := range pbft.missingReqBatches {
		frb := &FetchRequestBatch{
			BatchDigest: digest,
			ReplicaId:   pbft.id,
		}
		payload, err := proto.Marshal(frb)
		if err != nil {
			logger.Errorf("ConsensusMessage_FRTCH_REQUEST_BATCH Marshal Error", err)
			return nil
		}
		consensusMsg := &ConsensusMessage{
			Type:    ConsensusMessage_FRTCH_REQUEST_BATCH,
			Payload: payload,
		}
		msg := consensusMsgHelper(consensusMsg, pbft.id)
		pbft.helper.InnerBroadcast(msg)
	}

	return
}

func (pbft *pbftProtocal) recvFetchRequestBatch(fr *FetchRequestBatch) (err error) {
	digest := fr.BatchDigest
	if _, ok := pbft.validatedBatchStore[digest]; !ok {
		return nil // we don't have it either
	}

	reqBatch := pbft.validatedBatchStore[digest]
	payload, err := proto.Marshal(reqBatch)
	if err != nil {
		logger.Errorf("ConsensusMessage_RETURN_REQUEST_BATCH Marshal Error", err)
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_RETURN_REQUEST_BATCH,
		Payload: payload,
	}
	msg := consensusMsgHelper(consensusMsg, pbft.id)

	receiver := fr.ReplicaId
	err = pbft.helper.InnerUnicast(msg, receiver)

	return
}

func (pbft *pbftProtocal) recvReturnRequestBatch(reqBatch *TransactionBatch) events.Event {

	digest := hash(reqBatch)
	if _, ok := pbft.missingReqBatches[digest]; !ok {
		return nil // either the wrong digest, or we got it already from someone else
	}
	pbft.validatedBatchStore[digest] = reqBatch
	delete(pbft.missingReqBatches, digest)
	//pbft.persistRequestBatch(digest)

	if len(pbft.missingReqBatches) == 0 {
		//return pbft.processNewView()
		nv, ok := pbft.newViewStore[pbft.view]
		if !ok {
			logger.Debugf("Replica %d ignoring processNewView as it could not find view %d in its newViewStore", pbft.id, pbft.view)
			return nil
		}
		if pbft.activeView {
			logger.Infof("Replica %d ignoring new-view from %d, v:%d: we are active in view %d",
				pbft.id, nv.ReplicaId, nv.View, pbft.view)
			return nil
		}
		return pbft.processReqInNewView(nv)
	}
	return nil

}

func (pbft *pbftProtocal) weakCheckpointSetOutOfRange(chkpt *Checkpoint) bool {
	H := pbft.h + pbft.L

	// Track the last observed checkpoint sequence number if it exceeds our high watermark, keyed by replica to prevent unbounded growth
	if chkpt.SequenceNumber < H {
		// For non-byzantine nodes, the checkpoint sequence number increases monotonically
		delete(pbft.hChkpts, chkpt.ReplicaId)
	} else {
		// We do not track the highest one, as a byzantine node could pick an arbitrarilly high sequence number
		// and even if it recovered to be non-byzantine, we would still believe it to be far ahead
		pbft.hChkpts[chkpt.ReplicaId] = chkpt.SequenceNumber

		// If f+1 other replicas have reported checkpoints that were (at one time) outside our watermarks
		// we need to check to see if we have fallen behind.
		if len(pbft.hChkpts) >= pbft.f+1 {
			chkptSeqNumArray := make([]uint64, len(pbft.hChkpts))
			index := 0
			for replicaID, hChkpt := range pbft.hChkpts {
				chkptSeqNumArray[index] = hChkpt
				index++
				if hChkpt < H {
					delete(pbft.hChkpts, replicaID)
				}
			}
			sort.Sort(sortableUint64Slice(chkptSeqNumArray))

			// If f+1 nodes have issued checkpoints above our high water mark, then
			// we will never record 2f+1 checkpoints for that sequence number, we are out of date
			// (This is because all_replicas - missed - me = 3f+1 - f - 1 = 2f)
			if m := chkptSeqNumArray[len(chkptSeqNumArray)-(pbft.f+1)]; m > H {
				logger.Warningf("Replica %d is out of date, f+1 nodes agree checkpoint with seqNo %d exists but our high water mark is %d", pbft.id, chkpt.SequenceNumber, H)
				pbft.validatedBatchStore = make(map[string]*TransactionBatch) // Discard all our requests, as we will never know which were executed, to be addressed in #394
				pbft.persistDelAllRequestBatches()
				pbft.moveWatermarks(m)
				pbft.outstandingReqBatches = make(map[string]*TransactionBatch)
				pbft.skipInProgress = true
				pbft.invalidateState()
				pbft.stopTimer(chkpt.SequenceNumber)

				// TODO, reprocess the already gathered checkpoints, this will make recovery faster, though it is presently correct

				return true
			}
		}
	}

	return false
}

func (pbft *pbftProtocal) witnessCheckpointWeakCert(chkpt *Checkpoint) {

	// Only ever invoked for the first weak cert, so guaranteed to be f+1
	checkpointMembers := make([]uint64, pbft.f+1)
	i := 0
	for testChkpt := range pbft.checkpointStore {
		if testChkpt.SequenceNumber == chkpt.SequenceNumber && testChkpt.Id == chkpt.Id {
			checkpointMembers[i] = testChkpt.ReplicaId
			logger.Debugf("Replica %d adding replica %d (handle %v) to weak cert", pbft.id, testChkpt.ReplicaId, checkpointMembers[i])
			i++
		}
	}

	snapshotID, err := base64.StdEncoding.DecodeString(chkpt.Id)
	if err != nil {
		err = fmt.Errorf("Replica %d received a weak checkpoint cert which could not be decoded (%s)", pbft.id, chkpt.Id)
		logger.Error(err.Error())
		return
	}

	target := &stateUpdateTarget{
		checkpointMessage: checkpointMessage{
			seqNo: chkpt.SequenceNumber,
			id:    snapshotID,
		},
		replicas: checkpointMembers,
	}
	pbft.updateHighStateTarget(target)

	if pbft.skipInProgress {
		logger.Infof("Replica %d is catching up and witnessed a weak certificate for checkpoint %d, weak cert attested to by %d of %d (%v)",
			pbft.id, chkpt.SequenceNumber, i, pbft.replicaCount, checkpointMembers)
		// The view should not be set to active, this should be handled by the yet unimplemented SUSPECT, see https://github.com/hyperledger/fabric/issues/1120
		pbft.retryStateTransfer(target)
	}
}

func (pbft *pbftProtocal) moveWatermarks(n uint64) {

	// round down n to previous low watermark
	h := n / pbft.K * pbft.K

	for idx, cert := range pbft.certStore {
		if idx.n <= h {
			logger.Debugf("Replica %d cleaning quorum certificate for view=%d/seqNo=%d",
				pbft.id, idx.v, idx.n)
			pbft.persistDelRequestBatch(cert.digest)
			delete(pbft.validatedBatchStore, cert.digest)
			delete(pbft.certStore, idx)
			if pbft.primary(pbft.id) == pbft.id {
				delete(pbft.reqBatchStore, cert.rawDigest)
			}
		}
	}

	for testChkpt := range pbft.checkpointStore {
		if testChkpt.SequenceNumber <= h {
			logger.Debugf("Replica %d cleaning checkpoint message from replica %d, seqNo %d, b64 snapshot id %s",
				pbft.id, testChkpt.ReplicaId, testChkpt.SequenceNumber, testChkpt.Id)
			delete(pbft.checkpointStore, testChkpt)
		}
	}

	for cid := range pbft.chkptCertStore {
		if cid.n <= h {
			logger.Debugf("Replica %d cleaning checkpoint message, seqNo %d, b64 snapshot id %s",
				pbft.id, cid.n, cid.id)
			delete(pbft.chkptCertStore, cid)
		}
	}

	for n := range pbft.pset {
		if n <= h {
			delete(pbft.pset, n)
			//pbft.persistDelPSet(n)
		}
	}

	for idx := range pbft.qset {
		if idx.n <= h {
			delete(pbft.qset, idx)
			//pbft.persistDelQSet(idx)
		}
	}

	for n := range pbft.chkpts {
		if n < h {
			delete(pbft.chkpts, n)
			pbft.persistDelCheckpoint(n)
		}
	}

	pbft.h = h

	logger.Infof("Replica %d updated low watermark to %d",
		pbft.id, pbft.h)

	pbft.resubmitRequestBatches()
}

func (pbft *pbftProtocal) updateHighStateTarget(target *stateUpdateTarget) {
	if pbft.highStateTarget != nil && pbft.highStateTarget.seqNo >= target.seqNo {
		logger.Infof("Replica %d not updating state target to seqNo %d, has target for seqNo %d", pbft.id, target.seqNo, pbft.highStateTarget.seqNo)
		return
	}

	pbft.highStateTarget = target
}

func (pbft *pbftProtocal) stateTransfer(optional *stateUpdateTarget) {

	if !pbft.skipInProgress {
		logger.Debugf("Replica %d is out of sync, pending state transfer", pbft.id)
		pbft.skipInProgress = true
		pbft.invalidateState()
	}

	pbft.retryStateTransfer(optional)
}

func (pbft *pbftProtocal) retryStateTransfer(optional *stateUpdateTarget) {

	if pbft.stateTransferring {
		logger.Debugf("Replica %d is currently mid state transfer, it must wait for this state transfer to complete before initiating a new one", pbft.id)
		return
	}

	target := optional
	if target == nil {
		if pbft.highStateTarget == nil {
			logger.Debugf("Replica %d has no targets to attempt state transfer to, delaying", pbft.id)
			return
		}
		target = pbft.highStateTarget
	}

	pbft.stateTransferring = true

	logger.Debugf("Replica %d is initiating state transfer to seqNo %d", pbft.id, target.seqNo)

	//pbft.batch.pbftManager.Queue() <- stateUpdateEvent // Todo for stateupdate
	//pbft.consumer.skipTo(target.seqNo, target.id, target.replicas)

	pbft.skipTo(target.seqNo, target.id, target.replicas)

}

func (pbft *pbftProtocal) resubmitRequestBatches() {
	if pbft.primary(pbft.view) != pbft.id {
		return
	}

	var submissionOrder []*TransactionBatch

outer:
	for d, reqBatch := range pbft.outstandingReqBatches {
		for _, cert := range pbft.certStore {
			if cert.digest == d {
				logger.Debugf("Replica %d already has certificate for request batch %s - not going to resubmit", pbft.id, d)
				continue outer
			}
		}
		logger.Infof("Replica %d has detected request batch %s must be resubmitted", pbft.id, d)
		submissionOrder = append(submissionOrder, reqBatch)
	}

	if len(submissionOrder) == 0 {
		return
	}

	for _, reqBatch := range submissionOrder {
		// This is a request batch that has not been pre-prepared yet
		// Trigger request batch processing again
		pbft.recvRequestBatch(reqBatch)
	}
}

func (pbft *pbftProtocal) skipTo(seqNo uint64, id []byte, replicas []uint64) {
	info := &protos.BlockchainInfo{}
	err := proto.Unmarshal(id, info)
	if err != nil {
		logger.Error(fmt.Sprintf("Error unmarshaling: %s", err))
		return
	}
	//pbft.UpdateState(&checkpointMessage{seqNo, id}, info, replicas)
	pbft.updateState(seqNo, id, replicas)
}

// UpdateState attempts to synchronize state to a particular target, implicitly calls rollback if needed
func (pbft *pbftProtocal) updateState(seqNo uint64, targetId []byte, replicaId []uint64) {
	//if pbft.valid {
	//	logger.Warning("State transfer is being called for, but the state has not been invalidated")
	//}

	updateStateMsg := stateUpdateHelper(pbft.id, seqNo, targetId, replicaId)
	pbft.helper.UpdateState(updateStateMsg) // TODO: stateUpdateEvent

}

func (pbft *pbftProtocal) updateViewChangeSeqNo() {
	if pbft.viewChangePeriod <= 0 {
		return
	}
	// Ensure the view change always occurs at a checkpoint boundary
	pbft.viewChangeSeqNo = pbft.seqNo + pbft.viewChangePeriod*pbft.K - pbft.seqNo%pbft.K
	logger.Debugf("Replica %d updating view change sequence number to %d", pbft.id, pbft.viewChangeSeqNo)
}
