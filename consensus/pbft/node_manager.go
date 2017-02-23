//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"sync/atomic"

	"hyperchain/protos"

	"github.com/golang/protobuf/proto"
	"hyperchain/consensus/events"
	"fmt"
	"encoding/base64"
	"reflect"
	"time"
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
	pbft.persistNewNode(uint64(1))
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
	logger.Debugf("Replica %d received local delnode message for newId: %d, del node: %d", pbft.id, msg.Id, msg.Del)

	if pbft.N == 4 {
		logger.Criticalf("Replica %d receive del msg, but we don't support delete as there're only 4 nodes", pbft.id)
		return nil
	}

	if len(msg.DelPayload) == 0 || len(msg.RouterHash) == 0 || msg.Id == 0 || msg.Del == 0 {
		logger.Warningf("New replica %d received invalid local delNode message", pbft.id)
		return nil
	}

	pbft.inDeletingNode = true
	pbft.sendAgreeDelNode(key, msg.RouterHash, msg.Id, msg.Del)

	return nil
}

func (pbft *pbftProtocal) recvLocalRouters(routers []byte) {

	logger.Debugf("Replica %d received local routers: %v", pbft.id, routers)
	pbft.routers = routers

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
func (pbft *pbftProtocal) sendAgreeDelNode(key string, routerHash string, newId uint64, delId uint64) {

	logger.Debugf("Replica %d try to send delnode message for quit node", pbft.id)

	cert := pbft.getDelNodeCert(key)
	cert.newId = newId
	cert.delId = delId
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

	logger.Debugf("Replica %d received addnode from replica %d",
		pbft.id, add.ReplicaId)

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

	logger.Debugf("Replica %d received agree delnode from replica %d",
		pbft.id, del.ReplicaId)

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

	if pbft.isNewNode {
		logger.Debugf("Replica %d is new node, reject update routingTable", pbft.id)
		return nil
	}

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
	logger.Debugf("Replica %d update routingTable for %v", pbft.id, key)
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

	logger.Debugf("Replica %d update routingTable for %v", pbft.id, key)
	pbft.helper.UpdateTable(payload, false)
	time.Sleep(20 * time.Millisecond)
	pbft.inDeletingNode = false

	atomic.StoreUint32(&pbft.inUpdatingN, 1)
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

	logger.Noticef("Replica %d send ready_for_n as it already finished recovery", pbft.id)
	atomic.StoreUint32(&pbft.inUpdatingN, 1)


	ready := &ReadyForN{
		ReplicaId:	pbft.id,
		Key:		pbft.localKey,
	}

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
func (pbft *pbftProtocal) recvReadyforNforAdd(ready *ReadyForN) events.Event {

	logger.Debugf("Replica %d received ready_for_n from replica %d", pbft.id, ready.ReplicaId)

	if active := atomic.LoadUint32(&pbft.activeView); active == 0 {
		logger.Warningf("Replica %d is in view change, reject the ready_for_n message", pbft.id)
		return nil
	}

	cert := pbft.getAddNodeCert(ready.Key)

	if !cert.finishAdd {
		logger.Debugf("Replica %d has not done with addnode for key=%s", pbft.id, ready.Key)
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

	return pbft.sendAgreeUpdateNForAdd(agree)
}

func (pbft *pbftProtocal) sendAgreeUpdateNForAdd(agree *AgreeUpdateN) events.Event {

	if int(agree.N) == pbft.N && agree.View == pbft.view {
		logger.Debugf("Replica %d alreay finish update for N=%d/view=%d", pbft.id, pbft.N, pbft.view)
	}

	pbft.stopTimer()
	atomic.StoreUint32(&pbft.inUpdatingN, 1)

	if pbft.isNewNode {
		logger.Debugf("Replica %d does not need to send agree-update-n", pbft.id)
		return nil
	}

	pbft.agreeUpdateHelper(agree)
	logger.Infof("Replica %d sending agree-update-N, v:%d, h:%d, |C|:%d, |P|:%d, |Q|:%d",
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
	atomic.StoreUint32(&pbft.inUpdatingN, 1)

	// calculate the new N and view
	n, view := pbft.getDelNV(cert.delId)

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
	logger.Infof("Replica %d sending agree-update-n, v:%d, h:%d, |C|:%d, |P|:%d, |Q|:%d",
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
	pbft.recvAgreeUpdateN(agree)
	return nil
}

func (pbft *pbftProtocal) recvAgreeUpdateN(agree *AgreeUpdateN) events.Event {

	logger.Debugf("Replica %d received agree-update-n from replica %d, v:%d, h:%d, |C|:%d, |P|:%d, |Q|:%d",
		pbft.id, agree.ReplicaId, agree.View, agree.H, len(agree.Cset), len(agree.Pset), len(agree.Qset))

	if atomic.LoadUint32(&pbft.activeView) == 0 {
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

	if !pbft.checkAgreeUpdateN(agree) {
		logger.Debugf("Replica %d found agree-update-n message incorrect", pbft.id)
		return nil
	}

	key := aidx{
		v:	agree.View,
		n:	agree.N,
		id:	agree.ReplicaId,
		flag: agree.Flag,
	}
	if _, ok := pbft.agreeUpdateStore[key]; ok {
		logger.Warningf("Replica %d already has a agree-update-n message" +
			" for view=%d/n=%d from replica %d", pbft.id, agree.View, agree.N, agree.ReplicaId)
		return nil
	}

	pbft.agreeUpdateStore[key] = agree

	replicas := make(map[uint64]bool)
	for idx := range pbft.agreeUpdateStore {
		if !(idx.v == agree.View && idx.n == agree.N) {
			continue
		}
		replicas[idx.id] = true
	}

	// We only enter this if there are enough agree-update-n messages but locally not inUpdateN
	if agree.Flag && len(replicas) > pbft.f+1 && atomic.LoadUint32(&pbft.inUpdatingN) == 0 {
		logger.Warningf("Replica %d received f+1 agree-update-n messages, triggering sendAgreeUpdateNForAdd",
			pbft.id)
		pbft.firstRequestTimer.Stop()
		agree.ReplicaId = pbft.id
		return pbft.sendAgreeUpdateNForAdd(agree)
	}

	// We only enter this if there are enough agree-update-n messages but locally not inUpdateN
	if !agree.Flag && len(replicas) >= pbft.f+1 && atomic.LoadUint32(&pbft.inUpdatingN) == 0 {
		logger.Warningf("Replica %d received f+1 agree-update-n messages, triggering sendAgreeUpdateNForDel",
			pbft.id)
		pbft.firstRequestTimer.Stop()
		agree.ReplicaId = pbft.id
		return pbft.sendAgreeUpdateNforDel(agree.Key, agree.RouterHash)
	}

	quorum := 0
	for idx := range pbft.agreeUpdateStore {
		if idx.v == agree.View {
			quorum++
		}
	}
	logger.Debugf("Replica %d now has %d agree-update requests for view=%d/n=%d", pbft.id, quorum, agree.View, agree.N)

	if quorum >= pbft.allCorrectReplicasQuorum() {
		pbft.updateTarget = uidx{v:agree.View, n:agree.N, flag:agree.Flag, key:agree.Key}
		return agreeUpdateNQuorumEvent{}
	}

	return nil
}

func (pbft *pbftProtocal) sendUpdateN() events.Event {

	if pbft.inNegoView {
		logger.Debugf("Replica %d try to sendUpdateN, but it's in nego-view", pbft.id)
		return nil
	}

	if _, ok := pbft.updateStore[pbft.updateTarget]; ok {
		logger.Debugf("Replica %d already has update in store for n=%d/view=%d, skipping", pbft.id, pbft.updateTarget.n, pbft.updateTarget.v)
		return nil
	}

	aset := pbft.getAgreeUpdates()

	cp, ok, replicas := pbft.selectInitialCheckpointForUpdate(aset)

	if !ok {
		logger.Infof("Replica %d could not find consistent checkpoint: %+v", pbft.id, pbft.viewChangeStore)
		return nil
	}

	msgList := pbft.assignSequenceNumbersForUpdate(aset, cp.SequenceNumber)
	if msgList == nil {
		logger.Infof("Replica %d could not assign sequence numbers for new view", pbft.id)
		return nil
	}

	update := &UpdateN{
		Flag: pbft.updateTarget.flag,
		N: pbft.updateTarget.n,
		View:      pbft.updateTarget.v,
		Key: pbft.updateTarget.key,
		Aset:      aset,
		Xset:      msgList,
		ReplicaId: pbft.id,
	}

	logger.Infof("Replica %d is new primary, sending update-n, v:%d, X:%+v",
		pbft.id, update.View, update.Xset)
	payload, err := proto.Marshal(update)
	if err != nil {
		logger.Errorf("ConsensusMessage_UPDATE_N Marshal Error", err)
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_UPDATE_N,
		Payload: payload,
	}
	msg := consensusMsgHelper(consensusMsg, pbft.id)
	pbft.helper.InnerBroadcast(msg)
	pbft.updateStore[pbft.updateTarget] = update
	return pbft.primaryProcessUpdateN(cp, replicas, update)
}

func (pbft *pbftProtocal) recvUpdateN(update *UpdateN) events.Event {

	logger.Infof("Replica %d received update-n from replica %d",
		pbft.id, update.ReplicaId)

	if atomic.LoadUint32(&pbft.activeView) == 0 {
		logger.Warningf("Replica %d try to recvUpdateN, but it's in view-change", pbft.id)
		return nil
	}

	if pbft.inNegoView {
		logger.Debugf("Replica %d try to recvUpdateN, but it's in nego-view", pbft.id)
		return nil
	}

	if pbft.inRecovery {
		logger.Noticef("Replica %d try to recvUpdateN, but it's in recovery", pbft.id)
		pbft.recvNewViewInRecovery = true
		return nil
	}

	if !(update.View >= 0 && pbft.primary(pbft.view) == update.ReplicaId) {
		logger.Infof("Replica %d rejecting invalid new-view from %d, v:%d",
			pbft.id, update.ReplicaId, update.View)
		return nil
	}

	key := uidx{
		flag: update.Flag,
		key: update.Key,
		n: update.N,
		v: update.View,
	}
	pbft.updateStore[key] = update

	quorum := 0
	for idx := range pbft.agreeUpdateStore {
		if idx.v == update.View {
			quorum++
		}
	}

	if quorum < pbft.allCorrectReplicasQuorum() {
		logger.Warningf("Replica %d has not meet agreeUpdateNQuorum", pbft.id)
		return nil
	}

	return pbft.processUpdateN()
}

func (pbft *pbftProtocal) primaryProcessUpdateN(initialCp ViewChange_C, replicas []uint64, update *UpdateN) events.Event {
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
			err = fmt.Errorf("Replica %d received a agree-update-n whose hash could not be decoded (%s)", pbft.id, initialCp.Id)
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

	newReqBatchMissing = pbft.feedMissingReqBatchIfNeeded(update.Xset)

	if len(pbft.missingReqBatches) == 0 {
		return pbft.processReqInUpdate(update)
	} else if newReqBatchMissing {
		pbft.fetchRequestBatches()
	}

	return nil
}

func (pbft *pbftProtocal) processUpdateN() events.Event {

	update, ok := pbft.updateStore[pbft.updateTarget]
	if !ok {
		logger.Debugf("Replica %d ignoring processUpdateN as it could not find n=%d/view=%d in its updateStore", pbft.id, pbft.updateTarget.n, pbft.updateTarget.v)
		return nil
	}

	if atomic.LoadUint32(&pbft.activeView) == 0 {
		logger.Infof("Replica %d ignoring update-n from %d, v:%d: we are in view change to view=%d",
			pbft.id, update.ReplicaId, update.View, pbft.view)
		return nil
	}

	if atomic.LoadUint32(&pbft.inUpdatingN) == 0 {
		logger.Infof("Replica %d ignoring update-n from %d, v:%d: we are not in updatingN",
			pbft.id, update.ReplicaId, update.View)
		return nil
	}

	cp, ok, replicas := pbft.selectInitialCheckpointForUpdate(update.Aset)
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
	msgList := pbft.assignSequenceNumbersForUpdate(update.Aset, cp.SequenceNumber)

	if msgList == nil {
		logger.Warningf("Replica %d could not assign sequence numbers: %+v",
			pbft.id, pbft.viewChangeStore)
		return pbft.sendViewChange()
	}

	if !(len(msgList) == 0 && len(update.Xset) == 0) && !reflect.DeepEqual(msgList, update.Xset) {
		logger.Warningf("Replica %d failed to verify new-view Xset: computed %+v, received %+v",
			pbft.id, msgList, update.Xset)
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

	return pbft.processReqInUpdate(update)

}

func (pbft *pbftProtocal) processReqInUpdate(update *UpdateN) events.Event {
	logger.Debugf("Replica %d accepting update-n to target %v", pbft.id, pbft.updateTarget)

	if pbft.updateHandled {
		logger.Debugf("Replica %d repeated enter processReqInUpdate, ignore it")
		return nil
	}
	pbft.updateHandled = true

	pbft.stopTimer()
	pbft.nullRequestTimer.Stop()

	pbft.seqNo = pbft.h
	pbft.lastExec = pbft.h
	pbft.seqNo = pbft.h
	if pbft.primary(pbft.view) != pbft.id {
		pbft.vid = pbft.h
		pbft.lastVid = pbft.h
	}

	pbft.view = update.View
	pbft.N = int(update.N)
	pbft.f = (pbft.N - 1) / 3

	if !update.Flag {
		cert := pbft.getDelNodeCert(update.Key)
		pbft.id = cert.newId
	}

	delete(pbft.updateStore, pbft.updateTarget)

	for idx := range pbft.agreeUpdateStore {
		if (idx.v == update.View && idx.n == update.N && idx.flag == update.Flag) {
			delete(pbft.agreeUpdateStore, idx)
		}
	}

	for idx, cert := range pbft.certStore {
		if idx.v == pbft.view {
			tmpId := msgID{n:idx.n, v:update.View}
			tmpCert := updateCert{
				digest: cert.digest,
				sentPrepare: cert.sentPrepare,
				validated: cert.validated,
				sentCommit: cert.sentCommit,
				sentExecute: cert.sentExecute,
			}
			pbft.updateCertStore[tmpId] = tmpCert
			delete(pbft.certStore, idx)
			pbft.persistDelQPCSet(idx.v, idx.n)
		}
	}

	pbft.addNodeCertStore = make(map[string]*addNodeCert)
	pbft.delNodeCertStore = make(map[string]*delNodeCert)

	if !pbft.skipInProgress && !pbft.inVcReset && !pbft.inClearCache {
		pbft.helper.ClearValidateCache()
		time.Sleep(10 * time.Millisecond)
	}

	xSetLen := len(update.Xset)
	upper := uint64(xSetLen) + pbft.h + uint64(1)
	if pbft.primary(pbft.view) == pbft.id {
		for i := pbft.h + uint64(1); i < upper; i++ {
			d, ok := update.Xset[i]
			if !ok {
				logger.Critical("update_n Xset miss batch number %d", i)
			} else if d == "" {
				// This should not happen
				logger.Critical("update_n Xset has null batch, kick it out")
			} else {
				batch, ok := pbft.validatedBatchStore[d]
				if !ok {
					logger.Criticalf("In Xset %s exists, but in Replica %d validatedBatchStore there is no such batch digest", d, pbft.id)
				} else {
					pbft.sendPrePrepareForUpdate(batch, d)
				}
			}
		}
	}

	pbft.startTimerIfOutstandingRequests()

	return updatedNEvent{}
}

func (pbft *pbftProtocal) agreeUpdateHelper(agree *AgreeUpdateN) {

	pset := pbft.calcPSet()
	qset := pbft.calcQSet()

	pbft.plist = pset
	pbft.qlist = qset

	for idx := range pbft.agreeUpdateStore {
		if !(idx.v == agree.View && idx.n == agree.N && idx.flag == agree.Flag) {
			delete(pbft.agreeUpdateStore, idx)
		}
	}

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

func (pbft *pbftProtocal) checkAgreeUpdateN(agree *AgreeUpdateN) bool {

	if agree.Flag {
		cert := pbft.getAddNodeCert(agree.Key)
		if !pbft.isNewNode && !cert.finishAdd {
			logger.Debugf("Replica %d has not complete add node", pbft.id)
			return false
		}
		n, view := pbft.getAddNV()
		if n != agree.N || view != agree.View {
			logger.Debugf("Replica %d invalid p entry in agree-update: expected n=%d/view=%d, get n=%d/view=%d", pbft.id, n, view, agree.N, agree.View)
			return false
		}

		for _, p := range append(agree.Pset, agree.Qset...) {
			if !(p.View <= agree.View && p.SequenceNumber > agree.H && p.SequenceNumber <= agree.H+pbft.L) {
				logger.Debugf("Replica %d invalid p entry in agree-update: vc(v:%d h:%d) p(v:%d n:%d)", pbft.id, agree.View, agree.H, p.View, p.SequenceNumber)
				return false
			}
		}

	} else {
		cert := pbft.getDelNodeCert(agree.Key)
		if !cert.finishDel {
			logger.Debugf("Replica %d has not complete del node", pbft.id)
			return false
		}
		n, view := pbft.getDelNV(cert.delId)
		if n != agree.N || view != agree.View {
			logger.Debugf("Replica %d invalid p entry in agree-update: expected n=%d/view=%d, get n=%d/view=%d", pbft.id, n, view, agree.N, agree.View)
			return false
		}

		for _, p := range append(agree.Pset, agree.Qset...) {
			if !(p.View <= agree.View+1 && p.SequenceNumber > agree.H && p.SequenceNumber <= agree.H+pbft.L) {
				logger.Debugf("Replica %d invalid p entry in agree-update: vc(v:%d h:%d) p(v:%d n:%d)", pbft.id, agree.View, agree.H, p.View, p.SequenceNumber)
				return false
			}
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

func (pbft *pbftProtocal) getAgreeUpdates() (agrees []*AgreeUpdateN) {
	for _, agree := range pbft.agreeUpdateStore {
		agrees = append(agrees, agree)
	}
	return
}

func (pbft *pbftProtocal) selectInitialCheckpointForUpdate(aset []*AgreeUpdateN) (checkpoint ViewChange_C, ok bool, replicas []uint64) {
	checkpoints := make(map[ViewChange_C][]*AgreeUpdateN)
	for _, agree := range aset {
		for _, c := range agree.Cset { // TODO, verify that we strip duplicate checkpoints from this set
			checkpoints[*c] = append(checkpoints[*c], agree)
			logger.Debugf("Replica %d appending checkpoint from replica %d with seqNo=%d, h=%d, and checkpoint digest %s", pbft.id, agree.ReplicaId, agree.H, c.SequenceNumber, c.Id)
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
		for _, vc := range aset {
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

func (pbft *pbftProtocal) assignSequenceNumbersForUpdate(aset []*AgreeUpdateN, h uint64) (msgList map[uint64]string) {
	msgList = make(map[uint64]string)

	maxN := h + 1

	// "for all n such that h < n <= h + L"
	nLoop:
	for n := h + 1; n <= h+pbft.L; n++ {
		// "∃m ∈ S..."
		for _, m := range aset {
			// "...with <n,d,v> ∈ m.P"
			for _, em := range m.Pset {
				quorum := 0
				// "A1. ∃2f+1 messages m' ∈ S"
				mpLoop:
				for _, mp := range aset {
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
				for _, mp := range aset {
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
		for _, m := range aset {
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
		if n > maxN || msg == "" {
			delete(msgList, n)
		}
	}

	return
}