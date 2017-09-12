//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"encoding/base64"
	"sync/atomic"

	"hyperchain/consensus/events"
	"hyperchain/consensus/helper/persist"

	"github.com/golang/protobuf/proto"
)

/**
This file contains recovery related issues
*/

// recoveryManager manages negotiate view and recovery related events
type recoveryManager struct {
	recoveryToSeqNo       *uint64                           // target seqNo expected to recover to
	rcRspStore            map[uint64]*RecoveryResponse      // store recovery responses received from other nodes
	rcPQCSenderStore      map[uint64]bool                   // store those who sent PQC info to self
	recvNewViewInRecovery bool                              // record whether receive new view during recovery
	negoViewRspStore      map[uint64]*NegotiateViewResponse // stores recovery negotiate view response received from other nodes
}

// newRecoveryMgr news an instance of recoveryManager
func newRecoveryMgr() *recoveryManager {
	rm := &recoveryManager{}
	rm.negoViewRspStore = make(map[uint64]*NegotiateViewResponse)
	rm.rcRspStore = make(map[uint64]*RecoveryResponse)
	rm.rcPQCSenderStore = make(map[uint64]bool)
	rm.recoveryToSeqNo = nil
	rm.recvNewViewInRecovery = false

	return rm
}

// dispatchRecoveryMsg dispatches recovery service messages using service type
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

// initNegoView initializes negotiate view:
// 1. start negotiate view response timer
// 2. broadcast negotiate view
// 3. construct NegotiateViewResponse and send it to self
func (pbft *pbftImpl) initNegoView() error {
	if !pbft.status.getState(&pbft.status.inNegoView) {
		pbft.logger.Errorf("Replica %d try to negotiateView, but it's not inNegoView. This indicates a bug", pbft.id)
		return nil
	}

	atomic.StoreUint32(&pbft.normal, 0)

	pbft.logger.Debugf("Replica %d now negotiate view...", pbft.id)

	event := &LocalEvent{
		Service:   RECOVERY_SERVICE,
		EventType: RECOVERY_NEGO_VIEW_RSP_TIMER_EVENT,
	}

	pbft.timerMgr.startTimer(NEGO_VIEW_RSP_TIMER, event, pbft.pbftEventQueue)

	pbft.recoveryMgr.negoViewRspStore = make(map[uint64]*NegotiateViewResponse)

	// broadcast the negotiate message to other replica
	negoViewMsg := &NegotiateView{
		ReplicaId: pbft.id,
	}
	payload, err := proto.Marshal(negoViewMsg)
	if err != nil {
		pbft.logger.Errorf("Marshal negotiateView Error!")
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_NEGOTIATE_VIEW,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, pbft.id)
	pbft.helper.InnerBroadcast(msg)
	pbft.logger.Debugf("Replica %d broadcast nego_view message", pbft.id)

	// post the negotiate message event to myself
	nvr := &NegotiateViewResponse{
		ReplicaId: pbft.id,
		View:      pbft.view,
		N:         uint64(pbft.N),
		Routers:   pbft.nodeMgr.routers,
	}
	consensusPayload, err := proto.Marshal(nvr)
	if err != nil {
		pbft.logger.Errorf("Marshal NegotiateViewResponse Error!")
		return nil
	}
	responseMsg := &ConsensusMessage{
		Type:    ConsensusMessage_NEGOTIATE_VIEW_RESPONSE,
		Payload: consensusPayload,
	}
	go pbft.pbftEventQueue.Push(responseMsg)

	return nil
}

// restartNegoView restarts negotiate view
func (pbft *pbftImpl) restartNegoView() {
	pbft.logger.Debugf("Replica %d restart negotiate view", pbft.id)
	pbft.initNegoView()
}

// recvNegoView firstly checks current state, then send response to the sender
func (pbft *pbftImpl) recvNegoView(nv *NegotiateView) events.Event {
	if atomic.LoadUint32(&pbft.activeView) == 0 {
		pbft.logger.Warningf("Replica %d is in viewChange, reject negoView from replica %d", pbft.id, nv.ReplicaId)
		return nil
	}
	sender := nv.ReplicaId
	pbft.logger.Debugf("Replica %d receive nego_view from %d", pbft.id, sender)

	negoViewRsp := &NegotiateViewResponse{
		ReplicaId: pbft.id,
		View:      pbft.view,
		N:         uint64(pbft.N),
		Routers:   pbft.nodeMgr.routers,
	}
	payload, err := proto.Marshal(negoViewRsp)
	if err != nil {
		pbft.logger.Errorf("Marshal NegotiateViewResponse Error!")
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_NEGOTIATE_VIEW_RESPONSE,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, pbft.id)
	pbft.helper.InnerUnicast(msg, sender)
	pbft.logger.Debugf("Replica %d send nego-view response to replica %d, for view=%d/N=%d", pbft.id, sender, pbft.view, pbft.N)
	return nil
}

// recvNegoViewRsp stores the negotiate view response, if the counts of negoViewRsp reach the commonCaseQuorum, update
// current info if needed
func (pbft *pbftImpl) recvNegoViewRsp(nvr *NegotiateViewResponse) events.Event {
	if !pbft.status.getState(&pbft.status.inNegoView) {
		pbft.logger.Debugf("Replica %d already finished nego_view, ignore incoming nego_view_rsp", pbft.id)
		return nil
	}

	if _, ok := pbft.recoveryMgr.negoViewRspStore[nvr.ReplicaId]; ok {
		pbft.logger.Warningf("Already recv view number from replica %d, replace it with view %d", nvr.ReplicaId, nvr.View)
	}

	pbft.logger.Debugf("Replica %d receive nego_view_rsp from replica %d, view=%d/N=%d", pbft.id, nvr.ReplicaId, nvr.View, nvr.N)

	pbft.recoveryMgr.negoViewRspStore[nvr.ReplicaId] = nvr

	if len(pbft.recoveryMgr.negoViewRspStore) >= pbft.commonCaseQuorum() {
		// can we find same view from 2f+1 peers?
		type resp struct {
			n    uint64
			view uint64
		}
		viewCount := make(map[resp]uint64)

		// check if we can find quorum negoViewRsp whose have the same n and view, if we can find, which means
		// quorum nodes agree to a N and view, save to quorumResp, set canFind to true and update N, view if needed
		var quorumResp resp
		canFind := false

		// find the quorum negoViewRsp
		for _, rs := range pbft.recoveryMgr.negoViewRspStore {
			ret := resp{rs.N, rs.View}
			if _, ok := viewCount[ret]; ok {
				viewCount[ret]++
			} else {
				viewCount[ret] = uint64(1)
			}
			if viewCount[ret] >= uint64(pbft.commonCaseQuorum()) {
				quorumResp = ret
				canFind = true
				break
			}
		}

		// we can find the quorum negoViewRsp with the same N and view, judge if the response.view equals to the
		// current view, if so, just update N and view, else update N, view and then re-constructs certStore
		if canFind {
			pbft.timerMgr.stopTimer(NEGO_VIEW_RSP_TIMER)
			var needUpdate bool
			if pbft.view == quorumResp.view {
				needUpdate = false
			} else {
				needUpdate = true
				pbft.view = quorumResp.view
			}
			// update N and f
			pbft.N = int(quorumResp.n)
			pbft.f = (pbft.N - 1) / 3
			pbft.status.inActiveState(&pbft.status.inNegoView)
			if atomic.LoadUint32(&pbft.activeView) == 0 {
				atomic.StoreUint32(&pbft.activeView, 1)
			}
			if needUpdate {
				pbft.parseSpecifyCertStore()
			}
			return &LocalEvent{
				Service:   RECOVERY_SERVICE,
				EventType: RECOVERY_NEGO_VIEW_DONE_EVENT,
			}
		}
	}
	return nil
}

// initRecovery broadcasts a procative recovery message to ask others for recent blocks info and send RecoveryResponse
// to itself
func (pbft *pbftImpl) initRecovery() events.Event {
	pbft.logger.Debugf("Replica %d now initRecovery", pbft.id)

	pbft.recoveryMgr.recoveryToSeqNo = nil
	pbft.recoveryMgr.rcRspStore = make(map[uint64]*RecoveryResponse)

	// broadcast recovery message
	recoveryMsg := &RecoveryInit{
		ReplicaId: pbft.id,
	}
	payload, err := proto.Marshal(recoveryMsg)
	if err != nil {
		pbft.logger.Errorf("Marshal recovery init Error!")
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_RECOVERY_INIT,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, pbft.id)
	pbft.helper.InnerBroadcast(msg)

	// start recovery restart timer
	event := &LocalEvent{
		Service:   RECOVERY_SERVICE,
		EventType: RECOVERY_RESTART_TIMER_EVENT,
	}
	pbft.timerMgr.startTimer(RECOVERY_RESTART_TIMER, event, pbft.pbftEventQueue)

	// send RecoveryResponse to itself
	height, curHash := persist.GetBlockHeightAndHash(pbft.namespace)
	genesis := pbft.getGenesisInfo()
	rc := &RecoveryResponse{
		ReplicaId:     pbft.id,
		Chkpts:        pbft.storeMgr.chkpts,
		BlockHeight:   height,
		LastBlockHash: curHash,
		Genesis:       genesis,
	}
	pbft.recvRecoveryRsp(rc)
	return nil
}

// restartRecovery restarts recovery immediately
func (pbft *pbftImpl) restartRecovery() {
	pbft.logger.Noticef("Replica %d now restartRecovery", pbft.id)

	// clean the negoViewRspStore and recoveryRspStore
	pbft.recoveryMgr.negoViewRspStore = make(map[uint64]*NegotiateViewResponse)
	pbft.recoveryMgr.rcRspStore = make(map[uint64]*RecoveryResponse)

	pbft.status.activeState(&pbft.status.inNegoView)
	pbft.status.activeState(&pbft.status.inRecovery)
	pbft.initNegoView()
}

// recvRecovery processes incoming proactive recovery message and send the response to the sender
func (pbft *pbftImpl) recvRecovery(recoveryInit *RecoveryInit) events.Event {
	pbft.logger.Debugf("Replica %d now recvRecovery from replica %d", pbft.id, recoveryInit.ReplicaId)

	if pbft.status.getState(&pbft.status.skipInProgress) {
		pbft.logger.Debugf("Replica %d recvRecovery, but it's in state transfer and ignores it.", pbft.id)
		return nil
	}

	height, curHash := persist.GetBlockHeightAndHash(pbft.namespace)
	genesis := pbft.getGenesisInfo()

	rc := &RecoveryResponse{
		ReplicaId:     pbft.id,
		Chkpts:        pbft.storeMgr.chkpts,
		BlockHeight:   height,
		LastBlockHash: curHash,
		Genesis:       genesis,
	}

	pbft.logger.Debugf("Replica %d recovery response to replica %d is : %v", pbft.id, recoveryInit.ReplicaId, rc)

	rcMsg, err := proto.Marshal(rc)
	if err != nil {
		pbft.logger.Errorf("RecoveryResponse marshal error")
		return nil
	}

	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_RECOVERY_RESPONSE,
		Payload: rcMsg,
	}
	dest := recoveryInit.ReplicaId
	msg := cMsgToPbMsg(consensusMsg, pbft.id)
	pbft.helper.InnerUnicast(msg, dest)

	return nil
}

// recvRecoveryRsp stores the recovery response, if the counts of rcRspStore reach the commonCaseQuorum, find highest
// checkpoint quorum and use that checkpoint seqNo as the recovery target to continue recovery
func (pbft *pbftImpl) recvRecoveryRsp(rsp *RecoveryResponse) events.Event {
	pbft.logger.Debugf("Replica %d now recvRecoveryRsp from replica %d for block height=%d", pbft.id, rsp.ReplicaId, rsp.BlockHeight)

	if !pbft.status.getState(&pbft.status.inRecovery) {
		pbft.logger.Debugf("Replica %d finished recovery, ignore recovery response", pbft.id)
		return nil
	}
	from := rsp.ReplicaId
	if _, ok := pbft.recoveryMgr.rcRspStore[from]; ok {
		pbft.logger.Debugf("Replica %d receive recovery response again from replica %d, replace it with height: %d", pbft.id, from, rsp.BlockHeight)
	}
	pbft.recoveryMgr.rcRspStore[from] = rsp

	if len(pbft.recoveryMgr.rcRspStore) < pbft.commonCaseQuorum() {
		pbft.logger.Debugf("Replica %d recv recoveryRsp from replica %d, rsp count: %d, not "+
			"beyond %d, waiting...", pbft.id, rsp.ReplicaId, len(pbft.recoveryMgr.rcRspStore), pbft.commonCaseQuorum())
		return nil
	}

	if pbft.recoveryMgr.recoveryToSeqNo != nil {
		pbft.logger.Debugf("Replica %d in recovery receive rcRsp from replica %d "+
			"but chkpt quorum and seqNo quorum already found. "+
			"Ignore it", pbft.id, rsp.ReplicaId)
		return nil
	}

	// find quorum high checkpoint
	chkptSeqNo, chkptid, replicas, find, chkptBehind := pbft.findHighestChkptQuorum()
	pbft.logger.Debug("chkptSeqNo: ", chkptSeqNo, "chkptid: ", chkptid, "find: ", find, "chkptBehind: ", chkptBehind)

	if !find {
		pbft.logger.Debugf("Replica %d did not find chkpt quorum", pbft.id)
		return nil
	}

	pbft.timerMgr.stopTimer(RECOVERY_RESTART_TIMER)
	// use the highest checkpoint seqNo as recoveryToSeqNo, and fetchPQC after recovery done
	pbft.recoveryMgr.recoveryToSeqNo = &chkptSeqNo

	pbft.logger.Debugf("Replica %d in recovery self lastExec: %d, self h: %d, others h: %d",
		pbft.id, pbft.exec.lastExec, pbft.h, chkptSeqNo)

	var chkptId []byte
	var err error
	if chkptSeqNo != 0 {
		chkptId, err = base64.StdEncoding.DecodeString(chkptid)
		if err != nil {
			pbft.logger.Errorf("Replica %d cannot decode blockInfoId %s to %+v", pbft.id, chkptid, chkptId)
			return nil
		}
	} else {
		chkptId = []byte("XXX GENESIS")
	}

	target := &stateUpdateTarget{
		checkpointMessage: checkpointMessage{
			seqNo: chkptSeqNo,
			id:    chkptId,
		},
		replicas: replicas,
	}
	pbft.updateHighStateTarget(target)

	if chkptBehind {
		// if we are behind by checkpoint, move watermark and state transfer to the target
		pbft.moveWatermarks(chkptSeqNo)
		pbft.stateTransfer(target)
	} else if !pbft.status.getState(&pbft.status.skipInProgress) && !pbft.status.getState(&pbft.status.inVcReset) {
		// if we are not behind by checkpoint, just VcReset to delete the useless tmp validate result, after
		// VcResetDone, we will finish recovery or come into stateUpdate
		pbft.helper.VcReset(pbft.exec.lastExec + 1)
		pbft.status.activeState(&pbft.status.inVcReset)
	} else {
		// if we are state transferring, just continue
		pbft.logger.Debugf("Replica %d try to recovery but find itself in state update", pbft.id)
	}
	return nil
}

// findHighestChkptQuorum finds highest one of chkpts which achieve oneCorrectQuorum
func (pbft *pbftImpl) findHighestChkptQuorum() (chkptSeqNo uint64, chkptId string, replicas []replicaInfo, find bool, chkptBehind bool) {
	pbft.logger.Debugf("Replica %d now enter findHighestChkptQuorum", pbft.id)

	chkpts := make(map[cidx]map[replicaInfo]bool)

	// classify the checkpoints using Checkpoint index(seqNo,digest)
	for from, rsp := range pbft.recoveryMgr.rcRspStore {
		for chkptN, chkptD := range rsp.GetChkpts() {
			chkptIdx := cidx{
				n: chkptN,
				d: chkptD,
			}
			peers, ok := chkpts[chkptIdx]
			if ok {
				replica := replicaInfo{
					id:      from,
					height:  rsp.BlockHeight,
					genesis: rsp.Genesis,
				}
				peers[replica] = true
			} else {
				peers = make(map[replicaInfo]bool)
				replica := replicaInfo{
					id:      from,
					height:  rsp.BlockHeight,
					genesis: rsp.Genesis,
				}
				peers[replica] = true
				chkpts[chkptIdx] = peers
			}
		}
	}

	find = false
	chkptBehind = false
	chkptSeqNo = pbft.h

	// Since replica sends all of its chkpt, we may encounter several, instead of single one, chkpts, which
	// reach oneCorrectQuorum. In this case, others will move watermarks sooner or later, so we find the highest
	// quorum checkpoint as the recovery target.
	for ci, peers := range chkpts {
		if len(peers) >= pbft.oneCorrectQuorum() {
			find = true
			if ci.n >= chkptSeqNo {
				chkptSeqNo = ci.n
				chkptId = ci.d
				replicas = make([]replicaInfo, 0, len(peers))
				for peer := range peers {
					replicas = append(replicas, peer)
				}
			}
		}
	}
	if chkptSeqNo > pbft.h {
		chkptBehind = true
	}

	return
}

// fetchRecoveryPQC always fetches PQC info after recovery done to fetch PQC info after target checkpoint
func (pbft *pbftImpl) fetchRecoveryPQC() events.Event {
	pbft.logger.Debugf("Replica %d now fetchRecoveryPQC", pbft.id)
	pbft.recoveryMgr.rcPQCSenderStore = make(map[uint64]bool)

	fetch := &RecoveryFetchPQC{
		ReplicaId: pbft.id,
		H:         pbft.h,
	}
	payload, err := proto.Marshal(fetch)
	if err != nil {
		pbft.logger.Errorf("ConsensusMessage_RECOVERY_FETCH_QPC marshal error")
		return nil
	}
	conMsg := &ConsensusMessage{
		Type:    ConsensusMessage_RECOVERY_FETCH_QPC,
		Payload: payload,
	}

	msg := cMsgToPbMsg(conMsg, pbft.id)
	pbft.helper.InnerBroadcast(msg)

	return nil
}

// returnRecoveryPQC returns all PQC info we have sent before to the sender
func (pbft *pbftImpl) returnRecoveryPQC(fetch *RecoveryFetchPQC) events.Event {
	pbft.logger.Debugf("Replica %d now returnRecoveryPQC", pbft.id)

	destID, h := fetch.ReplicaId, fetch.H

	if h >= pbft.h+pbft.L {
		pbft.logger.Errorf("Replica %d receives fetch QPC request, but its pbft.h ≥ highwatermark", pbft.id)
		return nil
	}

	var prepres []*PrePrepare
	var pres []*Prepare
	var cmts []*Commit
	var vid uint64

	// replica just send all PQC info itself had sent before
	for idx, cert := range pbft.storeMgr.certStore {
		// send all PQC that n > h in current view, since it maybe wait others to execute
		if idx.n > h && idx.v == pbft.view {
			if idx.n == h+1 {
				vid = cert.vid
			}
			if cert.prePrepare == nil {
				pbft.logger.Warningf("Replica %d in returnRcPQC finds nil pre-prepare for view=%d/seqNo=%d",
					pbft.id, idx.v, idx.n)
			} else {
				prepres = append(prepres, cert.prePrepare)
			}
			for pre := range cert.prepare {
				// remove the prepare sent from destID as the dest node must have these messages
				if pre.ReplicaId != destID {
					prepare := pre
					pres = append(pres, &prepare)
				}
			}
			for cmt := range cert.commit {
				// remove the commit sent from destID as the dest node must have these messages
				if cmt.ReplicaId != destID {
					commit := cmt
					cmts = append(cmts, &commit)
				}
			}
		}
	}
	for idx, prep := range pbft.batchVdr.spNullRequest {
		if idx.v == pbft.view && idx.n >= vid {
			prepres = append(prepres, prep)
		}
	}
	rcReturn := &RecoveryReturnPQC{ReplicaId: pbft.id}

	// in case replica doesn't have PQC, we cannot assign a nil one
	if prepres != nil {
		rcReturn.PrepreSet = prepres
	}
	if pres != nil {
		rcReturn.PreSet = pres
	}
	if cmts != nil {
		rcReturn.CmtSet = cmts
	}

	payload, err := proto.Marshal(rcReturn)
	if err != nil {
		pbft.logger.Errorf("ConsensusMessage_RECOVERY_RETURN_QPC marshal error: %v", err)
		return nil
	}
	msg := &ConsensusMessage{
		Type:    ConsensusMessage_RECOVERY_RETURN_QPC,
		Payload: payload,
	}
	conMsg := cMsgToPbMsg(msg, pbft.id)
	pbft.helper.InnerUnicast(conMsg, destID)

	pbft.logger.Debugf("Replica %d send recovery_return_pqc %v", pbft.id, rcReturn)

	return nil
}

// recvRecoveryReturnPQC re-processes all the PQC received from others
func (pbft *pbftImpl) recvRecoveryReturnPQC(PQCInfo *RecoveryReturnPQC) events.Event {
	pbft.logger.Debugf("Replica %d now recvRecoveryReturnPQC from replica %d, return_pqc %v", pbft.id, PQCInfo.ReplicaId, PQCInfo)

	sender := PQCInfo.ReplicaId
	if _, exist := pbft.recoveryMgr.rcPQCSenderStore[sender]; exist {
		pbft.logger.Warningf("Replica %d receive duplicate RecoveryReturnPQC, ignore it", pbft.id)
		return nil
	}
	pbft.recoveryMgr.rcPQCSenderStore[sender] = true

	// post all the PQC
	if pbft.primary(pbft.view) != pbft.id {
		for _, preprep := range PQCInfo.GetPrepreSet() {
			pbft.recvPrePrepare(preprep)
		}
	}
	for _, prep := range PQCInfo.GetPreSet() {
		pbft.recvPrepare(prep)
	}
	for _, cmt := range PQCInfo.GetCmtSet() {
		pbft.recvCommit(cmt)
	}

	return nil
}
