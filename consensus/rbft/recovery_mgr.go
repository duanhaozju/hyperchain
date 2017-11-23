//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import (
	"encoding/base64"

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
func (rbft *rbftImpl) dispatchRecoveryMsg(e consensusEvent) consensusEvent {
	switch et := e.(type) {
	case *NegotiateView:
		return rbft.recvNegoView(et)
	case *NegotiateViewResponse:
		return rbft.recvNegoViewRsp(et)
	case *RecoveryInit:
		return rbft.recvRecovery(et)
	case *RecoveryResponse:
		return rbft.recvRecoveryRsp(et)
	case *RecoveryFetchPQC:
		return rbft.returnRecoveryPQC(et)
	case *RecoveryReturnPQC:
		return rbft.recvRecoveryReturnPQC(et)
	}
	return nil
}

// initNegoView initializes negotiate view:
// 1. start negotiate view response timer
// 2. broadcast negotiate view
// 3. construct NegotiateViewResponse and send it to self
func (rbft *rbftImpl) initNegoView() error {

	if !rbft.in(inNegotiateView) {
		rbft.logger.Warningf("Replica %d try to send negotiateView, but it's not in negotiateView", rbft.id)
		return nil
	}

	rbft.setAbNormal()

	rbft.logger.Infof("Replica %d now negotiateView...", rbft.id)

	event := &LocalEvent{
		Service:   RECOVERY_SERVICE,
		EventType: RECOVERY_NEGO_VIEW_RSP_TIMER_EVENT,
	}

	rbft.timerMgr.startTimer(NEGO_VIEW_RSP_TIMER, event, rbft.eventMux)

	rbft.recoveryMgr.negoViewRspStore = make(map[uint64]*NegotiateViewResponse)

	// broadcast the negotiate message to other replica
	negoViewMsg := &NegotiateView{
		ReplicaId: rbft.id,
	}
	payload, err := proto.Marshal(negoViewMsg)
	if err != nil {
		rbft.logger.Errorf("Marshal negotiateView Error!")
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_NEGOTIATE_VIEW,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, rbft.id)
	rbft.helper.InnerBroadcast(msg)
	rbft.logger.Debugf("Replica %d sending negotiateView message", rbft.id)

	// post the negotiate message event to myself
	nvr := &NegotiateViewResponse{
		ReplicaId: rbft.id,
		View:      rbft.view,
		N:         uint64(rbft.N),
		Routers:   rbft.nodeMgr.routers,
	}
	consensusPayload, err := proto.Marshal(nvr)
	if err != nil {
		rbft.logger.Errorf("Marshal NegotiateViewResponse Error!")
		return nil
	}
	responseMsg := &ConsensusMessage{
		Type:    ConsensusMessage_NEGOTIATE_VIEW_RESPONSE,
		Payload: consensusPayload,
	}
	go rbft.eventMux.Post(responseMsg)

	return nil
}

// restartNegoView restarts negotiate view
func (rbft *rbftImpl) restartNegoView() {

	rbft.logger.Debugf("Replica %d restart negotiateView", rbft.id)
	rbft.initNegoView()
}

// recvNegoView firstly checks current state, then send response to the sender
func (rbft *rbftImpl) recvNegoView(nv *NegotiateView) consensusEvent {

	if rbft.in(inViewChange) {
		rbft.logger.Debugf("Replica %d is in viewChange, reject negotiateView from replica %d", rbft.id, nv.ReplicaId)
		return nil
	}
	sender := nv.ReplicaId
	rbft.logger.Debugf("Replica %d received negotiateView from %d", rbft.id, sender)

	negoViewRsp := &NegotiateViewResponse{
		ReplicaId: rbft.id,
		View:      rbft.view,
		N:         uint64(rbft.N),
		Routers:   rbft.nodeMgr.routers,
	}
	payload, err := proto.Marshal(negoViewRsp)
	if err != nil {
		rbft.logger.Errorf("Marshal NegotiateViewResponse Error!")
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_NEGOTIATE_VIEW_RESPONSE,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, rbft.id)
	rbft.helper.InnerUnicast(msg, sender)
	rbft.logger.Debugf("Replica %d sending negotiateViewResponse to replica %d, for view=%d/N=%d", rbft.id, sender, rbft.view, rbft.N)
	return nil
}

// recvNegoViewRsp stores the negotiate view response, if the counts of negoViewRsp reach the commonCaseQuorum, update
// current info if needed
func (rbft *rbftImpl) recvNegoViewRsp(nvr *NegotiateViewResponse) consensusEvent {

	if !rbft.in(inNegotiateView) {
		rbft.logger.Debugf("Replica %d already finished negotiateView, ignore incoming negotiateViewResponse from replica %d", rbft.id, nvr.ReplicaId)
		return nil
	}

	if rsp, ok := rbft.recoveryMgr.negoViewRspStore[nvr.ReplicaId]; ok {
		rbft.logger.Warningf("Replica %d already received negotiateView response from replica %d, "+
			"for view=%d, now receive again for view=%d, replace it", rbft.id, nvr.ReplicaId, rsp.View, nvr.View)
	}

	rbft.logger.Debugf("Replica %d received negotiateViewResponse from replica %d, view=%d/N=%d", rbft.id, nvr.ReplicaId, nvr.View, nvr.N)

	rbft.recoveryMgr.negoViewRspStore[nvr.ReplicaId] = nvr

	if len(rbft.recoveryMgr.negoViewRspStore) >= rbft.commonCaseQuorum() {
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
		for _, rs := range rbft.recoveryMgr.negoViewRspStore {
			ret := resp{rs.N, rs.View}
			if _, ok := viewCount[ret]; ok {
				viewCount[ret]++
			} else {
				viewCount[ret] = uint64(1)
			}
			if viewCount[ret] >= uint64(rbft.commonCaseQuorum()) {
				quorumResp = ret
				canFind = true
				break
			}
		}

		// we can find the quorum negoViewRsp with the same N and view, judge if the response.view equals to the
		// current view, if so, just update N and view, else update N, view and then re-constructs certStore
		if canFind {
			rbft.timerMgr.stopTimer(NEGO_VIEW_RSP_TIMER)
			var needUpdate bool
			if rbft.view == quorumResp.view {
				needUpdate = false
			} else {
				needUpdate = true
				rbft.view = quorumResp.view
			}
			// update N and f
			rbft.N = int(quorumResp.n)
			rbft.f = (rbft.N - 1) / 3
			rbft.off(inNegotiateView)
			if rbft.in(inViewChange) {
				rbft.off(inViewChange)
			}
			if needUpdate {
				rbft.parseSpecifyCertStore()
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
func (rbft *rbftImpl) initRecovery() consensusEvent {

	rbft.logger.Infof("Replica %d now initRecovery", rbft.id)

	rbft.recoveryMgr.recoveryToSeqNo = nil
	rbft.recoveryMgr.rcRspStore = make(map[uint64]*RecoveryResponse)

	// broadcast recovery message
	recoveryMsg := &RecoveryInit{
		ReplicaId: rbft.id,
	}
	payload, err := proto.Marshal(recoveryMsg)
	if err != nil {
		rbft.logger.Errorf("Marshal recovery init Error!")
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_RECOVERY_INIT,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, rbft.id)
	rbft.helper.InnerBroadcast(msg)

	// start recovery restart timer
	event := &LocalEvent{
		Service:   RECOVERY_SERVICE,
		EventType: RECOVERY_RESTART_TIMER_EVENT,
	}
	rbft.timerMgr.startTimer(RECOVERY_RESTART_TIMER, event, rbft.eventMux)

	// send RecoveryResponse to itself
	height, curHash := rbft.GetBlockHeightAndHash(rbft.namespace)
	genesis := rbft.getGenesisInfo()
	rc := &RecoveryResponse{
		ReplicaId:     rbft.id,
		Chkpts:        rbft.storeMgr.chkpts,
		BlockHeight:   height,
		LastBlockHash: curHash,
		Genesis:       genesis,
	}
	rbft.recvRecoveryRsp(rc)
	return nil
}

// restartRecovery restarts recovery immediately
func (rbft *rbftImpl) restartRecovery() {

	rbft.logger.Infof("Replica %d now restartRecovery", rbft.id)

	// clean the negoViewRspStore and recoveryRspStore
	rbft.recoveryMgr.negoViewRspStore = make(map[uint64]*NegotiateViewResponse)
	rbft.recoveryMgr.rcRspStore = make(map[uint64]*RecoveryResponse)

	rbft.on(inNegotiateView)
	rbft.on(inRecovery)
	rbft.initNegoView()
}

// recvRecovery processes incoming proactive recovery message and send the response to the sender
func (rbft *rbftImpl) recvRecovery(recoveryInit *RecoveryInit) consensusEvent {

	rbft.logger.Debugf("Replica %d now received recovery from replica %d", rbft.id, recoveryInit.ReplicaId)

	if rbft.in(skipInProgress) {
		rbft.logger.Debugf("Replica %d receive recovery, but it's in state transfer, ignore it.", rbft.id)
		return nil
	}

	height, curHash := rbft.GetBlockHeightAndHash(rbft.namespace)
	genesis := rbft.getGenesisInfo()

	rc := &RecoveryResponse{
		ReplicaId:     rbft.id,
		Chkpts:        rbft.storeMgr.chkpts,
		BlockHeight:   height,
		LastBlockHash: curHash,
		Genesis:       genesis,
	}

	rbft.logger.Debugf("Replica %d sending recovery response to replica %d is : %v", rbft.id, recoveryInit.ReplicaId, rc)

	rcMsg, err := proto.Marshal(rc)
	if err != nil {
		rbft.logger.Errorf("RecoveryResponse marshal error")
		return nil
	}

	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_RECOVERY_RESPONSE,
		Payload: rcMsg,
	}
	dest := recoveryInit.ReplicaId
	msg := cMsgToPbMsg(consensusMsg, rbft.id)
	rbft.helper.InnerUnicast(msg, dest)

	return nil
}

// recvRecoveryRsp stores the recovery response, if the counts of rcRspStore reach the commonCaseQuorum, find highest
// checkpoint quorum and use that checkpoint seqNo as the recovery target to continue recovery
func (rbft *rbftImpl) recvRecoveryRsp(rsp *RecoveryResponse) consensusEvent {

	rbft.logger.Debugf("Replica %d received recoveryResponse from replica %d for block height=%d", rbft.id, rsp.ReplicaId, rsp.BlockHeight)

	if !rbft.in(inRecovery) {
		rbft.logger.Debugf("Replica %d has finished recovery, ignore recoveryResponse", rbft.id)
		return nil
	}
	from := rsp.ReplicaId
	if _, ok := rbft.recoveryMgr.rcRspStore[from]; ok {
		rbft.logger.Warningf("Replica %d received recoveryResponse again from replica %d, replace it with height: %d", rbft.id, from, rsp.BlockHeight)
	}
	rbft.recoveryMgr.rcRspStore[from] = rsp

	if len(rbft.recoveryMgr.rcRspStore) < rbft.commonCaseQuorum() {
		rbft.logger.Debugf("Replica %d received recoveryResponse from replica %d, response count: %d, not "+
			"beyond %d, waiting...", rbft.id, rsp.ReplicaId, len(rbft.recoveryMgr.rcRspStore), rbft.commonCaseQuorum())
		return nil
	}

	if rbft.recoveryMgr.recoveryToSeqNo != nil {
		rbft.logger.Debugf("Replica %d received recoveryResponse from replica %d but checkpoint quorum and "+
			"seqNo quorum already found, ignore it", rbft.id, rsp.ReplicaId)
		return nil
	}

	// find quorum high checkpoint
	chkptSeqNo, chkptid, replicas, find, chkptBehind := rbft.findHighestChkptQuorum()
	rbft.logger.Debugf("chkptSeqNo: %d, chkptid: %s, find: %v, chkptBehind: %v", chkptSeqNo, chkptid, find, chkptBehind)

	if !find {
		rbft.logger.Debugf("Replica %d could not find chkpt quorum", rbft.id)
		return nil
	}

	rbft.timerMgr.stopTimer(RECOVERY_RESTART_TIMER)
	// use the highest checkpoint seqNo as recoveryToSeqNo, and fetchPQC after recovery done
	rbft.recoveryMgr.recoveryToSeqNo = &chkptSeqNo

	rbft.logger.Debugf("Replica %d in recovery self lastExec: %d, self h: %d, others' h: %d",
		rbft.id, rbft.exec.lastExec, rbft.h, chkptSeqNo)

	var chkptId []byte
	var err error
	if chkptSeqNo != 0 {
		chkptId, err = base64.StdEncoding.DecodeString(chkptid)
		if err != nil {
			rbft.logger.Errorf("Replica %d cannot decode blockInfoId %s to %+v", rbft.id, chkptid, chkptId)
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
	rbft.updateHighStateTarget(target)

	if chkptBehind {
		// if we are behind by checkpoint, move watermark and state transfer to the target
		rbft.moveWatermarks(chkptSeqNo)
		rbft.stateTransfer(target)
	} else if !rbft.inOne(skipInProgress, inVcReset) {
		// if we are not behind by checkpoint, just VcReset to delete the useless tmp validate result, after
		// VcResetDone, we will finish recovery or come into stateUpdate
		//rbft.helper.VcReset(rbft.exec.lastExec+1, rbft.view)
		//rbft.on(inVcReset)
		return &LocalEvent{
			Service:   RECOVERY_SERVICE,
			EventType: RECOVERY_DONE_EVENT,
		}
	} else {
		// if we are state transferring, just continue
		rbft.logger.Debugf("Replica %d try to recovery but find itself in stateUpdate", rbft.id)
	}
	return nil
}

// findHighestChkptQuorum finds highest one of chkpts which achieve oneCorrectQuorum
func (rbft *rbftImpl) findHighestChkptQuorum() (chkptSeqNo uint64, chkptId string, replicas []replicaInfo, find bool, chkptBehind bool) {

	chkpts := make(map[chkptID]map[replicaInfo]bool)

	// classify the checkpoints using Checkpoint index(seqNo,digest)
	for from, rsp := range rbft.recoveryMgr.rcRspStore {
		for chkptN, chkptD := range rsp.GetChkpts() {
			chkptIdx := chkptID{
				n:  chkptN,
				id: chkptD,
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
	chkptSeqNo = rbft.h

	// Since replica sends all of its chkpt, we may encounter several, instead of single one, chkpts, which
	// reach oneCorrectQuorum. in this case, others will move watermarks sooner or later, so we find the highest
	// quorum checkpoint as the recovery target.
	for ci, peers := range chkpts {
		if len(peers) >= rbft.oneCorrectQuorum() {
			find = true
			if ci.n >= chkptSeqNo {
				chkptSeqNo = ci.n
				chkptId = ci.id
				replicas = make([]replicaInfo, 0, len(peers))
				for peer := range peers {
					replicas = append(replicas, peer)
				}
			}
		}
	}
	if chkptSeqNo > rbft.h {
		chkptBehind = true
	}

	return
}

// fetchRecoveryPQC always fetches PQC info after recovery done to fetch PQC info after target checkpoint
func (rbft *rbftImpl) fetchRecoveryPQC() consensusEvent {

	rbft.logger.Debugf("Replica %d now fetchRecoveryPQC", rbft.id)
	rbft.recoveryMgr.rcPQCSenderStore = make(map[uint64]bool)

	fetch := &RecoveryFetchPQC{
		ReplicaId: rbft.id,
		H:         rbft.h,
	}
	payload, err := proto.Marshal(fetch)
	if err != nil {
		rbft.logger.Errorf("ConsensusMessage_RECOVERY_FETCH_QPC marshal error")
		return nil
	}
	conMsg := &ConsensusMessage{
		Type:    ConsensusMessage_RECOVERY_FETCH_QPC,
		Payload: payload,
	}

	msg := cMsgToPbMsg(conMsg, rbft.id)
	rbft.helper.InnerBroadcast(msg)

	return nil
}

// returnRecoveryPQC returns all PQC info we have sent before to the sender
func (rbft *rbftImpl) returnRecoveryPQC(fetch *RecoveryFetchPQC) consensusEvent {

	rbft.logger.Debugf("Replica %d now returnRecoveryPQC", rbft.id)

	destID, h := fetch.ReplicaId, fetch.H

	if h >= rbft.h+rbft.L {
		rbft.logger.Warningf("Replica %d receives recoveryFetchPQC request, but its rbft.h ≥ highwatermark", rbft.id)
		return nil
	}

	var prepres []*PrePrepare
	var pres []*Prepare
	var cmts []*Commit

	// replica just send all PQC info itself had sent before
	for idx, cert := range rbft.storeMgr.certStore {
		// send all PQC that n > h in current view, since it maybe wait others to execute
		if idx.n > h && idx.v == rbft.view {
			if cert.prePrepare == nil {
				rbft.logger.Warningf("Replica %d in returnRecoveryPQC finds nil prePrepare for view=%d/seqNo=%d",
					rbft.id, idx.v, idx.n)
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

	rcReturn := &RecoveryReturnPQC{ReplicaId: rbft.id}

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
		rbft.logger.Errorf("ConsensusMessage_RECOVERY_RETURN_QPC marshal error: %v", err)
		return nil
	}
	msg := &ConsensusMessage{
		Type:    ConsensusMessage_RECOVERY_RETURN_QPC,
		Payload: payload,
	}
	conMsg := cMsgToPbMsg(msg, rbft.id)
	rbft.helper.InnerUnicast(conMsg, destID)

	rbft.logger.Debugf("Replica %d send recoveryReturnPQC %v", rbft.id, rcReturn)

	return nil
}

// recvRecoveryReturnPQC re-processes all the PQC received from others
func (rbft *rbftImpl) recvRecoveryReturnPQC(PQCInfo *RecoveryReturnPQC) consensusEvent {

	rbft.logger.Debugf("Replica %d now recvRecoveryReturnPQC from replica %d, return_pqc %v", rbft.id, PQCInfo.ReplicaId, PQCInfo)

	sender := PQCInfo.ReplicaId
	if _, exist := rbft.recoveryMgr.rcPQCSenderStore[sender]; exist {
		rbft.logger.Warningf("Replica %d received duplicate recoveryReturnPQC from replica %d, ignore it", rbft.id, PQCInfo.ReplicaId)
		return nil
	}
	rbft.recoveryMgr.rcPQCSenderStore[sender] = true

	// post all the PQC
	if !rbft.isPrimary(rbft.id) {
		for _, preprep := range PQCInfo.GetPrepreSet() {
			rbft.recvPrePrepare(preprep)
		}
	}
	for _, prep := range PQCInfo.GetPreSet() {
		rbft.recvPrepare(prep)
	}
	for _, cmt := range PQCInfo.GetCmtSet() {
		rbft.recvCommit(cmt)
	}

	return nil
}
