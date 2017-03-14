//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"github.com/golang/protobuf/proto"
	"hyperchain/consensus/events"
	"encoding/base64"
	"hyperchain/consensus/helper/persist"
)
/**
	This file contains recovery related issues
 */

type recoveryManager struct {
	recoveryToSeqNo	       *uint64				 // recoveryToSeqNo is the target seqNo expected to recover to
	rcRspStore             map[uint64]*RecoveryResponse      // rcRspStore store recovery responses from replicas
	rcPQCSenderStore       map[uint64]bool			 // rcPQCSenderStore store those who sent PQC info to self
	recvNewViewInRecovery  bool				 // recvNewViewInRecovery record whether receive new view during recovery
	negoViewRspStore       map[uint64]*NegotiateViewResponse // track replicaId, viewNo.
}

type blkIdx struct {
	height uint64
	hash   string
}

func newRecoveryMgr() *recoveryManager  {
	rm := &recoveryManager{}

	rm.recoveryToSeqNo = nil
	rm.recvNewViewInRecovery = false

	return rm
}

//dispatchRecoveryMsg dispatch recovery service messages from other peers.
func (pbft *pbftImpl) dispatchRecoveryMsg(e events.Event) events.Event {
	switch et := e.(type) {
	case *NegotiateView:
		return pbft.recvNegoView(et)
	case *NegotiateViewResponse:
		return pbft.recvNegoViewRsp(et)
	case *RecoveryInit:
		return pbft.recvRecovery(et)
	case *RecoveryResponse:
		return pbft.recvRecoveryRsp(et)
	case *RecoveryFetchPQC:
		return pbft.returnRecoveryPQC(et)
	case *RecoveryReturnPQC:
		return pbft.recvRecoveryReturnPQC(et)
	}
	return nil
}

// procativeRecovery broadcast a procative recovery message to ask others for recent blocks info
func (pbft *pbftImpl) initRecovery() events.Event {

	logger.Debugf("Replica %d now initRecovery", pbft.id)

	pbft.recoveryMgr.recoveryToSeqNo = nil
	// update watermarks
	height := persist.GetHeightofChain(pbft.namespace)
	pbft.moveWatermarks(height)

	pbft.recoveryMgr.rcRspStore = make(map[uint64]*RecoveryResponse)

	recoveryMsg := &RecoveryInit{
		ReplicaId: pbft.id,
	}
	payload, err := proto.Marshal(recoveryMsg)
	if err != nil {
		logger.Errorf("Marshal recovery init Error!")
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_RECOVERY_INIT,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, pbft.id)
	pbft.helper.InnerBroadcast(msg)

	event := &LocalEvent{
		Service:   RECOVERY_SERVICE,
		EventType: RECOVERY_RESTART_TIMER_EVENT,
	}

	af := func(){
		pbft.pbftEventQueue.Push(event)
	}

	pbft.pbftTimerMgr.startTimer(RECOVERY_RESTART_TIMER, af)

	chkpts := make(map[uint64]string)
	for n, d := range pbft.storeMgr.chkpts {
		chkpts[n] = d
	}
	rc := &RecoveryResponse{
		ReplicaId: pbft.id,
		Chkpts:    chkpts,
	}
	pbft.recvRecoveryRsp(rc)
	return nil
}


// recvRcry process incoming proactive recovery message
func (pbft *pbftImpl) recvRecovery(recoveryInit *RecoveryInit) events.Event {

	logger.Debugf("Replica %d now recvRecovery from replica %d", pbft.id, recoveryInit.ReplicaId)

	if pbft.status.getState(&pbft.status.skipInProgress) {
		logger.Debugf("Replica %d recvRecovery, but it's in state transfer and ignores it.", pbft.id)
		return nil
	}
	chkpts := make(map[uint64]string)
	for n, d := range pbft.storeMgr.chkpts {
		chkpts[n] = d
	}

	height, curHash := persist.GetBlockHeightAndHash(pbft.namespace)

	rc := &RecoveryResponse{
		ReplicaId:     pbft.id,
		Chkpts:        chkpts,
		BlockHeight:   height,
		LastBlockHash: curHash,
	}

	rcMsg, err := proto.Marshal(rc)
	if err != nil {
		logger.Errorf("recovery response marshal error")
		return nil
	}

	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_RECOVERY_RESPONSE,
		Payload: rcMsg,
	}
	dest := recoveryInit.ReplicaId
	msg := cMsgToPbMsg(consensusMsg, dest)
	pbft.helper.InnerUnicast(msg, dest)

	return nil
}

// recvRcryRsp process other replicas' feedback as with initRecovery
func (pbft *pbftImpl) recvRecoveryRsp(rsp *RecoveryResponse) events.Event {

	logger.Debugf("Replica %d now recvRecoveryRsp from replica %d", pbft.id, rsp.ReplicaId)

	if !pbft.status.getState(&pbft.status.inRecovery) {
		logger.Debugf("Replica %d finished recovery, ignore recovery response", pbft.id)
		return nil
	}
	from := rsp.ReplicaId
	if _, ok := pbft.recoveryMgr.rcRspStore[from]; ok {
		logger.Debugf("Replica %d receive duplicate recovery response from replica %d, ignore it", pbft.id, from)
		return nil
	}
	pbft.recoveryMgr.rcRspStore[from] = rsp

	if len(pbft.recoveryMgr.rcRspStore) <= 2*pbft.f+1 {
		// Reason for not using '≤ pbft.N-pbft.f': if N==5, we are require more than we need
		logger.Debugf("Replica %d recv recoveryRsp from replica %d, rsp count: %d, not "+
			"beyond %d", pbft.id, rsp.ReplicaId, len(pbft.recoveryMgr.rcRspStore), 2*pbft.f+1)
		return nil
	}

	// find quorum chkpt
	n, lastid, replicas, find, chkptBehind := pbft.findHighestChkptQuorum()
	logger.Debug("n: ", n, "lastid: ", lastid, "replicas: ", replicas, "find: ", find, "chkptBehind: ", chkptBehind)
	lastExec, curHash, execFind := pbft.findLastExecQuorum()

	if !find {
		logger.Debugf("Replica %d did not find chkpt quorum", pbft.id)
		return nil
	}

	if !execFind {
		logger.Debugf("Replica %d did not find lastexec quorum", pbft.id)
		return nil
	}

	if pbft.recoveryMgr.recoveryToSeqNo != nil {
		logger.Debugf("Replica %d in recovery receive rcRsp from replica %d"+
			"but chkpt quorum and seqNo quorum already found. "+
			"Ignore it", pbft.id, rsp.ReplicaId)
		return nil
	}
	pbft.pbftTimerMgr.stopTimer(RECOVERY_RESTART_TIMER)
	pbft.recoveryMgr.recoveryToSeqNo = &lastExec

	//blockInfo := getBlockchainInfo()
	//id, _ := proto.Marshal(blockInfo)
	//idAsString := byteToString(id)
	selfLastExec, selfCurHash := persist.GetBlockHeightAndHash(pbft.namespace)

	logger.Debugf("Replica %d in recovery find quorum chkpt: %d, self: %d, "+
		"others lastExec: %d, self: %d", pbft.id, n, pbft.h, lastExec, pbft.exec.lastExec)
	logger.Debugf("Replica %d in recovery, "+
		"others lastBlockInfo: %s, self: %s", pbft.id, rsp.LastBlockHash, selfCurHash)

	// Fast catch up
	if lastExec == selfLastExec && curHash == selfCurHash {
		logger.Debugf("Replica %d in recovery same lastExec: %d, "+
			"same block hash: %s, fast catch up", pbft.id, selfLastExec, curHash)
		pbft.status.inActiveState(&pbft.status.inRecovery)
		pbft.recoveryMgr.recoveryToSeqNo = nil

		return &LocalEvent{
			Service:RECOVERY_SERVICE,
			EventType:RECOVERY_DONE_EVENT,
		}
	}

	logger.Debugf("Replica %d in recovery self lastExec: %d, others: %d"+
		"miss match self block hash: %s, other block hash %s", pbft.id, selfLastExec, lastExec, selfCurHash, curHash)

	var id []byte
	var err error
	if n != 0 {
		id, err = base64.StdEncoding.DecodeString(lastid)
		if nil != err {
			logger.Errorf("Replica %d cannot decode blockInfoId %s to %+v", pbft.id, lastid, id)
			return nil
		}
	} else {
		id = []byte("XXX GENESIS")
	}

	target := &stateUpdateTarget{
		checkpointMessage: checkpointMessage{
			seqNo: n,
			id:    id,
		},
		replicas: replicas,
	}

	pbft.updateHighStateTarget(target)

	if chkptBehind {
		pbft.moveWatermarks(n)
		pbft.stateTransfer(target)
		return nil
	} else if !pbft.status.getState(&pbft.status.skipInProgress) && !pbft.status.getState(&pbft.status.inVcReset) {
		pbft.helper.VcReset(n+1)
		//state := &LocalEvent{
		//	Service:CORE_PBFT_SERVICE,
		//	EventType:CORE_STATE_UPDATE_EVENT,
		//	Event:&stateUpdatedEvent{seqNo:n},
		//}
		//go pbft.pbftEventQueue.Push(state)
		pbft.status.activeState(&pbft.status.inVcReset)
		return nil
	} else {
		logger.Debugf("Replica %d try to recovery but find itself in state update", pbft.id)
		return nil
	}
}

// findHighestChkptQuorum finds highest one of chkpts which achieve quorum
func (pbft *pbftImpl) findHighestChkptQuorum() (n uint64, d string, replicas []uint64, find bool, chkptBehind bool) {

	logger.Debugf("Replica %d now enter findHighestChkptQuorum", pbft.id)

	chkpts := make(map[cidx]map[uint64]bool)

	for from, rsp := range pbft.recoveryMgr.rcRspStore {
		for chkptN, chkptD := range rsp.GetChkpts() {
			chkptIdx := cidx{
				n: chkptN,
				d: chkptD,
			}
			peers, ok := chkpts[chkptIdx]
			if ok {
				peers[from] = true
			} else {
				peers = make(map[uint64]bool)
				peers[from] = true
				chkpts[chkptIdx] = peers
			}
		}
	}

	find = false
	chkptBehind = false
	n = pbft.h

	// Since replica sends all of its chkpt, we may encounter several, instead of single one,
	// chkpts, which reach 2f+1

	// In this case, others will move watermarks sooner or later.
	// Hopefully, we find only one chkpt which reaches 2f+1 and this chkpt is their pbft.h
	for ci, peers := range chkpts {
		if len(peers) >= pbft.minimumCorrectQuorum() {
			find = true
			if ci.n >= n {
				if ci.n > n {
					chkptBehind = true
				}
				n = ci.n
				d = ci.d
				replicas = make([]uint64, 0, len(peers))
				for peer := range peers {
					replicas = append(replicas, peer)
				}
			}
		}
	}

	return
}

func (pbft *pbftImpl) findLastExecQuorum() (lastExec uint64, hash string, find bool) {

	lastExecs := make(map[blkIdx]map[uint64]bool)
	find = false
	for _, rsp := range pbft.recoveryMgr.rcRspStore {
		idx := blkIdx{
			height: rsp.BlockHeight,
			hash:   rsp.LastBlockHash,
		}
		replicas, ok := lastExecs[idx]
		if ok {
			replicas[rsp.ReplicaId] = true
		} else {
			replicas := make(map[uint64]bool)
			replicas[rsp.ReplicaId] = true
			lastExecs[idx] = replicas
		}

		if len(lastExecs[idx]) >= pbft.minimumCorrectQuorum() {
			lastExec = idx.height
			hash = idx.hash
			find = true
			break
		}
	}

	return
}

// fetchRecoveryPQC fetch PQC info after receive stateUpdated event
func (pbft *pbftImpl) fetchRecoveryPQC(peers []uint64) events.Event {

	logger.Debugf("Replica %d now fetchRecoveryPQC", pbft.id)

	if peers == nil {
		logger.Errorf("Replica %d try to fetchRecoveryPQC, but target peers are nil")
		return nil
	}

	pbft.recoveryMgr.rcPQCSenderStore = make(map[uint64]bool)

	fetch := &RecoveryFetchPQC{
		ReplicaId: pbft.id,
		H:         pbft.h,
	}

	payload, err := proto.Marshal(fetch)
	if err != nil {
		logger.Errorf("recovery response marshal error")
		return nil
	}
	conMsg := &ConsensusMessage{
		Type:    ConsensusMessage_RECOVERY_FETCH_QPC,
		Payload: payload,
	}

	for _, dest := range peers {
		msg := cMsgToPbMsg(conMsg, dest)
		pbft.helper.InnerUnicast(msg, dest)
	}

	return nil
}

// returnRecoveryPQC return  recovery PQC to the peer behind
func (pbft *pbftImpl) returnRecoveryPQC(fetch *RecoveryFetchPQC) events.Event {

	logger.Debugf("Replica %d now returnRecoveryPQC", pbft.id)

	dest, h := fetch.ReplicaId, fetch.H

	if h >= pbft.h+pbft.L {
		logger.Errorf("Replica %d receives fetch QPC request, but its pbft.h ≥ highwatermark", pbft.id)
		return nil
	}
	csLen := len(pbft.storeMgr.certStore) //certStore len
	prepres := make([]*PrePrepare, csLen)
	pres := make([]bool, csLen)
	cmts := make([]bool, csLen)
	i := 0
	for msgId, msgCert := range pbft.storeMgr.certStore {
		if msgId.n > h && msgId.n <= pbft.h+pbft.L {
			prepres[i] = msgCert.prePrepare
			pres[i] = msgCert.sentPrepare
			cmts[i] = msgCert.sentCommit
			i = i + 1
		}
	}
	rcReturn := &RecoveryReturnPQC{
		ReplicaId: pbft.id,
		PrepreSet: prepres,
		PreSent:   pres,
		CmtSent:   cmts,
	}

	payload, err := proto.Marshal(rcReturn)
	if err != nil {
		logger.Errorf("recovery response marshal error")
		return nil
	}
	msg := &ConsensusMessage{
		Type:    ConsensusMessage_RECOVERY_RETURN_QPC,
		Payload: payload,
	}
	conMsg := cMsgToPbMsg(msg, dest)
	pbft.helper.InnerUnicast(conMsg, dest)

	return nil
}

// recvRecoveryReturnPQC process PQC info target peers return
func (pbft *pbftImpl) recvRecoveryReturnPQC(PQCInfo *RecoveryReturnPQC) events.Event {

	logger.Debugf("Replica %d now recvRecoveryReturnPQC from replica %d", pbft.id, PQCInfo.ReplicaId)

	if !pbft.status.getState(&pbft.status.inRecovery) {
		logger.Warningf("Replica %d receive recoveryReturnQPC, but it's not in recovery", pbft.id)
		return nil
	}

	sender := PQCInfo.ReplicaId
	if _, exist := pbft.recoveryMgr.rcPQCSenderStore[sender]; exist {
		logger.Warningf("Replica %d receive duplicate RecoveryReturnPQC, ignore it", pbft.id)
		return nil
	}
	pbft.recoveryMgr.rcPQCSenderStore[sender] = true

	if len(pbft.recoveryMgr.rcPQCSenderStore) > pbft.f+1 {
		logger.Debugf("Replica %d already receive %d returnPQC", pbft.id, len(pbft.recoveryMgr.rcPQCSenderStore))
		pbft.pbftTimerMgr.stopTimer(RECOVERY_RESTART_TIMER)
	}

	prepreSet := PQCInfo.GetPrepreSet()
	preSent := PQCInfo.PreSent
	cmtSent := PQCInfo.CmtSent

	for i := 0; i < len(PQCInfo.PrepreSet); i++ {
		preprep := prepreSet[i]
		// recv preprepare
		cert := pbft.storeMgr.getCert(preprep.View, preprep.SequenceNumber)
		if cert.digest != preprep.BatchDigest {
			payload, err := proto.Marshal(preprep)
			if err != nil {
				logger.Errorf("ConsensusMessage_PRE_PREPARE Marshal Error", err)
				return false
			}

			redo_preprep := &ConsensusMessage{
				Type:    ConsensusMessage_PRE_PREPARE,
				Payload: payload,
			}
			go pbft.pbftEventQueue.Push(redo_preprep)
		}
		// recv prepare
		if preSent[i] {
			prep := &Prepare{
				View:           preprep.View,
				SequenceNumber: preprep.SequenceNumber,
				BatchDigest:    preprep.BatchDigest,
				ReplicaId:      sender,
			}
			payload, err := proto.Marshal(prep)
			if err != nil {
				logger.Errorf("ConsensusMessage_PREPARE Marshal Error", err)
				return nil
			}
			redo_prep := &ConsensusMessage{
				Type:    ConsensusMessage_PREPARE,
				Payload: payload,
			}
			go pbft.pbftEventQueue.Push(redo_prep)
		}
		// recv commit
		if cmtSent[i] {
			cmt := &Commit{
				View:           preprep.View,
				SequenceNumber: preprep.SequenceNumber,
				BatchDigest:    preprep.BatchDigest,
				ReplicaId:      sender,
			}
			payload, err := proto.Marshal(cmt)
			if err != nil {
				logger.Errorf("ConsensusMessage_COMMIT Marshal Error", err)
				return nil
			}
			redo_cmt := &ConsensusMessage{
				Type:    ConsensusMessage_COMMIT,
				Payload: payload,
			}
			go pbft.pbftEventQueue.Push(redo_cmt)
		}
	}

	return nil
}

// restartRecovery restart recovery immediately when recoveryRestartTimer expires
func (pbft *pbftImpl) restartRecovery() {

	logger.Noticef("Replica %d now restartRecovery", pbft.id)

	// recovery redo requires update new if need
	pbft.status.activeState(&pbft.status.inNegoView)
	pbft.status.activeState(&pbft.status.inRecovery)
	pbft.processNegotiateView()
}
