//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"github.com/golang/protobuf/proto"
	"hyperchain/consensus/events"
	"encoding/base64"
	"hyperchain/consensus/helper/persist"
	"sync/atomic"
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

func newRecoveryMgr() *recoveryManager {
	rm := &recoveryManager{}
	rm.negoViewRspStore = make(map[uint64]*NegotiateViewResponse)
	rm.rcRspStore = make(map[uint64]*RecoveryResponse)
	rm.rcPQCSenderStore=make(map[uint64]bool)
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

func (pbft *pbftImpl) initNegoView() error {
	if !pbft.status.getState(&pbft.status.inNegoView) {
		pbft.logger.Errorf("Replica %d try to negotiateView, but it's not inNegoView. This indicates a bug", pbft.id)
		return nil
	}

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

func (pbft *pbftImpl) restartNegoView() {
	pbft.logger.Debugf("Replica %d restart negotiate view", pbft.id)
	pbft.initNegoView()
}

func (pbft *pbftImpl) recvNegoView(nv *NegotiateView) events.Event {
	if atomic.LoadUint32(&pbft.activeView) == 0 {
		pbft.logger.Warningf("Replica %d is in viewChange, reject negoView from replica %d", pbft.id, nv.ReplicaId)
		return nil
	}
	sender := nv.ReplicaId
	pbft.logger.Debugf("Replica %d receive nego_view from %d", pbft.id, sender)

	//if pbft.nodeMgr.routers == nil {
	//	pbft.logger.Debugf("Replica %d ignore nego_view from %d since has not received local msg", pbft.id, sender)
	//	return nil
	//}

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

func (pbft *pbftImpl) recvNegoViewRsp(nvr *NegotiateViewResponse) events.Event {
	if !pbft.status.getState(&pbft.status.inNegoView) {
		pbft.logger.Debugf("Replica %d already finished nego_view, ignore incoming nego_view_rsp", pbft.id)
		return nil
	}

	//rspId, rspView := nvr.ReplicaId, nvr.View
	if _, ok := pbft.recoveryMgr.negoViewRspStore[nvr.ReplicaId]; ok {
		pbft.logger.Warningf("Already recv view number from replica %d, replace it with view %d", nvr.ReplicaId, nvr.View)
	}

	pbft.logger.Debugf("Replica %d receive nego_view_rsp from replica %d, view=%d/N=%d", pbft.id, nvr.ReplicaId, nvr.View, nvr.N)

	pbft.recoveryMgr.negoViewRspStore[nvr.ReplicaId] = nvr

	if len(pbft.recoveryMgr.negoViewRspStore) >= pbft.commonCaseQuorum() {
		// Reason for not using '> pbft.N-pbft.f': if N==5, we are require more than we need
		// Reason for not using '≥ pbft.N-pbft.f': if self is wrong, then we are impossible to find 2f+1 same view
		// can we find same view from 2f+1 peers?
		type resp struct {
			n uint64
			view uint64
			//routers string
		}
		viewCount := make(map[resp]uint64)
		var result resp
		spFind := false
		canFind := false
		view := uint64(0)
		n := 0
		for _, rs := range pbft.recoveryMgr.negoViewRspStore {
			//r := byteToString(rs.Routers)
			ret := resp{rs.N, rs.View}
			if _, ok := viewCount[ret]; ok {
				viewCount[ret]++
			} else {
				viewCount[ret] = uint64(1)
			}
			if viewCount[ret] >= uint64(pbft.commonCaseQuorum()) {
				// yes we find the view
				result = ret
				canFind = true
				break
			}
		}
		for rs, count := range viewCount {
			if count >= uint64(pbft.commonCaseQuorum()-1) && rs.view != pbft.view {
				spFind = true
				view = rs.view
				n = int(rs.n)
			}
		}

		if canFind {
			pbft.timerMgr.stopTimer(NEGO_VIEW_RSP_TIMER)
			var IfUpdata bool
			if pbft.view==result.view{
				IfUpdata=true
			}else{
				pbft.view = result.view
			}
			pbft.N = int(result.n)
			pbft.f = (pbft.N - 1) / 3
			//routers, _ := stringToByte(result.routers)
			//if !bytes.Equal(routers, pbft.nodeMgr.routers) && !pbft.status.getState(&pbft.status.isNewNode) {
			//	pbft.nodeMgr.routers = routers
			//	pbft.logger.Debugf("Replica %d update routing table according to nego result", pbft.id)
			//	pbft.helper.NegoRouters(routers)
			//}
			pbft.status.inActiveState(&pbft.status.inNegoView)
			if atomic.LoadUint32(&pbft.activeView) == 0 {
				atomic.StoreUint32(&pbft.activeView, 1)
			}
			if !IfUpdata{
				pbft.parseCertStore()
			}
			return &LocalEvent{
				Service:RECOVERY_SERVICE,
				EventType:RECOVERY_NEGO_VIEW_DONE_EVENT,
			}
		} else if spFind {
			pbft.timerMgr.stopTimer(NEGO_VIEW_RSP_TIMER)
			pbft.view = view
			pbft.N = n
			pbft.f = (pbft.N - 1) / 3
			pbft.parseCertStore()
			pbft.status.inActiveState(&pbft.status.inNegoView)
			if atomic.LoadUint32(&pbft.activeView) == 0 {
				atomic.StoreUint32(&pbft.activeView, 1)
			}
			return &LocalEvent{
				Service:RECOVERY_SERVICE,
				EventType:RECOVERY_NEGO_VIEW_DONE_EVENT,
			}
		}
	}
	return nil
}

// procativeRecovery broadcast a procative recovery message to ask others for recent blocks info
func (pbft *pbftImpl) initRecovery() events.Event {

	pbft.logger.Debugf("Replica %d now initRecovery", pbft.id)

	pbft.recoveryMgr.recoveryToSeqNo = nil
	// update watermarks
	height := persist.GetHeightOfChain(pbft.namespace)

	pbft.recoveryMgr.rcRspStore = make(map[uint64]*RecoveryResponse)

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

	event := &LocalEvent{
		Service:   RECOVERY_SERVICE,
		EventType: RECOVERY_RESTART_TIMER_EVENT,
	}

	pbft.timerMgr.startTimer(RECOVERY_RESTART_TIMER, event, pbft.pbftEventQueue)

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

// restartRecovery restart recovery immediately when recoveryRestartTimer expires
func (pbft *pbftImpl) restartRecovery() {

	pbft.logger.Noticef("Replica %d now restartRecovery", pbft.id)
	// clean the negoViewRspStore and recoveryRspStore
	pbft.recoveryMgr.negoViewRspStore = make(map[uint64]*NegotiateViewResponse)
	pbft.recoveryMgr.rcRspStore = make(map[uint64]*RecoveryResponse)
	// recovery redo requires update new if need
	pbft.status.activeState(&pbft.status.inNegoView)
	pbft.status.activeState(&pbft.status.inRecovery)
	pbft.initNegoView()
}

// recvRcry process incoming proactive recovery message
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

// recvRcryRsp process other replicas' feedback as with initRecovery
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
		// Reason for not using '≤ pbft.N-pbft.f': if N==5, we are require more than we need
		pbft.logger.Debugf("Replica %d recv recoveryRsp from replica %d, rsp count: %d, not "+
			"beyond %d", pbft.id, rsp.ReplicaId, len(pbft.recoveryMgr.rcRspStore), pbft.commonCaseQuorum())
		return nil
	}

	if pbft.recoveryMgr.recoveryToSeqNo!=nil{
		pbft.logger.Debugf("Replica %d in recovery receive rcRsp from replica %d "+
			"but chkpt quorum and seqNo quorum already found. "+
			"Ignore it", pbft.id, rsp.ReplicaId)
		return nil
	}
	// find quorum chkpt
	n, lastid, replicas, find, chkptBehind := pbft.findHighestChkptQuorum()
	pbft.logger.Debug("n: ", n, "lastid: ", lastid, "find: ", find, "chkptBehind: ", chkptBehind)

	lastExec, curHash, execFind := pbft.findLastExecQuorum()
	pbft.logger.Debug("lastExec:", lastExec, "curHash:", curHash, "execFind:", execFind, "replicas:", replicas)

	if !find {
		pbft.logger.Debugf("Replica %d did not find chkpt quorum", pbft.id)
		return nil
	}

	if !execFind {
		pbft.logger.Debugf("Replica %d did not find lastexec quorum", pbft.id)
		return nil
	}
	pbft.timerMgr.stopTimer(RECOVERY_RESTART_TIMER)
	pbft.recoveryMgr.recoveryToSeqNo = &lastExec

	selfLastExec, selfCurHash := persist.GetBlockHeightAndHash(pbft.namespace)

	pbft.logger.Debugf("Replica %d in recovery find quorum chkpt: %d, self: %d, "+
		"others lastExec: %d, self: %d", pbft.id, n, pbft.h, lastExec, pbft.exec.lastExec)
	pbft.logger.Debugf("Replica %d in recovery, "+
		"others lastBlockInfo: %s, self: %s", pbft.id, rsp.LastBlockHash, selfCurHash)

	// Fast catch up
	if lastExec == selfLastExec && curHash == selfCurHash {
		pbft.logger.Debugf("Replica %d in recovery same lastExec: %d, "+
			"same block hash: %s, fast catch up", pbft.id, selfLastExec, curHash)
		//pbft.status.inActiveState(&pbft.status.inRecovery)
		//pbft.recoveryMgr.recoveryToSeqNo = nil
		pbft.seqNo = selfLastExec
		pbft.exec.lastExec = selfLastExec
		pbft.batchVdr.vid = selfLastExec
		pbft.batchVdr.lastVid = selfLastExec

		if pbft.primary(pbft.view) == pbft.id {
			for idx := range pbft.storeMgr.certStore {
				if idx.n > selfLastExec {
					delete(pbft.storeMgr.certStore, idx)
					pbft.persistDelQPCSet(idx.v, idx.n)
				}
			}
		}
		if !pbft.status.getState(&pbft.status.inVcReset) {
			pbft.helper.VcReset(selfLastExec + 1)
			pbft.status.activeState(&pbft.status.inVcReset)
		}
		return nil
	}

	pbft.logger.Debugf("Replica %d in recovery self lastExec: %d, others: %d"+
		"miss match self block hash: %s, other block hash %s", pbft.id, selfLastExec, lastExec, selfCurHash, curHash)

	var id []byte
	var err error
	if n != 0 {
		id, err = base64.StdEncoding.DecodeString(lastid)
		if nil != err {
			pbft.logger.Errorf("Replica %d cannot decode blockInfoId %s to %+v", pbft.id, lastid, id)
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
	} else if !pbft.status.getState(&pbft.status.skipInProgress) && !pbft.status.getState(&pbft.status.inVcReset) {
		pbft.helper.VcReset(selfLastExec + 1)
		pbft.status.activeState(&pbft.status.inVcReset)
	} else {
		pbft.logger.Debugf("Replica %d try to recovery but find itself in state update", pbft.id)
	}
	return nil
}

// findHighestChkptQuorum finds highest one of chkpts which achieve quorum
func (pbft *pbftImpl) findHighestChkptQuorum() (n uint64, d string, replicas []replicaInfo, find bool, chkptBehind bool) {

	pbft.logger.Debugf("Replica %d now enter findHighestChkptQuorum", pbft.id)

	chkpts := make(map[cidx]map[replicaInfo]bool)

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
	n = pbft.h

	// Since replica sends all of its chkpt, we may encounter several, instead of single one,
	// chkpts, which reach 2f+1

	// In this case, others will move watermarks sooner or later.
	// Hopefully, we find only one chkpt which reaches 2f+1 and this chkpt is their pbft.h
	for ci, peers := range chkpts {
		if len(peers) >= pbft.oneCorrectQuorum() {
			find = true
			if ci.n > n {
				chkptBehind = true
			}
			if ci.n >= n {
				n = ci.n
				d = ci.d
				replicas = make([]replicaInfo, 0, len(peers))
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

		if len(lastExecs[idx]) >= pbft.oneCorrectQuorum() {
			lastExec = idx.height
			hash = idx.hash
			find = true
			break
		}
	}

	return
}

// fetchRecoveryPQC fetch PQC info after receive stateUpdated event
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

// returnRecoveryPQC return  recovery PQC to the peer behind
func (pbft *pbftImpl) returnRecoveryPQC(fetch *RecoveryFetchPQC) events.Event {

	pbft.logger.Debugf("Replica %d now returnRecoveryPQC", pbft.id)

	dest, h := fetch.ReplicaId, fetch.H

	if h >= pbft.h+pbft.L {
		pbft.logger.Errorf("Replica %d receives fetch QPC request, but its pbft.h ≥ highwatermark", pbft.id)
		return nil
	}

	var prepres []*PrePrepare
	var pres []*Prepare
	var cmts []*Commit

	// replica just send all PQC that itself had sent
	for idx, cert := range pbft.storeMgr.certStore {
		// send all PQC that n > h, since it maybe wait others to execute
		if idx.n > h && idx.v == pbft.view {
			if cert.prePrepare == nil {
				pbft.logger.Warningf("Replica %d in returnRcPQC finds nil pre-prepare for view=%d/seqNo=%d",
					pbft.id, idx.v, idx.n)
			} else {
				prepres = append(prepres, cert.prePrepare)
			}
			for pre := range cert.prepare {
				if pre.ReplicaId != dest {
					prepare := pre
					pres = append(pres, &prepare)
					//break
				}
			}
			for cmt := range cert.commit {
				if cmt.ReplicaId != dest {
					commit := cmt
					cmts = append(cmts, &commit)
					//break
				}
			}
		}
	}
	rcReturn := &RecoveryReturnPQC{ ReplicaId: pbft.id }

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
	pbft.helper.InnerUnicast(conMsg, dest)

	pbft.logger.Debugf("Replica %d send recovery_return_pqc %v", pbft.id, rcReturn)

	return nil
}

// recvRecoveryReturnPQC process PQC info target peers return
func (pbft *pbftImpl) recvRecoveryReturnPQC(PQCInfo *RecoveryReturnPQC) events.Event {

	pbft.logger.Debugf("Replica %d now recvRecoveryReturnPQC from replica %d, return_pqc %v", pbft.id, PQCInfo.ReplicaId, PQCInfo)


	sender := PQCInfo.ReplicaId
	if _, exist := pbft.recoveryMgr.rcPQCSenderStore[sender]; exist {
		pbft.logger.Warningf("Replica %d receive duplicate RecoveryReturnPQC, ignore it", pbft.id)
		return nil
	}
	pbft.recoveryMgr.rcPQCSenderStore[sender] = true

	// post all the PQC
	for _, preprep := range PQCInfo.GetPrepreSet() {
		payload, err := proto.Marshal(preprep)
		if err != nil {
			pbft.logger.Errorf("ConsensusMessage_PRE_PREPARE Marshal Error", err)
			return nil
		}
		consensusMsg := &ConsensusMessage{
			Type:    ConsensusMessage_PRE_PREPARE,
			Payload: payload,
		}
		go pbft.pbftEventQueue.Push(consensusMsg)
	}
	for _, prep := range PQCInfo.GetPreSet() {
		payload, err := proto.Marshal(prep)
		if err != nil {
			pbft.logger.Errorf("ConsensusMessage_PRE_PREPARE Marshal Error", err)
			return nil
		}
		consensusMsg := &ConsensusMessage{
			Type:    ConsensusMessage_PREPARE,
			Payload: payload,
		}
		go pbft.pbftEventQueue.Push(consensusMsg)
	}
	for _, cmt := range PQCInfo.GetCmtSet() {
		payload, err := proto.Marshal(cmt)
		if err != nil {
			pbft.logger.Errorf("ConsensusMessage_COMMIT Marshal Error", err)
			return nil
		}
		consensusMsg := &ConsensusMessage{
			Type:    ConsensusMessage_COMMIT,
			Payload: payload,
		}
		go pbft.pbftEventQueue.Push(consensusMsg)
	}

	return nil
}
