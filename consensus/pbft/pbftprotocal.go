//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"encoding/base64"
	"fmt"
	"sort"
	"sync"
	"time"
	//"encoding/hex"

	"hyperchain/consensus/events"
	"hyperchain/consensus/helper"
	"hyperchain/core/types"
	//"hyperchain/event"
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
	batchTimer       	events.Timer
	batchTimerActive 	bool
	batchTimeout     	time.Duration
	batchSize        	int
	batchStore       	[]*types.Transaction            //ordered message batch
	helper           	helper.Stack
	batchManager     	events.Manager
	pbftManager      	events.Manager
	muxBatch			sync.Mutex
	muxPbft				sync.Mutex
	reqStore         	*requestStore                   //received messages
	duplicator			map[uint64]*transactionStore

	// PBFT data
	activeView     bool   	// view change happening
	byzantine      bool   	// whether this node is intentionally acting as Byzantine; useful for debugging on the testnet
	f              int		// max. number of faults we can tolerate
	N              int		// max.number of validators in the network
	h              uint64 	// low watermark
	id             uint64 	// replica ID; PBFT `i`
	K              uint64 	// checkpoint period
	logMultiplier  uint64 	// use this value to calculate log size : k*logMultiplier
	L              uint64 	// log size
	lastExec       uint64 	// last request we executed
	seqNo          uint64 	// PBFT "n", strictly monotonic increasing sequence number
	view           uint64 	// current view
	nvInitialSeqNo uint64 	// initial seqNo in a new view
	valid          bool   	// whether we believe the state is up to date

	chkpts map[uint64]string                                 // state checkpoints; map lastExec to global hash
	pset   map[uint64]*ViewChange_PQ                         // state checkpoints; map lastExec to global hash
	qset   map[qidx]*ViewChange_PQ                           // state checkpoints; map lastExec to global hash

	skipInProgress    bool                                   // Set when we have detected a fall behind scenario until we pick a new starting point
	stateTransferring bool                                   // Set when state transfer is executing
	highStateTarget   *stateUpdateTarget                     // Set to the highest weak checkpoint cert we have observed
	hChkpts           map[uint64]uint64                      // highest checkpoint sequence number observed for each replica

	currentExec           *uint64                            // currently executing request
	timerActive           bool                               // is the timer running?
	vcResendTimer         events.Timer                       // timer triggering resend of a view change
	vcResendTimeout       time.Duration                      // timeout before resending view change
	requestTimeout        time.Duration                      // progress timeout for requests
	lastNewViewTimeout    time.Duration                      // last timeout we used during this view change
	outstandingReqBatches map[string]*TransactionBatch       // track whether we are waiting for request batches to execute
	newViewTimeout        time.Duration                      // progress timeout for new views
	newViewTimer          events.Timer                       // track the timeout for each requestBatch
	newViewTimerReason    string                             // what triggered the timer
	nullRequestTimer      events.Timer                       // timeout triggering a null request
	nullRequestTimeout    time.Duration                      // duration for this timeout
	firstRequestTimer     events.Timer		         // firstRequestTimer is set for replicas in case of primary start and shut down immediately
	firstRequestTimeout   time.Duration			 // duration for this timeout

	viewChangePeriod  uint64                                 // period between automatic view changes
	viewChangeSeqNo   uint64                                 // next seqNo to perform view change
	missingReqBatches map[string]bool                        // for all the assigned, non-checkpointed request batches we might be missing during view-change

								 // implementation of PBFT `in`
	certStore       map[msgID]*msgCert                       // track quorum certificates for requests
	checkpointStore map[Checkpoint]bool                      // track checkpoints as set
	committedCert   map[msgID]string                         // track the committed cert to help excute
	chkptCertStore  map[chkptID]*chkptCert                   // track quorum certificates for checkpoints
	newViewStore    map[uint64]*NewView                      // track last new-view we received or sent
	viewChangeStore map[vcidx]*ViewChange                    // track view-change messages

								 // implement the validate transaction batch process
	vid                 	uint64                       // track the validate squence number
	lastVid             	uint64                       // track the last validate batch seqNo
	currentVid          	*uint64                      // track the current validate batch seqNo
	validatedBatchStore 	map[string]*TransactionBatch // track the validated transaction rnnbatch
	cacheValidatedBatch 	map[string]*cacheBatch       // track the cached validated batch
	validateTimer			events.Timer
	validateTimeout			time.Duration

								 // negotiate view
	inNegoView			bool
	negoViewRspStore 	map[uint64]uint64	// track replicaId, viewNo.
	negoViewRspTimer 	events.Timer		// track timeout for N-f nego-view responses
	negoViewRspTimeout	time.Duration		// time limit for N-f nego-view responses

								 // recovery
	inRecovery             bool                              // inRecovery indicate if replica is in proactive recovery process
	recoveryToSeqNo	       *uint64				 // recoveryToSeqNo is the target seqNo expected to recover to
	recoveryRestartTimer   events.Timer                      // recoveryRestartTimer track how long a recovery is finished and fires if needed
	recoveryRestartTimeout time.Duration                     // time limit for recovery process
	rcRspStore             map[uint64]*RecoveryResponse      // rcRspStore store recovery responses from replicas
	rcPQCSenderStore       map[uint64]bool			 // rcPQCSenderStore store those who sent PQC info to self
	recvNewViewInRecovery  bool				 // recvNewViewInRecovery record whether receive new view during recovery

	vcResendLimit	       	int				// vcResendLimit indicates a replica's view change resending upbound.
	vcResendCount	       	int				// vcResendCount represent times of same view change info resend

	// add and del node
	isNewNode			bool						// track if replica is the new node
	localKey			string						// track new node's local key (payload from local)
	addNodeTimer 		events.Timer				// track timeout for new node responses
	addNodeTimeout		time.Duration           	// time limit for new node responses
	inAddingNode		bool						// track if replica is in adding node
	addNodeCertStore	map[string]*addNodeCert		// track the received add node agree message
	inUpdatingN			bool						// track if there exist previous N and new N
	previousN			int							// track the previous N
	previousF			int							// track the previous F
	previousView		uint64						// track the previous View
	mux					sync.Mutex
	keypoint			uint64						// track the key seqNo decided by primary
	newid				uint64						// track the local new id after delete
	delNodeTimer 		events.Timer				// track timeout for del node responses
	delNodeTimeout		time.Duration           	// time limit for del node responses
	inDeletingNode		bool						// track if replica is in adding node
	delNodeCertStore	map[delID]*delNodeCert		// track the received add node agree message
}

type qidx struct {
	d string
	n uint64
}

type cidx struct {
	n uint64
	d string
}

type msgID struct { // our index through certStore
	v uint64
	n uint64
}

type msgCert struct {
	digest       	string
	prePrepare   	*PrePrepare
	sentPrepare  	bool
	prepare      	map[Prepare]bool
	prepareCount 	int
	sentValidate	bool
	validated		bool
	sentCommit   	bool
	commit       	map[Commit]bool
	commitCount  	int
	sentExecute  	bool
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
	vid       uint64
}

type addNodeCert struct {
	addNodes		map[AddNode]bool
	addCount		int
	finishAdd		bool
	update			*UpdateN
	agrees			map[AgreeUpdateN]bool
	updateCount		int
	finishUpdate	bool
}

type delID struct{
	delHash		string
	routerHash	string
}

type delNodeCert struct {
	newId			uint64
	delNodes		map[DelNode]bool
	delCount		int
	finishDel		bool
	update			*UpdateN
	agrees			map[AgreeUpdateN]bool
	updateCount		int
	finishUpdate	bool
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
	pbft.firstRequestTimer = pbftTimerFactory.CreateTimer()
	pbft.N = config.GetInt("pbft.nodes")
	//pbft.f = config.GetInt("general.f")
	pbft.f = (pbft.N-1) / 3

	//if pbft.f*3+1 > pbft.N {
	//	panic(fmt.Sprintf("need at least %d enough replicas to tolerate %d byzantine faults, but only %d replicas configured", pbft.f*3+1, pbft.f, pbft.N))
	//}

	//pbft.K = uint64(config.GetInt("general.K"))
	pbft.K = uint64(10)
	//pbft.logMultiplier = uint64(config.GetInt("general.logmultiplier"))
	pbft.logMultiplier = uint64(4)
	//if pbft.logMultiplier < 2 {
	//	panic("Log multiplier must be greater than or equal to 2")
	//}

	pbft.L = pbft.logMultiplier * pbft.K // log size
	//pbft.viewChangePeriod = uint64(config.GetInt("general.viewchangeperiod"))
	pbft.viewChangePeriod = uint64(0)
	//pbft.byzantine = config.GetBool("general.byzantine")
	pbft.byzantine = false

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

	pbft.firstRequestTimeout, err = time.ParseDuration(config.GetString("timeout.firstrequest"))
	if err != nil {
		logger.Noticef("Replica %d read first request timeout fail", pbft.id)
		pbft.firstRequestTimeout = 30
	}

	pbft.activeView = true

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
	//pbft.reqBatchStore = make(map[string]*TransactionBatch)
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
	pbft.addNodeCertStore = make(map[string]*addNodeCert)
	pbft.delNodeCertStore = make(map[delID]*delNodeCert)

	pbft.isNewNode = false
	pbft.inAddingNode = false
	pbft.inDeletingNode = false
	pbft.inUpdatingN = false
	pbft.duplicator = make(map[uint64]*transactionStore)

	pbft.restoreState()

	pbft.viewChangeSeqNo = ^uint64(0) // infinity
	pbft.updateViewChangeSeqNo()

	pbft.pbftManager.Start()

	// negotiate view
	pbft.inNegoView = true
	pbft.negoViewRspTimeout, err = time.ParseDuration(config.GetString("timeout.negoview"))
	if err != nil {
		panic(fmt.Errorf("Cannot parse negotiate view timeout: %s", err))
	}
	pbft.negoViewRspTimer = pbftTimerFactory.CreateTimer()

	// recovery
	pbft.inRecovery = true
	pbft.recoveryToSeqNo = nil
	pbft.recvNewViewInRecovery = false
	pbft.recoveryRestartTimeout, err = time.ParseDuration(config.GetString("timeout.recovery"))
	if err != nil {
		panic(fmt.Errorf("Cannot parse recovery timeout: %s", err))
	}
	pbft.recoveryRestartTimer = pbftTimerFactory.CreateTimer()

	// vcResendLimit
	pbft.vcResendLimit = config.GetInt("pbft.vcresendlimit")
	logger.Noticef("Replica %d set vcResendLimit %d", pbft.id, pbft.vcResendLimit)
	pbft.vcResendCount = 0

	// batchManager is used to solve batch message, like *Request
	pbft.batchManager = events.NewManagerImpl()
	pbft.batchManager.SetReceiver(pbft)
	etf := events.NewTimerFactoryImpl(pbft.batchManager)
	pbft.batchManager.Start()

	// initialize the batchTimeout
	pbft.batchTimer = etf.CreateTimer()
	pbft.batchSize = config.GetInt("pbft.batchsize")
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
	} else if msg.Type == protos.Message_NEGOTIATE_VIEW {
		return pbft.processNegotiateView()
	}
		logger.Errorf("Unknown recvMsg: %+v", msg)

	return nil
}

func (pbft *pbftProtocal) RecvLocal(msg interface{}) error {

	logger.Debugf("Replica %d received local message", pbft.id)

	go pbft.postPbftEvent(msg)

	return nil
}

func (pbft *pbftProtocal) RemoveCachedBatch(vid uint64) {

	logger.Debugf("Replica %d received local remove message", pbft.id)
	event := removeCache{vid: vid}
	go pbft.postRequestEvent(event)
}

func (pbft *pbftProtocal) ProcessEvent(ee events.Event) events.Event {

	logger.Debugf("Replica %d start solve event", pbft.id)

	switch e := ee.(type) {
	case removeCache:
		vid := e.vid
		ok := pbft.recvRemoveCache(vid)
		if !ok {
			logger.Warningf("Replica %d received local remove cached batch msg, but can not find mapping batch", pbft.id)
		}
		return nil
	case clearDuplicator:
		pbft.duplicator = make(map[uint64]*transactionStore)
	case *types.Transaction:
		tx := e
		return pbft.processTxEvent(tx)
	case viewChangedEvent:
		primary := pbft.primary(pbft.view)
		pbft.persistView(pbft.view)
		pbft.helper.InformPrimary(primary)
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
		logger.Warningf("Replica %d view change timer expired, sending view change: %s", pbft.id, pbft.newViewTimerReason)
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
		if pbft.inNegoView {
			logger.Debugf("Replica %d try to process viewChangeQuorumEvent, but it's in nego-view", pbft.id)
			return nil
		}
		go pbft.postRequestEvent(clearDuplicator{})
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
		logger.Warningf("Replica %d view change resend timer expired before view change quorum was reached, resending", pbft.id)
		pbft.view-- // sending the view change increments this
		return pbft.sendViewChange()
	case *NegotiateView:
		return pbft.recvNegoView(et)
	case *NegotiateViewResponse:
		return pbft.recvNegoViewRsp(et)
	case negoViewRspTimerEvent:
		if !pbft.inNegoView {
			logger.Warningf("Replica %d had its nego-view response timer expire but it's not in nego-view, this is benign but may indicate a bug", pbft.id)
			return nil
		}
		logger.Debugf("Replica %d nego-view response timer expired before N-f was reached, resending", pbft.id)
		pbft.restartNegoView()
	case protos.ValidatedTxs:
		err = pbft.recvValidatedResult(et)
	case negoViewDoneEvent:
		logger.Noticef("######## Replica %d finished negotiating view: %d", pbft.id, pbft.view)
		primary := pbft.primary(pbft.view)
		if primary == pbft.id {
			pbft.sendNullRequest()
		} else {
			pbft.firstRequestTimer.Reset(pbft.firstRequestTimeout, firstRequestTimerEvent{})
		}
		pbft.persistView(pbft.view)
		pbft.helper.InformPrimary(primary)
		pbft.processRequestsDuringNegoView()
		pbft.initRecovery()
		return nil
	case *RecoveryInit:
		return pbft.recvRecovery(et)
	case *RecoveryResponse:
		return pbft.recvRecoveryRsp(et)
	case *RecoveryFetchPQC:
		return pbft.returnRecoveryPQC(et)
	case *RecoveryReturnPQC:
		return pbft.recvRecoveryReturnPQC(et)
	case recoveryDoneEvent:
		logger.Noticef("######## Replica %d finished recovery, height: %d", pbft.id, pbft.lastExec)
		if pbft.recvNewViewInRecovery {
			logger.Noticef("#  Replica %d find itself received NewView during Recovery" +
				", will restart negotiate view", pbft.id)
			pbft.inRecovery = true
			pbft.inNegoView = true
			pbft.restartNegoView()
		}
		if pbft.isNewNode {
			logger.Debug("new node")
			pbft.sendReadyForN()
		}
		pbft.processRequestsDuringRecovery()
		return nil
	case recoveryRestartTimerEvent:
		logger.Noticef("Replica %d recovery restart timer expires", pbft.id)
		pbft.restartRecovery()
		return nil
	case *protos.NewNodeMessage:
		err = pbft.recvLocalNewNode(et)
	case *protos.AddNodeMessage:
		err = pbft.recvLocalAddNode(et)
	case *protos.DelNodeMessage:
		err = pbft.recvLocalDelNode(et)
	case *AddNode:
		err = pbft.recvAgreeAddNode(et)
	case *DelNode:
		err = pbft.recvAgreeDelNode(et)
	case *ReadyForN:
		err = pbft.recvReadyforNforAdd(et)
	case *UpdateN:
		err = pbft.recvUpdateN(et)
	case *AgreeUpdateN:
		err = pbft.recvAgreeUpdateN(et)
	case firstRequestTimerEvent:
		logger.Noticef("Replica %d first request timer expires", pbft.id)
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
		go pbft.postPbftEvent(consensus)
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
	go pbft.postPbftEvent(e)
	return nil
}

// processNullRequest process when a null request come
func (pbft *pbftProtocal) processNullRequest(msg *protos.Message) error {
	if pbft.inNegoView {
		return nil
	}
	if pbft.primary(pbft.view) != pbft.id {
		pbft.firstRequestTimer.Stop()
	}
	pbft.nullReqTimerReset()
	return nil
}

func (pbft *pbftProtocal) processTxEvent(tx *types.Transaction) error {

	primary := pbft.primary(pbft.view)
	if !pbft.activeView || pbft.inNegoView || pbft.inRecovery {
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
	if pbft.activeView {
		pbft.processCachedTransactions()
	} else {
		logger.Critical("peer try to processReqDuringViewChange but view change is not finished")
	}
	return nil
}

func (pbft *pbftProtocal) processCachedTransactions() {
	for pbft.reqStore.outstandingRequests.Len() != 0 {
		temp := pbft.reqStore.outstandingRequests.order.Front().Value
		reqc, ok := interface{}(temp).(requestContainer)
		if !ok {
			logger.Error("type assert error:", temp)
			return
		}
		req := reqc.req
		if req != nil {
			pbft.processTxEvent(req)
		}
		pbft.reqStore.remove(req)
	}
}

func (pbft *pbftProtocal) leaderProcReq(tx *types.Transaction) error {

	logger.Debugf("Batch primary %d queueing new request", pbft.id)

	if pbft.checkDuplicate(tx) {
		_, ok := pbft.duplicator[pbft.vid+1]
		if !ok {
			txStore := newTransactionStore()
			pbft.duplicator[pbft.vid+1] = txStore
		}
		pbft.duplicator[pbft.vid+1].add(tx)
	} else {
		logger.Warningf("Replica %d receive duplicate transaction", pbft.id)
		return nil
	}

	pbft.batchStore = append(pbft.batchStore, tx)

	if !pbft.batchTimerActive {
		pbft.startBatchTimer()
	}

	if len(pbft.batchStore) >= pbft.batchSize {
		return pbft.sendBatch()
	}

	return nil
}

func (pbft *pbftProtocal) checkDuplicate(tx *types.Transaction) (ok bool) {

	ok = true

	for _, txStore := range pbft.duplicator {
		key := string(tx.TransactionHash)
		if txStore.has(key) {
			ok = false
			break
		}
	}

	return
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

	go pbft.postRequestEvent(reqBatch)

	return nil
}

// =============================================================================
// receive methods
// =============================================================================

func (pbft *pbftProtocal) nullRequestHandler() {

	if pbft.inNegoView {
		logger.Debugf("Replica %d try to nullRequestHandler, but it's in nego-view", pbft.id)
		return
	}

	if !pbft.activeView {
		return
	}

	if pbft.primary(pbft.view) != pbft.id {
		// backup expected a null request, but primary never sent one
		logger.Warningf("Replica %d null request timer expired, sending view change", pbft.id)

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
			logger.Error("Unmarshal stringerror, can not unmarshal ConsensusMessage_RETURN_REQUEST_BATCH:", err)
			return nil, err
		}
		return returnRequestBatchEvent(rrb), nil
	case ConsensusMessage_NEGOTIATE_VIEW:
		nv := &NegotiateView{}
		err := proto.Unmarshal(msg.Payload, nv)
		if err != nil {
			logger.Error("Unmarshal error, can not unmarshal ConsensusMessage_NEGOTIATE_VIEW:", err)
			return nil, err
		}
		return nv, nil
	case ConsensusMessage_NEGOTIATE_VIEW_RESPONSE:
		nvr := &NegotiateViewResponse{}
		err := proto.Unmarshal(msg.Payload, nvr)
		if err != nil {
			logger.Error("Unmarshal error, can not unmarshal ConsensusMessage_NEGOTIATE_VIEW_RESPONSE:", err)
			return nil, err
		}
		return nvr, nil
	case ConsensusMessage_RECOVERY_INIT:
		rci := &RecoveryInit{}
		err := proto.Unmarshal(msg.Payload, rci)
		if err != nil {
			logger.Error("Unmarshal error, can not unmarshal ConsensusMessage_RECOVERY_INIT:", err)
			return nil, err
		}
		return rci, nil
	case ConsensusMessage_RECOVERY_RESPONSE:
		rcr := &RecoveryResponse{}
		err := proto.Unmarshal(msg.Payload, rcr)
		if err != nil {
			logger.Error("Unmarshal error, can not unmarshal ConsensusMessage_RECOVERY_RESPONSE:", err)
			return nil, err
		}
		return rcr, nil
	case ConsensusMessage_RECOVERY_FETCH_QPC:
		rcf := &RecoveryFetchPQC{}
		err := proto.Unmarshal(msg.Payload, rcf)
		if err != nil {
			logger.Error("Unmarshal error, can not unmarshal ConsensusMessage_RECOVERY_FETCH_QPC:", err)
		}
		return rcf, err
	case ConsensusMessage_RECOVERY_RETURN_QPC:
		rcr := &RecoveryReturnPQC{}
		err := proto.Unmarshal(msg.Payload, rcr)
		if err != nil {
			logger.Errorf("Unmarshal error, can not unmarshal ConsensusMessage_RECOVERY_RETURN_QPC:", err)
		}
		return rcr, err
	case ConsensusMessage_ADD_NODE:
		an := &AddNode{}
		err := proto.Unmarshal(msg.Payload, an)
		if err != nil {
			logger.Errorf("Unmarshal error, can not unmarshal ConsensusMessage_ADD_NODE:", err)
		}
		return an, err
	case ConsensusMessage_DEL_NODE:
		dn := &DelNode{}
		err := proto.Unmarshal(msg.Payload, dn)
		if err != nil {
			logger.Errorf("Unmarshal error, can not unmarshal ConsensusMessage_DEL_NODE:", err)
		}
		return dn, err
	case ConsensusMessage_READY_FOR_N:
		rfn := &ReadyForN{}
		err := proto.Unmarshal(msg.Payload, rfn)
		if err != nil {
			logger.Errorf("Unmarshal error, can not unmarshal ConsensusMessage_READY_FOR_N:", err)
		}
		return rfn, err
	case ConsensusMessage_UPDATE_N:
		un := &UpdateN{}
		err := proto.Unmarshal(msg.Payload, un)
		if err != nil {
			logger.Errorf("Unmarshal error, can not unmarshal ConsensusMessage_UPDATE_N:", err)
		}
		return un, err
	case ConsensusMessage_AGREE_UPDATE_N:
		aun := &AgreeUpdateN{}
		err := proto.Unmarshal(msg.Payload, aun)
		if err != nil {
			logger.Errorf("Unmarshal error, can not unmarshal ConsensusMessage_AGREE_UPDATE_N:", err)
		}
		return aun, err
	default:
		return nil, fmt.Errorf("Invalid message: %v", msg)
	}

}

func (pbft *pbftProtocal) recvStateUpdatedEvent(et *stateUpdatedEvent) error {

	if pbft.inNegoView {
		logger.Debugf("Replica %d try to recvStateUpdatedEvent, but it's in nego-view", pbft.id)
		return nil
	}

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
	pbft.vid = et.seqNo
	pbft.lastVid = et.seqNo
	bcInfo := getCurrentBlockInfo()
	id, _ := proto.Marshal(bcInfo)
	pbft.persistCheckpoint(et.seqNo, id)
	pbft.moveWatermarks(pbft.lastExec) // The watermark movement handles moving this to a checkpoint boundary
	pbft.skipInProgress = false
	pbft.validateState()
	pbft.executeAfterStateUpdate()

	if pbft.inRecovery {
		if pbft.recoveryToSeqNo == nil {
			logger.Errorf("Replica %d in recovery recvStateUpdatedEvent but " +
				"its recoveryToSeqNo is nil")
			return nil
		}
		if pbft.lastExec == *pbft.recoveryToSeqNo {
			// This is a somewhat subtle situation, we are behind by checkpoint, but others are just on chkpt.
			// Hence, no need to fetch preprepare, prepare, commit
			pbft.inRecovery = false
			pbft.recoveryToSeqNo = nil
			pbft.recoveryRestartTimer.Stop()
			go pbft.postPbftEvent(recoveryDoneEvent{})
			return nil
		}

		pbft.recoveryRestartTimer.Reset(pbft.recoveryRestartTimeout, recoveryRestartTimerEvent{})
		if pbft.highStateTarget == nil {
			logger.Errorf("Try to fetch QPC, but highStateTarget is nil")
			return nil
		}
		peers := pbft.highStateTarget.replicas
		pbft.fetchRecoveryPQC(peers)
		return nil
	}

	return nil
}

func (pbft *pbftProtocal) recvRequestBatch(reqBatch *TransactionBatch) error {

	if pbft.inNegoView {
		logger.Debugf("Replica %d try to recvRequestBatch, but it's in nego-view", pbft.id)
		return nil
	}

	digest := hash(reqBatch)
	logger.Debugf("Replica %d received request batch %s", pbft.id, digest)

	if pbft.activeView {
		pbft.softStartTimer(pbft.requestTimeout, fmt.Sprintf("new request batch %s", digest))
	}
	if pbft.primary(pbft.view) == pbft.id && pbft.activeView {
		pbft.validateBatch(reqBatch, 0, 0)
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

func (pbft *pbftProtocal) validateBatch(txBatch *TransactionBatch, vid uint64, view uint64) {

	primary := pbft.primary(pbft.view)
	if primary == pbft.id {
		logger.Debugf("Primary %d try to validate batch %s", pbft.id, hash(txBatch))

		n := pbft.vid + 1

		pbft.vid = n
		pbft.helper.ValidateBatch(txBatch.Batch, txBatch.Timestamp, n, pbft.view, true)
	} else {
		logger.Debugf("Replica %d try to validate batch", pbft.id)

		if !pbft.inWV(pbft.view, vid) {
			logger.Debugf("Replica %d not validating for transaction batch because it is out of sequence numbers", pbft.id)
			return
		}
		pbft.helper.ValidateBatch(txBatch.Batch, txBatch.Timestamp, vid, view, false)
	}

}

func (pbft *pbftProtocal) trySendPrePrepare() {


	if pbft.currentVid != nil {
		logger.Debugf("Replica %d not attempting to send pre-prepare bacause it is currently send %d, retry.", pbft.id, pbft.currentVid)
		return
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

	if len(cache.batch.Batch) == 0 {
		logger.Warningf("Replica %d is primary, receives validated result %s that is empty", pbft.id, digest)
		pbft.lastVid = *pbft.currentVid
		pbft.currentVid = nil
		delete(pbft.cacheValidatedBatch, digest)
		delete(pbft.validatedBatchStore, digest)
		delete(pbft.outstandingReqBatches, digest)
		pbft.stopTimer()
		return true
	}

	pbft.nullRequestTimer.Stop()
	pbft.sendPrePrepare(cache.batch, digest)

	return true
}

func (pbft *pbftProtocal) sendPrePrepare(reqBatch *TransactionBatch, digest string) {

	logger.Debugf("Replica %d is primary, issuing pre-prepare for request batch %s", pbft.id, digest)

	n := pbft.seqNo + 1

	for _, cert := range pbft.certStore { // check for other PRE-PREPARE for same digest, but different seqNo
		if p := cert.prePrepare; p != nil {
			if p.View == pbft.view && p.SequenceNumber != n && p.BatchDigest == digest && digest != "" {
				// This will happen if primary receive same digest result of txs
				// It may result in DDos attack
				logger.Warningf("Other pre-prepare found with same digest but different seqNo: %d instead of %d", p.SequenceNumber, n)
				pbft.lastVid = *pbft.currentVid
				pbft.currentVid = nil
				delete(pbft.cacheValidatedBatch, digest)
				delete(pbft.validatedBatchStore, digest)
				delete(pbft.outstandingReqBatches, digest)
				pbft.stopTimer()
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
	cert.sentValidate = true
	cert.validated = true
	delete(pbft.cacheValidatedBatch, digest)
	//pbft.persistQSet()
	payload, err := proto.Marshal(preprep)
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

	if pbft.inNegoView {
		logger.Debugf("Replica %d try recvPrePrepare, but it's in nego-view", pbft.id)
		return nil
	}

	if pbft.primary(pbft.view) != pbft.id {
		pbft.firstRequestTimer.Stop()
	}

	//logger.Notice("receive  pre-prepare first seq is:",preprep.SequenceNumber)

	logger.Debugf("Replica %d received pre-prepare from replica %d for view=%d/seqNo=%d, digest=%s ",
		pbft.id, preprep.ReplicaId, preprep.View, preprep.SequenceNumber, preprep.BatchDigest)

	if !pbft.activeView {
		logger.Debugf("Replica %d ignoring pre-prepare as we sre in view change", pbft.id)
		return nil
	}

	if pbft.primary(pbft.view) != preprep.ReplicaId {
		logger.Warningf("Pre-prepare from other than primary: got %d, should be %d",
			preprep.ReplicaId, pbft.primary(pbft.view))
		return nil
	}

	pbft.nullRequestTimer.Stop()

	if !pbft.inWV(preprep.View, preprep.SequenceNumber) {
		if preprep.SequenceNumber != pbft.h && !pbft.skipInProgress {
			logger.Warningf("Replica %d pre-prepare view different, or sequence number outside watermarks: preprep.View %d, expected.View %d, seqNo %d, low-mark %d", pbft.id, preprep.View, pbft.view, preprep.SequenceNumber, pbft.h)
		} else {
			// This is perfectly normal
			logger.Debugf("Replica %d pre-prepare view different, or sequence number outside watermarks: preprep.View %d, expected.View %d, seqNo %d, low-mark %d", pbft.id, preprep.View, pbft.view, preprep.SequenceNumber, pbft.h)
		}

		return nil
	}

	// add this for recovery, avoid saving batch with seqno that already executed
	if pbft.currentExec != nil {
		if preprep.SequenceNumber <= *pbft.currentExec {
			logger.Debugf("Replica %d reject out-of-date pre-prepare for seqNo=%d/view=%d", pbft.id, preprep.SequenceNumber, preprep.View)
			return nil
		}
	} else {
		if preprep.SequenceNumber <= pbft.lastExec {
			logger.Debugf("Replica %d reject out-of-date pre-prepare for seqNo=%d/view=%d", pbft.id, preprep.SequenceNumber, preprep.View)
			return nil
		}
	}

	cert := pbft.getCert(preprep.View, preprep.SequenceNumber)

	if cert.digest != "" && cert.digest != preprep.BatchDigest {
		logger.Warningf("Pre-prepare found for same view/seqNo but different digest: received %s, stored %s",
			preprep.BatchDigest, cert.digest)
		return nil
	}

	cert.prePrepare = preprep
	cert.digest = preprep.BatchDigest

	// Store the request batch if, for whatever reason, we haven't received it from an earlier broadcast
	if _, ok := pbft.validatedBatchStore[preprep.BatchDigest]; !ok && preprep.BatchDigest != "" {
		digest := preprep.BatchDigest
		pbft.validatedBatchStore[digest] = preprep.GetTransactionBatch()
		logger.Debugf("Replica %d storing request batch %s in outstanding request batch store", pbft.id, digest)
		pbft.outstandingReqBatches[digest] = preprep.GetTransactionBatch()
		pbft.persistRequestBatch(digest)
	}

	if !pbft.skipInProgress && !pbft.inRecovery {
		pbft.softStartTimer(pbft.requestTimeout, fmt.Sprintf("new pre-prepare for request batch %s", preprep.BatchDigest))
	}

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

	logger.Debugf("Replica %d received prepare from replica %d for view=%d/seqNo=%d",
		pbft.id, prep.ReplicaId, prep.View, prep.SequenceNumber)

	if pbft.inNegoView {
		logger.Debugf("Replica %d try to recvPrepare, but it's in nego-view", pbft.id)
		return nil
	}

	if pbft.primary(prep.View) == prep.ReplicaId && !pbft.inRecovery{
		logger.Warningf("Replica %d received prepare from primary, ignoring", pbft.id)
		return nil
	}

	if !pbft.inWV(prep.View, prep.SequenceNumber) {
		if prep.SequenceNumber != pbft.h && !pbft.skipInProgress {
			logger.Warningf("Replica %d ignoring prepare for view=%d/seqNo=%d: not in-wv, in view %d, low water mark %d",
				pbft.id, prep.View, prep.SequenceNumber, pbft.view, pbft.h)
		} else {
			// This is perfectly normal
			logger.Debugf("Replica %d ignoring prepare for view=%d/seqNo=%d: not in-wv, in view %d, low water mark %d",
				pbft.id, prep.View, prep.SequenceNumber, pbft.view, pbft.h)
		}

		return nil
	}

	cert := pbft.getCert(prep.View, prep.SequenceNumber)

	ok := cert.prepare[*prep]

	if ok {
		logger.Warningf("Ignoring duplicate prepare from %d, --------view=%d/seqNo=%d--------",
			prep.ReplicaId, prep.View, prep.SequenceNumber)
		return nil
	}

	cert.prepare[*prep] = true
	cert.prepareCount++

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

	if pbft.skipInProgress {
		logger.Debugf("Replica %d do not try to validate batch because it's in state update", pbft.id)
		return nil
	}

	if pbft.primary(pbft.view) == pbft.id {

		return pbft.sendCommit(digest, v, n)
	} else {
		if !cert.sentValidate {
			pbft.validateBatch(cert.prePrepare.TransactionBatch, n, v)
			cert.sentValidate = true
		}

		return nil
	}

}

func (pbft *pbftProtocal) sendCommit(digest string, v uint64, n uint64) error {


	cert := pbft.getCert(v, n)

	if cert == nil {
		logger.Errorf("Replica %d can't get the cert for the view=%d/seqNo=%d", pbft.id, v, n)
		return nil
	}

	if !cert.sentCommit {
		logger.Debugf("Replica %d broadcasting commit for view=%d/seqNo=%d",
			pbft.id, v, n)
		commit := &Commit{
			View:           v,
			SequenceNumber: n,
			BatchDigest:    digest,
			ReplicaId:      pbft.id,
		}
		cert.sentCommit = true

		payload, err := proto.Marshal(commit)
		if err != nil {
			logger.Errorf("ConsensusMessage_COMMIT Marshal Error", err)
			return nil
		}
		consensusMsg := &ConsensusMessage{
			Type:    ConsensusMessage_COMMIT,
			Payload: payload,
		}
		go pbft.postPbftEvent(consensusMsg)
		msg := consensusMsgHelper(consensusMsg, pbft.id)
		return pbft.helper.InnerBroadcast(msg)
	}

	return nil
}

func (pbft *pbftProtocal) recvCommit(commit *Commit) error {

	logger.Debugf("Replica %d received commit from replica %d for view=%d/seqNo=%d",
		pbft.id, commit.ReplicaId, commit.View, commit.SequenceNumber)

	if pbft.inNegoView {
		logger.Debugf("Replica %d try to recvCommit, but it's in nego-view", pbft.id)
		return nil
	}

	if !pbft.inWV(commit.View, commit.SequenceNumber) {
		if commit.SequenceNumber != pbft.h && !pbft.skipInProgress {
			logger.Warningf("Replica %d ignoring commit from replica %d for view=%d/seqNo=%d: not in-wv, in view %d, low water mark %d", pbft.id, commit.ReplicaId, commit.View, commit.SequenceNumber, pbft.view, pbft.h)
		} else {
			// This is perfectly normal
			logger.Debugf("Replica %d ignoring commit for view=%d/seqNo=%d: not in-wv, in view %d, low water mark %d", pbft.id, commit.View, commit.SequenceNumber, pbft.view, pbft.h)
		}
		return nil
	}

	cert := pbft.getCert(commit.View, commit.SequenceNumber)

	ok := cert.commit[*commit]

	if ok {
		logger.Warningf("Ignoring duplicate commit from %d, --------view=%d/seqNo=%d--------",
			commit.ReplicaId, commit.View, commit.SequenceNumber)
		return nil
	}

	cert.commit[*commit] = true
	cert.commitCount++

	if pbft.committed(commit.BatchDigest, commit.View, commit.SequenceNumber) {
		pbft.stopTimer()
		if !cert.sentExecute && cert.validated {
			pbft.lastNewViewTimeout = pbft.newViewTimeout
			delete(pbft.outstandingReqBatches, commit.BatchDigest)
			idx := msgID{v: commit.View, n: commit.SequenceNumber}
			pbft.committedCert[idx] = cert.digest
			pbft.executeOutstanding()
			if commit.SequenceNumber == pbft.viewChangeSeqNo {
				logger.Warningf("Replica %d cycling view for seqNo=%d", pbft.id, commit.SequenceNumber)
				pbft.sendViewChange()
			}
		} else {
			logger.Debugf("Replica %d committed for seqNo: %d, but sentExecute: %v, validated: %v", pbft.id, commit.SequenceNumber,cert.sentExecute, cert.validated)
		}
	}

	return nil
}

func (pbft *pbftProtocal) executeAfterStateUpdate() {

	logger.Debugf("Replica %d try to execute after state update", pbft.id)

	for idx, cert := range pbft.certStore {
		if idx.n > pbft.seqNo && pbft.prepared(cert.digest, idx.v, idx.n) && !cert.validated {
			logger.Debugf("Replica %d try to vaidate batch %s", pbft.id, cert.digest)
			pbft.validateBatch(cert.prePrepare.TransactionBatch, idx.n, idx.v)
			cert.sentValidate = true
		}
	}

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
		logger.Noticef("--------Replica %d Call execute, view=%d/seqNo=%d--------", pbft.id, idx.v, idx.n)
		var isPrimary bool
		if pbft.primary(pbft.view) == pbft.id {
			isPrimary = true
		} else {
			isPrimary = false
		}
		pbft.helper.Execute(idx.n, digest, true, isPrimary, cert.prePrepare.TransactionBatch.Timestamp)
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
		if pbft.inRecovery {
			if pbft.recoveryToSeqNo == nil {
				logger.Errorf("Replica %d in recovery execDoneSync but " +
					"its recoveryToSeqNo is nil")
				return
			}
			if pbft.lastExec == *pbft.recoveryToSeqNo {
				pbft.inRecovery = false
				pbft.recoveryToSeqNo = nil
				pbft.recoveryRestartTimer.Stop()
				go pbft.postPbftEvent(recoveryDoneEvent{})
			}
		}
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
	// optimization: if we are in view changing waiting for executing to target seqNo,
	// one-time processNewView() is enough. No need to processNewView() every time in execDoneSync()
	if !pbft.activeView && pbft.lastExec == pbft.nvInitialSeqNo {
		pbft.processNewView()
	}

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

	logger.Infof("Replica %d preparing checkpoint for view=%d/seqNo=%d and b64 id of %s",
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

	logger.Debugf("Replica %d received checkpoint from replica %d, seqNo %d, digest %s",
		pbft.id, chkpt.ReplicaId, chkpt.SequenceNumber, chkpt.Id)

	if pbft.inNegoView {
		logger.Debugf("Replica %d try to recvCheckpoint, but it's in nego-view", pbft.id)
		return nil
	}

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

	if pbft.inNegoView {
		logger.Debugf("Replica %d try to recvFetchRequestBatch, but it's in nego-view", pbft.id)
		return nil
	}

	if pbft.inRecovery {
		logger.Noticef("Replica %d try to recvFetchRequestBatch, but it's in recovery", pbft.id)
		return nil
	}

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

	if pbft.inNegoView {
		logger.Debugf("Replica %d try to recvReturnRequestBatch, but it's in nego-view", pbft.id)
		return nil
	}
	if pbft.inRecovery {
		logger.Noticef("Replica %d try to recvReturnRequestBatch, but it's in recovery", pbft.id)
		return nil
	}

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
				pbft.stopTimer()

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
			pbft.id, chkpt.SequenceNumber, i, pbft.N, checkpointMembers)
		// The view should not be set to active, this should be handled by the yet unimplemented SUSPECT, see https://github.com/hyperledger/fabric/issues/1120
		pbft.retryStateTransfer(target)
	}
}

func (pbft *pbftProtocal) moveWatermarks(n uint64) {

	// round down n to previous low watermark
	h := n / pbft.K * pbft.K

	if pbft.h > n {
		logger.Critical("Replica %d movewatermark but pbft.h>n", pbft.id)
		return
	}

	if pbft.inUpdatingN && pbft.keypoint <= h {
		pbft.inUpdatingN = false
		logger.Noticef("Replica %d finish updating N after adding", pbft.id)
	}

	for idx, cert := range pbft.certStore {
		if idx.n <= h {
			logger.Debugf("Replica %d cleaning quorum certificate for view=%d/seqNo=%d",
				pbft.id, idx.v, idx.n)
			pbft.persistDelRequestBatch(cert.digest)
			delete(pbft.validatedBatchStore, cert.digest)
			delete(pbft.outstandingReqBatches, cert.digest)
			delete(pbft.certStore, idx)
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
		logger.Infof("Replica %d not updating state target to seqNo %d, has target for seqNo %d",
			pbft.id, target.seqNo, pbft.highStateTarget.seqNo)
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

	logger.Infof("Replica %d is initiating state transfer to seqNo %d", pbft.id, target.seqNo)

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
	logger.Debug("seqNo: ", seqNo, "id: ", id, "replicas: ", replicas)
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

func (pbft *pbftProtocal) processNegotiateView() error {
	if !pbft.inNegoView {
		logger.Critical("Replica %d try to negotiateView, but it's not inNegoView. This indicates a bug")
		return nil
	}

	logger.Debugf("Replica %d now negotiate view...", pbft.id)

	pbft.negoViewRspTimer.Reset(pbft.negoViewRspTimeout, negoViewRspTimerEvent{})
	pbft.negoViewRspStore = make(map[uint64]uint64)

	// broadcast the negotiate message to other replica
	negoViewMsg := &NegotiateView{
		ReplicaId:pbft.id,
	}
	payload, err := proto.Marshal(negoViewMsg)
	if err!=nil {
		logger.Errorf("Marshal NegotiateView Error!")
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:		ConsensusMessage_NEGOTIATE_VIEW,
		Payload:	payload,
	}
	msg := consensusMsgHelper(consensusMsg, pbft.id)
	pbft.helper.InnerBroadcast(msg)
	logger.Debugf("Replica %d broadcast negociate view message", pbft.id)

	// post the negotiate message event to myself
	nvr := &NegotiateViewResponse{
		ReplicaId:	pbft.id,
		View:		pbft.view,
	}
	consensusPayload, err := proto.Marshal(nvr)
	if err!=nil {
		logger.Errorf("Marshal NegotiateViewResponse Error!")
		return nil
	}
	responseMsg := &ConsensusMessage{
		Type:		ConsensusMessage_NEGOTIATE_VIEW_RESPONSE,
		Payload:	consensusPayload,
	}
	go pbft.postPbftEvent(responseMsg)

	return nil
}

func (pbft *pbftProtocal) recvNegoView(nv *NegotiateView) events.Event {
	if !pbft.activeView {
		return nil
	}

	sender := nv.ReplicaId
	logger.Debugf("Replica %d receive negotiate view from %d", pbft.id, sender)
	negoViewRsp := &NegotiateViewResponse{
		ReplicaId:pbft.id,
		View:pbft.view,
	}
	payload, err := proto.Marshal(negoViewRsp)
	if err!=nil {
		logger.Errorf("Marshal NegotiateViewResponse Error!")
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type: ConsensusMessage_NEGOTIATE_VIEW_RESPONSE,
		Payload: payload,
	}
	msg := consensusMsgHelper(consensusMsg, pbft.id)
	pbft.helper.InnerUnicast(msg, sender)
	return nil
}

func (pbft *pbftProtocal) recvNegoViewRsp(nvr *NegotiateViewResponse) events.Event {
	if !pbft.inNegoView {
		logger.Debugf("Replica %d already finished nego-view, ignore incoming nego-view response", pbft.id)
		return nil
	}

	rspId, rspView := nvr.ReplicaId, nvr.View
	if _, ok := pbft.negoViewRspStore[rspId]; ok {
		logger.Warningf("Already recv view number from %d, ignore it", rspId)
		return nil
	}

	logger.Debugf("Replica %d receive nego-view response from %d, view: %d", pbft.id, rspId, rspView)

	pbft.negoViewRspStore[rspId] = rspView

	if len(pbft.negoViewRspStore) > pbft.N-pbft.f {
		// Reason for not using '≥ pbft.N-pbft.f': if self is wrong, then we are impossible to find 2f+1 same view
		// can we find same view from 2f+1 peers?
		viewCount := make(map[uint64]uint64)
		var theView uint64
		canFind := false
		for _, view := range pbft.negoViewRspStore {
			if _, ok := viewCount[view]; ok {
				viewCount[view]++
			} else {
				viewCount[view] = uint64(1)
			}
			if viewCount[view] >= uint64(2*pbft.f+1) {
				// yes we find the view
				theView = view
				canFind = true
				break
			}
		}
		if canFind {
			pbft.negoViewRspTimer.Stop()
			pbft.view = theView
			pbft.inNegoView = false
			if !pbft.activeView {
				pbft.activeView = true
			}
			return negoViewDoneEvent{}
		} else {
			pbft.negoViewRspTimer.Reset(pbft.negoViewRspTimeout, negoViewRspTimerEvent{})
			logger.Warningf("pbft recv at least N-f nego-view responses, but cannot find same view from 2f+1.")
		}
	}
	return nil
}

func (pbft *pbftProtocal) restartNegoView() {
	logger.Debugf("Replica %d restart negotiate view", pbft.id)
	pbft.processNegotiateView()
}

func (pbft *pbftProtocal) processRequestsDuringNegoView() {
	if !pbft.inNegoView {
		pbft.processCachedTransactions()
	} else {
		logger.Critical("Replica %d try to processRequestsDuringNegoView but nego-view is not finished", pbft.id)
	}
}

func (pbft *pbftProtocal) processRequestsDuringRecovery() {
	if !pbft.inRecovery {
		pbft.processCachedTransactions()
	} else {
		logger.Critical("Replica %d try to processRequestsDuringRecovery but recovery is not finished", pbft.id)
	}
}


// =============================================================================
// receive local message methods
// =============================================================================
func (pbft *pbftProtocal) recvValidatedResult(result protos.ValidatedTxs) error {

	primary := pbft.primary(pbft.view)
	if primary == pbft.id {
		logger.Debugf("Primary %d received validated batch for sqeNo=%d, batch is: %s", pbft.id, result.SeqNo, result.Hash)

		batch := &TransactionBatch{
			Batch:     result.Transactions,
			Timestamp: result.Timestamp,
		}
		digest := result.Hash
		pbft.validatedBatchStore[digest] = batch
		pbft.outstandingReqBatches[digest] = batch
		cache := &cacheBatch{
			batch:     batch,
			vid:       result.SeqNo,
		}
		pbft.cacheValidatedBatch[digest] = cache
		//if pbft.seqNo-pbft.lastExec > 20 {
		//	time.Sleep(20 * time.Millisecond)
		//}
		pbft.trySendPrePrepare()
	} else {
		logger.Debugf("Replica %d recived validated batch for sqeNo=%d, batch is: %s", pbft.id, result.SeqNo, result.Hash)

		if !pbft.inWV(result.View, result.SeqNo) {
			logger.Debugf("Replica %d receives validated result %s that is out of sequence numbers", pbft.id, result.Hash)
			return nil
		}

		cert := pbft.getCert(result.View, result.SeqNo)
		cert.validated = true

		digest := result.Hash
		if digest == cert.digest {
			pbft.sendCommit(digest, result.View, result.SeqNo)
		} else {
			logger.Warningf("Relica %d cannot agree with the validate result sent from primary", pbft.id)
			pbft.sendViewChange()
		}
	}

	return nil
}

func (pbft *pbftProtocal) recvRemoveCache(vid uint64) bool {

	_, ok := pbft.duplicator[vid]
	if ok {
		delete(pbft.duplicator, vid)
	}

	return ok
}