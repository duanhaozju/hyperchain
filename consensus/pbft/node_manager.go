//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"sync/atomic"

	"hyperchain/protos"

	"github.com/golang/protobuf/proto"
	"hyperchain/consensus/events"
)

// New replica receive local NewNode message
func (pbft *pbftProtocal) recvLocalNewNode(msg *protos.NewNodeMessage) error {

	logger.Debugf("New replica %d received local newNode message", pbft.id)

	if pbft.isNewNode {
		logger.Warningf("New replica %d received duplicate local newNode message", pbft.id)
		return nil
	}

	if len(msg.Payload) == 0 {
		logger.Warningf("New replica %d received nil local newNode message", pbft.id)
		return nil
	}

	pbft.isNewNode = true
	pbft.inAddingNode = true
	key := string(msg.Payload)
	pbft.localKey = key

	return nil
}

// Replica receive local message about new node and routing table
func (pbft *pbftProtocal) recvLocalAddNode(msg *protos.AddNodeMessage) error {

	if pbft.isNewNode {
		logger.Warningf("New replica received local addNode message, there may be something wrong")
		return nil
	}

	if len(msg.Payload) == 0 {
		logger.Warningf("New replica %d received nil local addNode message", pbft.id)
		return nil
	}

	key := string(msg.Payload)
	logger.Debugf("Replica %d received local addNode message for new node %v", pbft.id, key)

	pbft.inAddingNode = true
	pbft.sendAgreeAddNode(key)

	return nil
}

// Replica receive local message about new node and routing table
func (pbft *pbftProtocal) recvLocalDelNode(msg *protos.DelNodeMessage) error {

	key := string(msg.DelPayload)
	logger.Debugf("Replica %d received local delnode message for del node %s", pbft.id, key)

	if pbft.N == 4 {
		logger.Criticalf("Replica %d receive del msg, but we don't support delete as there're only 4 nodes")
		return nil
	}

	if len(msg.DelPayload) == 0 || len(msg.RouterHash) == 0 || msg.Id == 0 {
		logger.Warningf("New replica %d received invalid local delNode message", pbft.id)
		return nil
	}

	pbft.inDeletingNode = true
	pbft.sendAgreeDelNode(key, msg.RouterHash, msg.Id)

	return nil
}

// Repica broadcast addnode message for new node
func (pbft *pbftProtocal) sendAgreeAddNode(key string) {

	logger.Debugf("Replica %d try to send addnode message for new node", pbft.id)

	if pbft.isNewNode {
		logger.Warningf("New replica try to send addnode message, there may be something wrong")
		return
	}

	add := &AddNode{
		ReplicaId:	pbft.id,
		Key:		key,
	}

	payload, err := proto.Marshal(add)
	if err != nil {
		logger.Errorf("Marshal AddNode Error!")
		return
	}
	msg := &ConsensusMessage{
		Type: ConsensusMessage_ADD_NODE,
		Payload: payload,
	}

	broadcast := consensusMsgHelper(msg, pbft.id)
	pbft.helper.BroadcastAddNode(broadcast)
	pbft.recvAgreeAddNode(add)

}

// Repica broadcast delnode message for quit node
func (pbft *pbftProtocal) sendAgreeDelNode(key string, routerHash string, newId uint64) {

	logger.Debugf("Replica %d try to send delnode message for quit node", pbft.id)

	cert := pbft.getDelNodeCert(key)
	cert.newId = newId
	cert.routerHash = routerHash

	del := &DelNode{
		ReplicaId:	pbft.id,
		Key:		key,
		RouterHash:	routerHash,
	}

	payload, err := proto.Marshal(del)
	if err != nil {
		logger.Errorf("Marshal DelNode Error!")
		return
	}
	msg := &ConsensusMessage{
		Type: ConsensusMessage_DEL_NODE,
		Payload: payload,
	}

	broadcast := consensusMsgHelper(msg, pbft.id)
	pbft.helper.BroadcastDelNode(broadcast)
	pbft.recvAgreeDelNode(del)

}

// Replica received addnode for new node
func (pbft *pbftProtocal) recvAgreeAddNode(add *AddNode) error {

	logger.Debugf("Replica %d received addnode from replica %d for %v",
		pbft.id, add.ReplicaId, add.Key)

	cert := pbft.getAddNodeCert(add.Key)

	ok := cert.addNodes[*add]
	if ok {
		logger.Warningf("Replica %d ignored duplicate addnode from %d", pbft.id, add.ReplicaId)
		return nil
	}

	cert.addNodes[*add] = true
	cert.addCount++

	return pbft.maybeUpdateTableForAdd(add.Key)
}

// Replica received delnode for quit node
func (pbft *pbftProtocal) recvAgreeDelNode(del *DelNode) error {

	logger.Debugf("Replica %d received agree delnode from replica %d for %v",
		pbft.id, del.ReplicaId, del.Key)

	cert := pbft.getDelNodeCert(del.Key)

	ok := cert.delNodes[*del]
	if ok {
		logger.Warningf("Replica %d ignored duplicate agree addnode from %d", pbft.id, del.ReplicaId)
		return nil
	}

	cert.delNodes[*del] = true
	cert.delCount++

	return pbft.maybeUpdateTableForDel(del.Key)
}

// Check if replica prepared for update routing table after add node
func (pbft *pbftProtocal) maybeUpdateTableForAdd(key string) error {

	cert := pbft.getAddNodeCert(key)

	if cert.addCount < pbft.committedReplicasQuorum() {
		return nil
	}

	if !pbft.inAddingNode {
		if cert.finishAdd {
			if cert.addCount <= pbft.N {
				logger.Debugf("Replica %d has already finished adding node", pbft.id)
				return nil
			} else {
				logger.Warningf("Replica %d has already finished adding node, but still recevice add msg from someone else", pbft.id)
				return nil
			}
		}
	}

	cert.finishAdd = true
	payload := []byte(key)

	pbft.helper.UpdateTable(payload, true)
	pbft.inAddingNode = false

	return nil
}

// Check if replica prepared for update routing table after del node
func (pbft *pbftProtocal) maybeUpdateTableForDel(key string) error {

	cert := pbft.getDelNodeCert(key)

	if cert.delCount < pbft.committedReplicasQuorum() {
		return nil
	}

	if !pbft.inDeletingNode {
		if cert.finishDel {
			if cert.delCount < pbft.N {
				logger.Debugf("Replica %d have already finished deleting node", pbft.id)
				return nil
			} else {
				logger.Warningf("Replica %d has already finished deleting node, but still recevice del msg from someone else", pbft.id)
				return nil
			}
		}
	}

	cert.finishDel = true
	payload := []byte(key)

	logger.Debugf("Replica %d try to update routing table", pbft.id)
	pbft.helper.UpdateTable(payload, false)
	pbft.inDeletingNode = false

	pbft.inUpdatingN = true
	pbft.sendAgreeUpdateNforDel(key, cert.routerHash)

	return nil
}

// New replica send ready_for_n to all replicas after recovery
func (pbft *pbftProtocal) sendReadyForN() error {

	if !pbft.isNewNode {
		logger.Errorf("Replica %d is not new one, but try to send ready_for_n", pbft.id)
		return nil
	}

	if pbft.localKey == "" {
		logger.Errorf("Replica %d doesn't have local key for ready_for_n", pbft.id)
		return nil
	}

	ready := &ReadyForN{
		ReplicaId:	pbft.id,
		Key:		pbft.localKey,
	}

	payload, err := proto.Marshal(ready)
	if err != nil {
		logger.Errorf("Marshal ReadyForN Error!")
		return nil
	}
	msg := &ConsensusMessage{
		Type: ConsensusMessage_READY_FOR_N,
		Payload: payload,
	}

	broadcast := consensusMsgHelper(msg, pbft.id)
	pbft.helper.InnerBroadcast(broadcast)

	return nil
}

// Primary receive ready_for_n from new replica
func (pbft *pbftProtocal) recvReadyforNforAdd(ready *ReadyForN) error {

	if active := atomic.LoadUint32(&pbft.activeView); active == 0 {
		logger.Warningf("Primary %d is in view change, reject the ready_for_n message", pbft.id)
		return nil
	}

	cert := pbft.getAddNodeCert(ready.Key)

	if !cert.finishAdd {
		logger.Errorf("Primary %d has not done with addnode for key=%s", pbft.id, ready.Key)
		return nil
	}

	// calculate the new N and view
	n, view := pbft.getAddNV()

	// broadcast the updateN message
	agree := &AgreeUpdateN{
		Flag:		true,
		ReplicaId:	pbft.id,
		Key:		ready.Key,
		N:			n,
		View:		view,
		H:			pbft.h,
	}

	pbft.agreeUpdateHelper(agree)
	logger.Infof("Replica %d sending update-N-View, v:%d, h:%d, |C|:%d, |P|:%d, |Q|:%d",
		pbft.id, agree.View, agree.H, len(agree.Cset), len(agree.Pset), len(agree.Qset))

	payload, err := proto.Marshal(agree)
	if err != nil {
		logger.Errorf("ConsensusMessage_AGREE_UPDATE_N Marshal Error", err)
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:		ConsensusMessage_AGREE_UPDATE_N,
		Payload:	payload,
	}
	msg := consensusMsgHelper(consensusMsg, pbft.id)
	pbft.helper.InnerBroadcast(msg)
	return pbft.recvAgreeUpdateN(agree)
}

// Primary send update_n after finish del node
func (pbft *pbftProtocal) sendAgreeUpdateNforDel(key string, routerHash string) error {

	logger.Debugf("Replica %d try to send update_n after finish del node", pbft.id)

	if atomic.LoadUint32(&pbft.activeView) == 0 {
		logger.Warningf("Primary %d is in view change, reject the ready_for_n message", pbft.id)
		return nil
	}

	cert := pbft.getDelNodeCert(key)

	if !cert.finishDel {
		logger.Errorf("Primary %d has not done with delnode for key=%s", pbft.id, key)
		return nil
	}

	pbft.stopTimer()
	atomic.StoreUint32(&pbft.inUpdatingN, 0)

	// calculate the new N and view
	n, view := pbft.getDelNV()

	agree := &AgreeUpdateN{
		Flag:		false,
		ReplicaId:	pbft.id,
		Key:		key,
		RouterHash:	routerHash,
		N:			n,
		View:		view,
		H:			pbft.h,
	}

	pbft.agreeUpdateHelper(agree)
	logger.Infof("Replica %d sending update-N-View, v:%d, h:%d, |C|:%d, |P|:%d, |Q|:%d",
		pbft.id, agree.View, agree.H, len(agree.Cset), len(agree.Pset), len(agree.Qset))

	payload, err := proto.Marshal(agree)
	if err != nil {
		logger.Errorf("ConsensusMessage_AGREE_UPDATE_N Marshal Error", err)
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:		ConsensusMessage_AGREE_UPDATE_N,
		Payload:	payload,
	}
	msg := consensusMsgHelper(consensusMsg, pbft.id)
	pbft.helper.InnerBroadcast(msg)
	return pbft.recvAgreeUpdateN(agree)
}

func (pbft *pbftProtocal) agreeUpdateHelper(agree *AgreeUpdateN) {

	// clear old messages
	for idx := range pbft.certStore {
		if idx.v < pbft.view {
			delete(pbft.certStore, idx)
		}
	}
	for idx := range pbft.agreeUpdateStore {
		if idx.v < pbft.view {
			delete(pbft.agreeUpdateStore, idx)
		}
	}

	pset := pbft.calcPSet()
	qset := pbft.calcQSet()

	for n, id := range pbft.chkpts {
		agree.Cset = append(agree.Cset, &ViewChange_C {
			SequenceNumber: n,
			Id:		id,
		})
	}

	for _, p := range pset {
		if p.SequenceNumber < pbft.h {
			logger.Errorf("BUG! Replica %d should not have anything in our pset less than h, found %+v", pbft.id, p)
		}
		agree.Pset = append(agree.Pset, p)
	}

	for _, q := range qset {
		if q.SequenceNumber < pbft.h {
			logger.Errorf("BUG! Replica %d should not have anything in our qset less than h, found %+v", pbft.id, q)
		}
		agree.Qset = append(agree.Qset, q)
	}
}

func (pbft *pbftProtocal) recvAgreeUpdateN(agree *AgreeUpdateN) events.Event {

	logger.Debugf("Replica %d received view-change from replica %d, v:%d, h:%d, |C|:%d, |P|:%d, |Q|:%d",
		pbft.id, agree.ReplicaId, agree.View, agree.H, len(agree.Cset), len(agree.Pset), len(agree.Qset))

	if pbft.activeView {
		logger.Warningf("Replica %d try to recvAgreeUpdateN, but it's in view-change", pbft.id)
		return nil
	}

	if pbft.inNegoView {
		logger.Warningf("Replica %d try to recvAgreeUpdateN, but it's in nego-view", pbft.id)
		return nil
	}

	if pbft.inRecovery {
		logger.Warningf("Replica %d try to recvAgreeUpdateN, but it's in recovery", pbft.id)
		return nil
	}

	if !pbft.correctViewChange(agree) {
		logger.Warningf("Replica %d found view-change message incorrect", pbft.id)
		return nil
	}

	key := aidx{
		v:agree.View,
		n:agree.N,
		id:agree.ReplicaId,
	}
	if _, ok := pbft.agreeUpdateStore[key]; ok {
		logger.Warningf("Replica %d already has a view change message" +
			" for view=%d/n=%d from replica %d", pbft.id, agree.View, agree.N, agree.ReplicaId)
		return nil
	}

	pbft.agreeUpdateStore[key] = agree

	replicas := make(map[uint64]bool)
	minView := uint64(0)
	for idx := range pbft.agreeUpdateStore {
		if idx.v < pbft.view {
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
	logger.Debugf("Replica %d now has %d agree-update requests for view=%d/n=%d", pbft.id, quorum, pbft.view, agree.N)

	if atomic.LoadUint32(&pbft.activeView) == 0 && quorum >= pbft.allCorrectReplicasQuorum() {
		return viewChangeQuorumEvent{}
	}

	return nil
}

func (pbft *pbftProtocal) checkAgreement(agree *AgreeUpdateN) bool {

	if agree.Flag {
		cert := pbft.getAddNodeCert(agree.Key)
		if !cert.finishAdd {
			logger.Warningf("Replica %d has not complete add node")
			return false
		}
		n, view := pbft.getAddNV()
		if n != agree.N || view != agree.View {
			logger.Warningf("Replica %d invalid p entry in agree-update: expected n=%d/view=%d, get n=%d/view=%d", pbft.id, n, view, agree.N, agree.View)
			return false
		}
	} else {
		cert := pbft.getDelNodeCert(agree.Key)
		if !cert.finishDel {
			logger.Warningf("Replica %d has not complete del node")
			return false
		}
		n, view := pbft.getDelNV()
		if n != agree.N || view != agree.View {
			logger.Warningf("Replica %d invalid p entry in agree-update: expected n=%d/view=%d, get n=%d/view=%d", pbft.id, n, view, agree.N, agree.View)
			return false
		}
	}

	for _, p := range append(agree.Pset, agree.Qset...) {
		if !(p.View <= agree.View && p.SequenceNumber > agree.H && p.SequenceNumber <= agree.H+pbft.L) {
			logger.Warningf("Replica %d invalid p entry in agree-update: vc(v:%d h:%d) p(v:%d n:%d)", pbft.id, agree.View, agree.H, p.View, p.SequenceNumber)
			return false
		}
	}

	for _, c := range agree.Cset {
		if !(c.SequenceNumber >= agree.H && c.SequenceNumber <= agree.H+pbft.L) {
			logger.Warningf("Replica %d invalid c entry in agree-update: vc(v:%d h:%d) c(n:%d)", pbft.id, agree.View, agree.H, c.SequenceNumber)
			return false
		}
	}

	return true
}
