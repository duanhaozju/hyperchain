//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"encoding/base64"

	"hyperchain/consensus/events"
	"hyperchain/consensus/helper/persist"

	"github.com/golang/protobuf/proto"
)

type blkIdx struct {
	height uint64
	hash   string
}

// procativeRecovery broadcast a procative recovery message to ask others for recent blocks info
func (pbft *pbftProtocal) initRecovery() events.Event {

	logger.Debugf("Replica %d now initRecovery", pbft.id)

	// update watermarks
	height := persist.GetHeightofChain()
	pbft.moveWatermarks(height)

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


	height, curHash := persist.GetBlockHeightAndHash()

	rc := &RecoveryResponse{
		ReplicaId:	 pbft.id,
		Chkpts:		 chkpts,
		BlockHeight:     height,
		LastBlockHash:   curHash,
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

	if len(pbft.rcRspStore) <= pbft.N-pbft.f {
		logger.Debugf("Replica %d recv recoveryRsp from replica %d, rsp count: %d, not " +
			"beyond %d", pbft.id, rsp.ReplicaId, len(pbft.rcRspStore), pbft.N-pbft.f)
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

	if pbft.recoveryToSeqNo != nil {
		logger.Debugf("Replica %d in recovery receive rcRsp from replica %d" +
			"but chkpt quorum and seqNo quorum already found. " +
			"Ignore it", pbft.id, rsp.ReplicaId)
		return nil
	}

	pbft.recoveryRestartTimer.Stop()
	pbft.recoveryToSeqNo = &lastExec

	//blockInfo := getBlockchainInfo()
	//id, _ := proto.Marshal(blockInfo)
	//idAsString := byteToString(id)
	selfLastExec, selfCurHash := persist.GetBlockHeightAndHash()


	logger.Debugf("Replica %d in recovery find quorum chkpt: %d, self: %d, " +
		"others lastExec: %d, self: %d", pbft.id, n, pbft.h, lastExec, pbft.lastExec)
	logger.Debugf("Replica %d in recovery, " +
		"others lastBlockInfo: %s, self: %s", pbft.id, rsp.BlockHeight, selfCurHash)

	// Fast catch up
	if lastExec == selfLastExec && curHash == selfCurHash {
		logger.Debugf("Replica %d in recovery same lastExec: %d, " +
			"same block hash: %s, fast catch up", pbft.id, selfLastExec, curHash)
		pbft.inRecovery = false
		return recoveryDoneEvent{}
	}

	logger.Debugf("Replica %d in recovery self lastExec: %d, others: %d" +
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
	} else {
		logger.Critical("send stateupdated")
		pbft.helper.VcReset(n+1)
		state := &stateUpdatedEvent{seqNo: n}
		go pbft.postPbftEvent(state)
		return nil
	}
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

	// Since replica sends all of its chkpt, we may encounter several, instead of single one,
	// chkpts, which reach 2f+1

	// In this case, others will move watermarks sooner or later.
	// Hopefully, we find only one chkpt which reaches 2f+1 and this chkpt is their pbft.h
	for ci, peers := range chkpts {
		if len(peers) >= 2*pbft.f+1 {
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

func (pbft *pbftProtocal) findLastExecQuorum() (lastExec uint64, hash string, find bool) {

	lastExecs := make(map[blkIdx]map[uint64]bool)
	find = false
	for _, rsp := range pbft.rcRspStore {
		idx := blkIdx{
			height:rsp.BlockHeight,
			hash:rsp.LastBlockHash,
		}
		replicas, ok := lastExecs[idx]
		if ok {
			replicas[rsp.ReplicaId] = true
		} else {
			replicas := make(map[uint64]bool)
			replicas[rsp.ReplicaId] = true
			lastExecs[idx] = replicas
		}

		if len(lastExecs[idx]) >= 2*pbft.f+1 {
			lastExec = idx.height
			hash = idx.hash
			find = true
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

	logger.Debugf("Replica %d now returnRecoveryPQC", pbft.id)

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
	}

	prepreSet := PQCInfo.GetPrepreSet()
	preSent   := PQCInfo.PreSent
	cmtSent   := PQCInfo.CmtSent

	for i:=0; i<len(PQCInfo.PrepreSet); i++ {
		preprep := prepreSet[i]
		// recv preprepare
		cert := pbft.getCert(preprep.View, preprep.SequenceNumber)
		if cert.digest != preprep.BatchDigest {
			go pbft.postPbftEvent(preprep)
		}
		// recv prepare
		if preSent[i] {
			prep := &Prepare{
				View:			preprep.View,
				SequenceNumber: 	preprep.SequenceNumber,
				BatchDigest:		preprep.BatchDigest,
				ReplicaId:		sender,
			}
			go pbft.postPbftEvent(prep)
		}
		// recv commit
		if cmtSent[i] {
			cmt := &Commit{
				View:			preprep.View,
				SequenceNumber:		preprep.SequenceNumber,
				BatchDigest:		preprep.BatchDigest,
				ReplicaId:		sender,
			}
			go pbft.postPbftEvent(cmt)
		}
	}

	return nil
}

// restartRecovery restart recovery immediately when recoveryRestartTimer expires
func (pbft *pbftProtocal) restartRecovery() {

	logger.Noticef("Replica %d now restartRecovery", pbft.id)

	pbft.initRecovery()
}

