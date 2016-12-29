//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"hyperchain/consensus/events"
	"github.com/golang/protobuf/proto"
)

type viewChangeQuorumEvent struct{}

func (pbft *pbftProtocal) correctViewChange(vc *ViewChange) bool {
	for _, p := range append(vc.Pset, vc.Qset...) {
		if !(p.View < vc.View && p.SequenceNumber > vc.H && p.SequenceNumber <= vc.H+pbft.L) {
			logger.Debugf("Replica %d invalid p entry in view-change: vc(v:%d h:%d) p(v:%d n:%d)", pbft.id, vc.View, vc.H, p.View, p.SequenceNumber)
			return false
		}
	}

	for _, c := range vc.Cset {
		if !(c.SequenceNumber >= vc.H && c.SequenceNumber <= vc.H+pbft.L) {
			logger.Debugf("Replica %d invalid c entry in view-change: vc(v:%d h:%d) c(n:%d)", pbft.id, vc.View, vc.H, c.SequenceNumber)
			return false
		}
	}

	return true
}

func (pbft *pbftProtocal) calcPSet() map[uint64]*ViewChange_PQ {
	pset := make(map[uint64]*ViewChange_PQ)

	for n, p := range pbft.pset {
		pset[n] = p
	}

	for idx, cert := range pbft.certStore {
		if cert.prePrepare == nil {
			continue
		}

		digest := cert.digest
		if !pbft.prepared(digest, idx.v, idx.n) {
			continue
		}

		if p, ok := pset[idx.n]; ok && p.View > idx.v {
			continue
		}

		pset[idx.n] = &ViewChange_PQ{
			SequenceNumber: idx.n,
			BatchDigest:    digest,
			View:           idx.v,
		}
	}

	return pset
}

func (pbft *pbftProtocal) calcQSet() map[qidx]*ViewChange_PQ {
	qset := make(map[qidx]*ViewChange_PQ)

	for n, q := range pbft.qset {
		qset[n] = q
	}

	for idx, cert := range pbft.certStore {
		if cert.prePrepare == nil {
			continue
		}

		digest := cert.digest
		if !pbft.prePrepared(digest, idx.v, idx.n) {
			continue
		}

		qi := qidx{digest, idx.n}
		if q, ok := qset[qi]; ok && q.View > idx.v {
			continue
		}

		qset[qi] = &ViewChange_PQ{
			SequenceNumber: idx.n,
			BatchDigest:    digest,
			View:           idx.v,
		}
	}

	return qset
}

func (pbft *pbftProtocal) sendViewChange() events.Event {

	if pbft.inNegoView {
		logger.Debugf("Replica %d try to send view change, but it's in nego-view", pbft.id)
		return nil
	}

	if pbft.inRecovery {
		logger.Noticef("Replica %d try to send view change, but it's in recovery", pbft.id)
		return nil
	}

	pbft.stopTimer()

	delete(pbft.newViewStore, pbft.view)
	pbft.view++
	pbft.activeView = false

	pbft.pset = pbft.calcPSet()
	pbft.qset = pbft.calcQSet()

	// clear old messages
	for idx := range pbft.certStore {
		if idx.v < pbft.view {
			delete(pbft.certStore, idx)
		}
	}
	for idx := range pbft.viewChangeStore {
		if idx.v < pbft.view {
			delete(pbft.viewChangeStore, idx)
		}
	}

	vc := &ViewChange {
		View:	pbft.view,
		H:	pbft.h,
		ReplicaId: pbft.id,
	}

	for n, id := range pbft.chkpts {
		vc.Cset = append(vc.Cset, &ViewChange_C {
			SequenceNumber: n,
			Id:		id,
		})
	}

	for _, p := range pbft.pset {
		if p.SequenceNumber < pbft.h {
			logger.Errorf("BUG! Replica %d should not have anything in our pset less than h, found %+v", pbft.id, p)
		}
		vc.Pset = append(vc.Pset, p)
	}

	for _, q := range pbft.qset {
		if q.SequenceNumber < pbft.h {
			logger.Errorf("BUG! Replica %d should not have anything in our qset less than h, found %+v", pbft.id, q)
		}
		vc.Qset = append(vc.Qset, q)
	}

	// TODO signature
	//pbft.sign(vc)

	logger.Infof("Replica %d sending view-change, v:%d, h:%d, |C|:%d, |P|:%d, |Q|:%d",
		pbft.id, vc.View, vc.H, len(vc.Cset), len(vc.Pset), len(vc.Qset))

	//todo
	payload, err := proto.Marshal(vc)
	if err != nil {
		logger.Errorf("ConsensusMessage_VIEW_CHANGE Marshal Error", err)
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:		ConsensusMessage_VIEW_CHANGE,
		Payload:	payload,
	}
	msg := consensusMsgHelper(consensusMsg, pbft.id)
	pbft.helper.InnerBroadcast(msg)
	pbft.vcResendTimer.Reset(pbft.vcResendTimeout, viewChangeResendTimerEvent{})
	return pbft.recvViewChange(vc)
}

func (pbft *pbftProtocal) recvViewChange(vc *ViewChange) events.Event {
	logger.Infof("Replica %d received view-change from replica %d, v:%d, h:%d, |C|:%d, |P|:%d, |Q|:%d",
		pbft.id, vc.ReplicaId, vc.View, vc.H, len(vc.Cset), len(vc.Pset), len(vc.Qset))

	if pbft.inNegoView {
		logger.Debugf("Replica %d try to recvViewChange, but it's in nego-view", pbft.id)
		return nil
	}

	if pbft.inRecovery {
		logger.Noticef("Replica %d try to recvcViewChange, but it's in recovery", pbft.id)
		return nil
	}

	// TODO verify
	//if err := pbft.verify(vc); err != nil {
	//	logger.Warningf("Replica %d found incorrect signature in view-change message: %s", pbft.id, err)
	//	return nil
	//}

	if vc.View < pbft.view {
		logger.Warningf("Replica %d found view-change message for old view", pbft.id)
		return nil
	}

	if !pbft.correctViewChange(vc) {
		logger.Warningf("Replica %d found view-change message incorrect", pbft.id)
		return nil
	}

	// record same vc from self times
	if vc.ReplicaId == pbft.id {
		pbft.vcResendCount++
		logger.Warningf("======== Replica %d already recv view change from itself for %d times", pbft.id, pbft.vcResendCount)
	}

	if _, ok := pbft.viewChangeStore[vcidx{vc.View, vc.ReplicaId}]; ok {
		logger.Warningf("Replica %d already has a view change message" +
			" for view %d from replica %d", pbft.id, vc.View, vc.ReplicaId)

		if pbft.vcResendCount >= pbft.vcResendLimit {
			logger.Noticef("Replica %d view change resend reach upbound, try to recovery", pbft.id)
			pbft.vcResendTimer.Stop()
			pbft.vcResendCount = 0
			pbft.inNegoView = true
			pbft.inRecovery = true
			pbft.activeView = true
			pbft.processNegotiateView()
		}

		return nil
	}

	pbft.viewChangeStore[vcidx{vc.View, vc.ReplicaId}] = vc

	// PBFT TOCS 4.5.1 Liveness: "if a replica receives a set of
	// f+1 valid VIEW-CHANGE messages from other replicas for
	// views greater than its current view, it sends a VIEW-CHANGE
	// message for the smallest view in the set, even if its timer
	// has not expired"
	replicas := make(map[uint64]bool)
	minView := uint64(0)
	for idx := range pbft.viewChangeStore {
		if idx.v <= pbft.view {
			continue
		}

		replicas[idx.id] = true
		if minView == 0 || idx.v < minView {
			minView = idx.v
		}
	}

	// We only enter this if there are enough view change messages _greater_ than our current view
	if len(replicas) >= pbft.f+1 {
		logger.Warningf("Replica %d received f+1 view-change messages, triggering view-change to view %d",
			pbft.id, minView)
		pbft.firstRequestTimer.Stop()
		// subtract one, because sendViewChange() increments
		pbft.view = minView - 1
		return pbft.sendViewChange()
	}

	quorum := 0
	for idx := range pbft.viewChangeStore {
		if idx.v == pbft.view {
			quorum++
		}
	}
	logger.Debugf("Replica %d now has %d view change requests for view %d", pbft.id, quorum, pbft.view)

	if !pbft.activeView && vc.View == pbft.view && quorum >= pbft.allCorrectReplicasQuorum() {
		pbft.vcResendTimer.Stop()
		// TODO first param
		pbft.startTimer(pbft.lastNewViewTimeout, "new view change")
		pbft.lastNewViewTimeout = 2 * pbft.lastNewViewTimeout
		return viewChangeQuorumEvent{}
	}

	return nil
}

func (pbft *pbftProtocal) sendNewView() events.Event {

	if pbft.inNegoView {
		logger.Debugf("Replica %d try to sendNewView, but it's in nego-view", pbft.id)
		return nil
	}

	if _, ok := pbft.newViewStore[pbft.view]; ok {
		logger.Debugf("Replica %d already has new view in store for view %d, skipping", pbft.id, pbft.view)
		return nil
	}

	vset := pbft.getViewChanges()

	cp, ok, replicas := pbft.selectInitialCheckpoint(vset)

	if !ok {
		logger.Infof("Replica %d could not find consistent checkpoint: %+v", pbft.id, pbft.viewChangeStore)
		return nil
	}

	msgList := pbft.assignSequenceNumbers(vset, cp.SequenceNumber)
	if msgList == nil {
		logger.Infof("Replica %d could not assign sequence numbers for new view", pbft.id)
		return nil
	}

	nv := &NewView{
		View:      pbft.view,
		Vset:      vset,
		Xset:      msgList,
		ReplicaId: pbft.id,
	}

	logger.Infof("Replica %d is new primary, sending new-view, v:%d, X:%+v",
		pbft.id, nv.View, nv.Xset)
	payload, err := proto.Marshal(nv)
	if err != nil {
		logger.Errorf("ConsensusMessage_NEW_VIEW Marshal Error", err)
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:		ConsensusMessage_NEW_VIEW,
		Payload:	payload,
	}
	msg := consensusMsgHelper(consensusMsg, pbft.id)
	pbft.helper.InnerBroadcast(msg)
	pbft.newViewStore[pbft.view] = nv
	//return pbft.processNewView()
	return pbft.primaryProcessNewView(cp, replicas, nv)
}

func (pbft *pbftProtocal) recvNewView(nv *NewView) events.Event {
	logger.Infof("Replica %d received new-view %d",
		pbft.id, nv.View)

	if pbft.inNegoView {
		logger.Debugf("Replica %d try to recvNewView, but it's in nego-view", pbft.id)
		return nil
	}

	if pbft.inRecovery {
		logger.Noticef("Replica %d try to recvNewView, but it's in recovery", pbft.id)
		pbft.recvNewViewInRecovery = true
		return nil
	}

	if !(nv.View > 0 && nv.View >= pbft.view && pbft.primary(nv.View) == nv.ReplicaId && pbft.newViewStore[nv.View] == nil) {
		logger.Infof("Replica %d rejecting invalid new-view from %d, v:%d",
			pbft.id, nv.ReplicaId, nv.View)
		return nil
	}

	pbft.newViewStore[nv.View] = nv
	return pbft.processNewView()
}

func (pbft *pbftProtocal) canExecuteToTarget(specLastExec uint64, initialCp ViewChange_C) bool {

	canExecuteToTarget := true
	outer:
	for seqNo := specLastExec + 1; seqNo <= initialCp.SequenceNumber; seqNo++ {
		found := false
		for idx, cert := range pbft.certStore {
			if idx.n != seqNo {
				continue
			}

			quorum := 0
			for p := range cert.commit {
				// Was this committed in the previous view
				if p.View == idx.v && p.SequenceNumber == seqNo {
					quorum++
				}
			}

			if quorum < pbft.intersectionQuorum() {
				logger.Debugf("Replica %d missing quorum of commit certificate for seqNo=%d, only has %d of %d", pbft.id, quorum, pbft.intersectionQuorum())
				continue
			}

			found = true
			break
		}

		if !found {
			canExecuteToTarget = false
			logger.Debugf("Replica %d missing commit certificate for seqNo=%d", pbft.id, seqNo)
			break outer
		}

	}

	if canExecuteToTarget {
		pbft.nvInitialSeqNo = initialCp.SequenceNumber
		logger.Debugf("Replica %d needs to process a new view, but can execute to the checkpoint seqNo %d, delaying processing of new view", pbft.id, initialCp.SequenceNumber)
	} else {
		pbft.nvInitialSeqNo = 0
		logger.Infof("Replica %d cannot execute to the view change checkpoint with seqNo %d", pbft.id, initialCp.SequenceNumber)
	}
	return  canExecuteToTarget
}

func (pbft *pbftProtocal) feedMissingReqBatchIfNeeded(nv *NewView) (newReqBatchMissing bool) {
	newReqBatchMissing = false
	for n, d := range nv.Xset {
		// PBFT: why should we use "h ≥ min{n | ∃d : (<n,d> ∈ X)}"?
		// "h ≥ min{n | ∃d : (<n,d> ∈ X)} ∧ ∀<n,d> ∈ X : (n ≤ h ∨ ∃m ∈ in : (D(m) = d))"
		if n <= pbft.h {
			continue
		} else {
			if d == "" {
				// NULL request; skip
				continue
			}


			if _, ok := pbft.validatedBatchStore[d]; !ok {
				logger.Warningf("Replica %d missing assigned, non-checkpointed request batch %s",
					pbft.id, d)
				if _, ok := pbft.missingReqBatches[d]; !ok {
					logger.Warningf("Replica %v requesting to fetch batch %s",
						pbft.id, d)
					newReqBatchMissing = true
					pbft.missingReqBatches[d] = true
				}
			}
		}
	}
	return newReqBatchMissing
}

func (pbft *pbftProtocal) primaryProcessNewView(initialCp ViewChange_C, replicas []uint64, nv *NewView) events.Event {
	var newReqBatchMissing bool

	speculativeLastExec := pbft.lastExec
	if pbft.currentExec != nil {
		speculativeLastExec = *pbft.currentExec
	}
	// If we have not reached the sequence number, check to see if we can reach it without state transfer
	// In general, executions are better than state transfer
	if speculativeLastExec < initialCp.SequenceNumber {
		if pbft.canExecuteToTarget(speculativeLastExec, initialCp) {
			return nil
		}
	}

	if pbft.h < initialCp.SequenceNumber {
		pbft.moveWatermarks(initialCp.SequenceNumber)
	}

	// true means we can not execToTarget need state transfer
	if speculativeLastExec < initialCp.SequenceNumber {
		logger.Warningf("Replica %d missing base checkpoint %d (%s), our most recent execution %d", pbft.id, initialCp.SequenceNumber, initialCp.Id, speculativeLastExec)

		snapshotID, err := base64.StdEncoding.DecodeString(initialCp.Id)
		if nil != err {
			err = fmt.Errorf("Replica %d received a view change whose hash could not be decoded (%s)", pbft.id, initialCp.Id)
			logger.Error(err.Error())
			return nil
		}

		target := &stateUpdateTarget{
			checkpointMessage: checkpointMessage{
				seqNo: initialCp.SequenceNumber,
				id:    snapshotID,
			},
			replicas: replicas,
		}

		pbft.updateHighStateTarget(target)
		pbft.stateTransfer(target)
	}

	newReqBatchMissing = pbft.feedMissingReqBatchIfNeeded(nv)

	if len(pbft.missingReqBatches) == 0 {
		return pbft.processReqInNewView(nv)
	} else if newReqBatchMissing {
		pbft.fetchRequestBatches()
	}

	return nil
}


func (pbft *pbftProtocal) processNewView() events.Event {
	var newReqBatchMissing bool
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

	cp, ok, replicas := pbft.selectInitialCheckpoint(nv.Vset)
	if !ok {
		logger.Warningf("Replica %d could not determine initial checkpoint: %+v",
			pbft.id, pbft.viewChangeStore)
		return pbft.sendViewChange()
	}
// 以上 primary 不必做
	speculativeLastExec := pbft.lastExec
	if pbft.currentExec != nil {
		speculativeLastExec = *pbft.currentExec
	}

	// If we have not reached the sequence number, check to see if we can reach it without state transfer
	// In general, executions are better than state transfer
	if speculativeLastExec < cp.SequenceNumber {
		if speculativeLastExec < cp.SequenceNumber {
			if pbft.canExecuteToTarget(speculativeLastExec, cp) {
				return nil
			}
		}
	}
// --
	msgList := pbft.assignSequenceNumbers(nv.Vset, cp.SequenceNumber)

	if msgList == nil {
		logger.Warningf("Replica %d could not assign sequence numbers: %+v",
			pbft.id, pbft.viewChangeStore)
		return pbft.sendViewChange()
	}

	if !(len(msgList) == 0 && len(nv.Xset) == 0) && !reflect.DeepEqual(msgList, nv.Xset) {
		logger.Warningf("Replica %d failed to verify new-view Xset: computed %+v, received %+v",
			pbft.id, msgList, nv.Xset)
		return pbft.sendViewChange()
	}
// -- primary 不必做
	if pbft.h < cp.SequenceNumber {
		pbft.moveWatermarks(cp.SequenceNumber)
	}

	if speculativeLastExec < cp.SequenceNumber {
		logger.Warningf("Replica %d missing base checkpoint %d (%s), our most recent execution %d", pbft.id, cp.SequenceNumber, cp.Id, speculativeLastExec)

		snapshotID, err := base64.StdEncoding.DecodeString(cp.Id)
		if nil != err {
			err = fmt.Errorf("Replica %d received a view change whose hash could not be decoded (%s)", pbft.id, cp.Id)
			logger.Error(err.Error())
			return nil
		}

		target := &stateUpdateTarget{
			checkpointMessage: checkpointMessage{
				seqNo: cp.SequenceNumber,
				id:    snapshotID,
			},
			replicas: replicas,
		}

		pbft.updateHighStateTarget(target)
		pbft.stateTransfer(target)
	}

	//TODO: 从节点不需要拿batch,只要更新状态信息就行
	newReqBatchMissing = pbft.feedMissingReqBatchIfNeeded(nv)
	if len(pbft.missingReqBatches) == 0 {
		return pbft.processReqInNewView(nv)
	} else if newReqBatchMissing {
		pbft.fetchRequestBatches()
	}

	return nil
}

func (pbft *pbftProtocal) processReqInNewView(nv *NewView) events.Event {
	logger.Debugf("Replica %d accepting new-view to view %d", pbft.id, pbft.view)

	pbft.stopTimer()
	pbft.nullRequestTimer.Stop()

	delete(pbft.newViewStore, pbft.view-1)
	// empty the outstandingReqBatch, it is useless since new primary will resend pre-prepare
	pbft.outstandingReqBatches = make(map[string]*TransactionBatch)
	pbft.lastExec = pbft.h
	pbft.seqNo = pbft.h
	pbft.vid = pbft.h
	pbft.lastVid = pbft.h
	if !pbft.skipInProgress {
		backendVid := uint64(pbft.vid+1)
		pbft.helper.VcReset(backendVid)
	}
	xSetLen := len(nv.Xset)
	upper := uint64(xSetLen) + pbft.h + uint64(1)
	if pbft.primary(pbft.view) == pbft.id {
		for i := pbft.h+uint64(1); i < upper; i++ {
			d, ok := nv.Xset[i]
			if !ok {
				logger.Critical("view change Xset miss batch number %d", i)
			} else if d == "" {
				// This should not happen
				logger.Critical("view change Xset has null batch, kick it out")
			} else {
				batch, ok := pbft.validatedBatchStore[d]
				if !ok {
					logger.Criticalf("In Xset %s exists, but in Replica %d validatedBatchStore there is no such batch digest", d, pbft.id)
				} else {
					logger.Critical("send validate")
					digest := hash(batch)
					pbft.softStartTimer(pbft.requestTimeout, fmt.Sprintf("new request batch %s", digest))
					pbft.primaryValidateBatch(batch)
				}
			}
		}
	}

	pbft.updateViewChangeSeqNo()
	pbft.startTimerIfOutstandingRequests()
	logger.Debugf("Replica %d done cleaning view change artifacts, calling into consumer", pbft.id)

	return viewChangedEvent{}
}

func (pbft *pbftProtocal) getViewChanges() (vset []*ViewChange) {
	for _, vc := range pbft.viewChangeStore {
		vset = append(vset, vc)
	}
	return
}

func (pbft *pbftProtocal) selectInitialCheckpoint(vset []*ViewChange) (checkpoint ViewChange_C, ok bool, replicas []uint64) {
	checkpoints := make(map[ViewChange_C][]*ViewChange)
	for _, vc := range vset {
		for _, c := range vc.Cset { // TODO, verify that we strip duplicate checkpoints from this set
			checkpoints[*c] = append(checkpoints[*c], vc)
			logger.Debugf("Replica %d appending checkpoint from replica %d with seqNo=%d, h=%d, and checkpoint digest %s", pbft.id, vc.ReplicaId, vc.H, c.SequenceNumber, c.Id)
		}
	}

	if len(checkpoints) == 0 {
		logger.Debugf("Replica %d has no checkpoints to select from: %d %s",
			pbft.id, len(pbft.viewChangeStore), checkpoints)
		return
	}

	for idx, vcList := range checkpoints {
		// need weak certificate for the checkpoint
		if len(vcList) <= pbft.f { // type casting necessary to match types
			logger.Debugf("Replica %d has no weak certificate for n:%d, vcList was %d long",
				pbft.id, idx.SequenceNumber, len(vcList))
			continue
		}

		quorum := 0
		// Note, this is the whole vset (S) in the paper, not just this checkpoint set (S') (vcList)
		// We need 2f+1 low watermarks from S below this seqNo from all replicas
		// We need f+1 matching checkpoints at this seqNo (S')
		for _, vc := range vset {
			if vc.H <= idx.SequenceNumber {
				quorum++
			}
		}

		if quorum < pbft.intersectionQuorum() {
			logger.Debugf("Replica %d has no quorum for n:%d", pbft.id, idx.SequenceNumber)
			continue
		}

		if checkpoint.SequenceNumber <= idx.SequenceNumber {
			replicas = make([]uint64, len(vcList))
			for i, vc := range vcList {
				replicas[i] = vc.ReplicaId
			}

			checkpoint = idx
			ok = true
		}
	}

	return
}

func (pbft *pbftProtocal) assignSequenceNumbers(vset []*ViewChange, h uint64) (msgList map[uint64]string) {
	msgList = make(map[uint64]string)

	maxN := h + 1

	// "for all n such that h < n <= h + L"
	nLoop:
	for n := h + 1; n <= h+pbft.L; n++ {
		// "∃m ∈ S..."
		for _, m := range vset {
			// "...with <n,d,v> ∈ m.P"
			for _, em := range m.Pset {
				quorum := 0
				// "A1. ∃2f+1 messages m' ∈ S"
				mpLoop:
				for _, mp := range vset {
					if mp.H >= n {
						continue
					}
					// "∀<n,d',v'> ∈ m'.P"
					for _, emp := range mp.Pset {
						if n == emp.SequenceNumber && !(emp.View < em.View || (emp.View == em.View && emp.BatchDigest == em.BatchDigest)) {
							continue mpLoop
						}
					}
					quorum++
				}

				if quorum < pbft.intersectionQuorum() {
					continue
				}

				quorum = 0
				// "A2. ∃f+1 messages m' ∈ S"
				for _, mp := range vset {
					// "∃<n,d',v'> ∈ m'.Q"
					for _, emp := range mp.Qset {
						if n == emp.SequenceNumber && emp.View >= em.View && emp.BatchDigest == em.BatchDigest {
							quorum++
						}
					}
				}

				if quorum < pbft.f+1 {
					continue
				}

				// "then select the request with digest d for number n"
				msgList[n] = em.BatchDigest
				maxN = n

				continue nLoop
			}
		}

		quorum := 0
		// "else if ∃2f+1 messages m ∈ S"
		nullLoop:
		for _, m := range vset {
			// "m.P has no entry"
			for _, em := range m.Pset {
				if em.SequenceNumber == n {
					continue nullLoop
				}
			}
			quorum++
		}

		if quorum >= pbft.intersectionQuorum() {
			// "then select the null request for number n"
			msgList[n] = ""

			continue nLoop
		}

		logger.Warningf("Replica %d could not assign value to contents of seqNo %d, found only %d missing P entries", pbft.id, n, quorum)
		return nil
	}

	// prune top null requests
	for n, msg := range msgList {
		if n > maxN && msg == "" {
			delete(msgList, n)
		}
	}

	return
}