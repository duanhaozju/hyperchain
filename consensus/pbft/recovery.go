//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"github.com/golang/protobuf/proto"
	"hyperchain/consensus/events"
	"encoding/base64"
	"fmt"
)

// procativeRecovery broadcast a procative recovery message to ask others for recent blocks info
func (pbft *pbftProtocal) initRecovery() events.Event {

	logger.Debugf("Replica %d now initRecovery", pbft.id)

	pbft.rcRspStore = make(map[uint64]*RecoveryResponse)

	recoveryMsg := &RecoveryInit{
		ReplicaId: pbft.id,
	}
	payload, err := proto.Marshal(recoveryMsg)
	if err != nil {
		logger.Errorf("Marshal recovery init Error!")
		return nil
	}
	consensusMsg := &ConsensusMessage {
		Type: ConsensusMessage_RECOVERY_INIT,
		Payload: payload,
	}
	msg := consensusMsgHelper(consensusMsg, pbft.id)
	pbft.helper.InnerBroadcast(msg)
	pbft.recoveryRestartTimer.Reset(pbft.recoveryRestartTimeout, recoveryRestartTimerEvent{})

	chkpts := make(map[uint64]string)
	for n, d := range pbft.chkpts {
		chkpts[n] = d
	}
	rc := &RecoveryResponse{
		ReplicaId:	pbft.id,
		Chkpts: 	chkpts,
	}
	pbft.recvRecoveryRsp(rc)
	return nil
}

// recvRcry process incoming proactive recovery message
func (pbft *pbftProtocal) recvRecovery(recoveryInit *RecoveryInit) events.Event {

	logger.Debugf("Replica %d now recvRecovery from replica %d", pbft.id, recoveryInit.ReplicaId)

	if pbft.skipInProgress {
		logger.Debugf("Replica %d recvRecovery, but it's in state transfer and ignores it.", pbft.id)
		return nil
	}
	chkpts := make(map[uint64]string)
	for n, d := range pbft.chkpts {
		chkpts[n] = d
	}

	lastExec := pbft.lastExec

	rc := &RecoveryResponse{
		ReplicaId:	pbft.id,
		Chkpts:		chkpts,
		LastExec:   lastExec,
	}

	rcMsg, err := proto.Marshal(rc)
	if err != nil {
		logger.Errorf("recovery response marshal error")
		return nil
	}

	consensusMsg := &ConsensusMessage{
		Type: 		ConsensusMessage_RECOVERY_RESPONSE,
		Payload: 	rcMsg,
	}
	dest := recoveryInit.ReplicaId
	msg := consensusMsgHelper(consensusMsg, dest)
	pbft.helper.InnerUnicast(msg, dest)

	return nil
}

// recvRcryRsp process other replicas' feedback as with initRecovery
func (pbft *pbftProtocal) recvRecoveryRsp(rsp *RecoveryResponse) events.Event {

	logger.Debugf("Replica %d now recvRecoveryRsp from replica %d", pbft.id, rsp.ReplicaId)

	if !pbft.inRecovery {
		logger.Debugf("Replica %d finished recovery, ignore recovery response", pbft.id)
		return nil
	}
	from := rsp.ReplicaId
	if _, ok := pbft.rcRspStore[from]; ok {
		logger.Debugf("Replica %d receive duplicate recovery response from replica %d, ignore it", pbft.id, from)
		return nil
	}
	pbft.rcRspStore[from] = rsp

	// find quorum chkpt
	if len(pbft.rcRspStore) > pbft.N-pbft.f {
		n, d, replicas, find, chkptBehind := pbft.findHighestChkptQuorum()
		lastExec, peers := pbft.findLastExecQuorum()

		if find {
			pbft.recoveryRestartTimer.Stop()
			pbft.recoveryToSeqNo = lastExec

			if chkptBehind {
				logger.Noticef("Replica %d in recovery find chkpt: %d, behind, others seqNo: %d, self: %d", pbft.id, n, lastExec, pbft.lastExec)

				pbft.moveWatermarks(n)

				id, err := base64.StdEncoding.DecodeString(d)
				if nil != err {
					err = fmt.Errorf("Replica %d received a view change whose hash could not be decoded (%s)", pbft.id, d)
					logger.Error(err.Error())
					return nil
				}
				target := &stateUpdateTarget{
					checkpointMessage: checkpointMessage{
						seqNo: n,
						id:    id,
					},
					replicas: replicas,
				}

				pbft.updateHighStateTarget(target)
				pbft.stateTransfer(target)
			} else if pbft.lastExec<lastExec {
				// This is a somewhat subtle situation: we are not behind by checkpoint, but are  behind by seqNo
				logger.Noticef("Replica %d in recovery find chkpt, same: %d, different lastExec, self: %d, others: %d" , pbft.id, n, pbft.lastExec, lastExec)
				pbft.recoveryRestartTimer.Reset(pbft.recoveryRestartTimeout, recoveryRestartTimerEvent{})
				pbft.fetchRecoveryPQC(peers)

			} else if pbft.lastExec==lastExec {
				// This case indicates we are exactly the same as others
				logger.Noticef("Replica %d in recovery find chkpt, same: %d, same lastExec: %d", pbft.id, n, pbft.lastExec)
				pbft.inRecovery = false
				return recoveryDoneEvent{}
			} else {
				logger.Errorf("This should not happen! Replica %d in recovery find chkpt, same: %d, but self.lastExec: is ahead of others: %d", pbft.id, n, pbft.lastExec, lastExec)
			}
		}
	}
	return nil
}

// findHighestChkptQuorum finds highest one of chkpts which achieve quorum
func (pbft *pbftProtocal) findHighestChkptQuorum() (n uint64, d string, replicas []uint64, find bool, chkptBehind bool) {

	logger.Debugf("Replica %d now enter findHighestChkptQuorum", pbft.id)

	chkpts := make(map[cidx]map[uint64]bool)

	for from, rsp := range pbft.rcRspStore {
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

	// Since replica sends all of its chkpt, we may encounter several chkpts which reach 2f+1.
	// In this case, others will move watermarks sooner or later.
	// Hopefully, we find only one chkpt which reaches 2f+1 and this chkpt is their pbft.h
	for ci, peers := range chkpts {
		if len(peers) >= 2*pbft.f+1 {
			find = true
			if ci.n > n {
				chkptBehind = true
				n = ci.n
				d = ci.d
				replicas = make([]uint64, len(peers))
				for peer := range peers {
					replicas = append(replicas, peer)
				}
			}
		}
	}

	return
}

func (pbft *pbftProtocal) findLastExecQuorum() (lastExec uint64, peers []uint64) {

	lastExecs := make(map[uint64]map[uint64]bool)
	peers	  = make([]uint64, 2*pbft.f+1)
	for _, rsp := range pbft.rcRspStore {

		replicas, ok := lastExecs[rsp.LastExec]
		if ok {
			replicas[rsp.ReplicaId] = true
		} else {
			replicas := make(map[uint64]bool)
			replicas[rsp.ReplicaId] = true
			lastExecs[rsp.LastExec] = replicas
		}

		if len(lastExecs[rsp.LastExec]) >= 2*pbft.f+1 {
			lastExec = rsp.LastExec
			replicas = lastExecs[rsp.LastExec]
			for peer := range replicas {
				peers = append(peers, peer)
			}
			break
		}
	}

	return
}

// fetchRecoveryPQC fetch PQC info after receive stateUpdated event
func (pbft *pbftProtocal) fetchRecoveryPQC(peers []uint64) events.Event {

	logger.Debugf("Replica %d now fetchRecoveryPQC", pbft.id)

	if peers==nil {
		logger.Errorf("Replica %d try to fetchRecoveryPQC, but target peers are nil")
		return nil
	}

	pbft.rcPQCSenderStore = make(map[uint64]bool)

	fetch := &RecoveryFetchPQC{
		ReplicaId: pbft.id,
		H:	   pbft.h,
	}

	payload, err := proto.Marshal(fetch)
	if err != nil {
		logger.Errorf("recovery response marshal error")
		return nil
	}
	conMsg := &ConsensusMessage{
		Type:	 ConsensusMessage_RECOVERY_FETCH_QPC,
		Payload: payload,
	}

	for _, dest := range peers {
		msg := consensusMsgHelper(conMsg, dest)
		pbft.helper.InnerUnicast(msg, dest)
	}

	return nil
}

// returnRecoveryPQC return  recovery PQC to the peer behind
func (pbft *pbftProtocal) returnRecoveryPQC(fetch *RecoveryFetchPQC) events.Event {

	logger.Noticef("Replica %d now returnRecoveryPQC", pbft.id)

	dest, h := fetch.ReplicaId, fetch.H

	if h >= pbft.h+pbft.L {
		logger.Errorf("Replica %d receives fetch QPC request, but its pbft.h ≥ highwatermark", pbft.id)
		return nil
	}

	prepres := make([]*PrePrepare, len(pbft.certStore))
	pres    := make([]bool, len(pbft.certStore))
	cmts    := make([]bool, len(pbft.certStore))
	i := 0
	for msgId, msgCert := range pbft.certStore {
		if msgId.n > h && msgId.n <= pbft.h + pbft.L {
			prepres[i] = msgCert.prePrepare
			pres[i] = msgCert.sentPrepare
			cmts[i] = msgCert.sentCommit
			i = i + 1
		}
	}
	rcReturn := &RecoveryReturnPQC{
		ReplicaId:	pbft.id,
		PrepreSet:	prepres,
		PreSent:	pres,
		CmtSent:	cmts,
	}

	payload, err := proto.Marshal(rcReturn)
	if err != nil {
		logger.Errorf("recovery response marshal error")
		return nil
	}
	msg := &ConsensusMessage{
		Type:		ConsensusMessage_RECOVERY_RETURN_QPC,
		Payload:	payload,
	}
	conMsg := consensusMsgHelper(msg, dest)
	pbft.helper.InnerUnicast(conMsg, dest)

	return nil
}

// recvRecoveryReturnPQC process PQC info target peers return
func (pbft *pbftProtocal) recvRecoveryReturnPQC(PQCInfo *RecoveryReturnPQC) events.Event {

	logger.Debugf("Replica %d now recvRecoveryReturnPQC from replica %d", pbft.id, PQCInfo.ReplicaId)

	if !pbft.inRecovery {
		logger.Warningf("Replica %d receive recoveryReturnQPC, but it's not in recovery", pbft.id)
		return nil
	}

	sender := PQCInfo.ReplicaId
	if _, exist := pbft.rcPQCSenderStore[sender]; exist {
		logger.Warningf("Replica %d receive duplicate RecoveryReturnPQC, ignore it", pbft.id)
		return nil
	}
	pbft.rcPQCSenderStore[sender] = true

	if len(pbft.rcPQCSenderStore) > pbft.f+1 {
		logger.Debugf("Replica %d already receive %d returnPQC", pbft.id, len(pbft.rcPQCSenderStore))
		pbft.recoveryRestartTimer.Stop()
		pbft.inRecovery = false
	}

	prepreSet := PQCInfo.GetPrepreSet()
	preSent   := PQCInfo.PreSent
	cmtSent   := PQCInfo.CmtSent

	for i:=0; i<len(PQCInfo.PrepreSet); i++ {
		preprep := prepreSet[i]
		// recv preprepare
		cert := pbft.getCert(preprep.View, preprep.SequenceNumber)
		if cert.digest != preprep.BatchDigest {
			pbft.recvPrePrepare(preprep)
		}
		// recv prepare
		if preSent[i] {
			prep := &Prepare{
				View:			preprep.View,
				SequenceNumber: 	preprep.SequenceNumber,
				BatchDigest:		preprep.BatchDigest,
				ReplicaId:		sender,
			}
			pbft.recvPrepare(prep)
		}
		// recv commit
		if cmtSent[i] {
			cmt := &Commit{
				View:			preprep.View,
				SequenceNumber:		preprep.SequenceNumber,
				BatchDigest:		preprep.BatchDigest,
				ReplicaId:		sender,
			}
			pbft.recvCommit(cmt)
		}
	}

	return nil
}

// restartRecovery restart recovery immediately when recoveryRestartTimer expires
func (pbft *pbftProtocal) restartRecovery() {

	logger.Noticef("Replica %d now restartRecovery", pbft.id)

	pbft.initRecovery()
}

