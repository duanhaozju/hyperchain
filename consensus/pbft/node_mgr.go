package pbft

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"sort"
	"sync/atomic"
	"time"

	"hyperchain/consensus/events"
	"hyperchain/manager/protos"

	"github.com/golang/protobuf/proto"
)

/**
Node control issues
*/

// nodeManager is the state manager of in configuration.
type nodeManager struct {
	localKey         string                  // track new node's local key (payload from local)
	addNodeCertStore map[string]*addNodeCert // track the received add node agree message
	delNodeCertStore map[string]*delNodeCert // track the received add node agree message

	routers           []byte // track the vp replicas' routers
	inUpdatingN       uint32 // track if there are updating
	updateTimer       events.Timer
	updateTimeout     time.Duration          // time limit for N-f agree on update n
	agreeUpdateStore  map[aidx]*AgreeUpdateN // track agree-update-n message
	updateStore       map[uidx]*UpdateN      // track last update-n we received or sent
	updateTarget      uidx                   // track the new view after update
	finishUpdateStore map[FinishUpdate]bool  // track finishUpdate message from others
}

// newNodeMgr creates a new nodeManager with the specified startup parameters.
func newNodeMgr() *nodeManager {
	nm := &nodeManager{}
	nm.addNodeCertStore = make(map[string]*addNodeCert)
	nm.delNodeCertStore = make(map[string]*delNodeCert)
	nm.agreeUpdateStore = make(map[aidx]*AgreeUpdateN)
	nm.updateStore = make(map[uidx]*UpdateN)
	nm.finishUpdateStore = make(map[FinishUpdate]bool)

	atomic.StoreUint32(&nm.inUpdatingN, 0)

	return nm
}

// dispatchNodeMgrMsg dispatches node manager service messages from other peers
// and uses corresponding function to handle them.
func (pbft *pbftImpl) dispatchNodeMgrMsg(e events.Event) events.Event {
	switch et := e.(type) {
	case *AddNode:
		return pbft.recvAgreeAddNode(et)
	case *DelNode:
		return pbft.recvAgreeDelNode(et)
	case *ReadyForN:
		return pbft.recvReadyforNforAdd(et)
	case *UpdateN:
		return pbft.recvUpdateN(et)
	case *AgreeUpdateN:
		return pbft.recvAgreeUpdateN(et)
	case *FinishUpdate:
		return pbft.recvFinishUpdate(et)
	}
	return nil
}

// recvLocalNewNode handles the local event about NewNode, which announces the
// consentor that it's a new node.
func (pbft *pbftImpl) recvLocalNewNode(msg *protos.NewNodeMessage) error {

	pbft.logger.Debugf("New replica %d received local newNode message", pbft.id)

	if pbft.status.getState(&pbft.status.isNewNode) {
		pbft.logger.Warningf("New replica %d received duplicate local newNode message", pbft.id)
		return nil
	}

	// the key about new node cannot be nil, it will results failure of updateN
	if len(msg.Payload) == 0 {
		pbft.logger.Warningf("New replica %d received nil local newNode message", pbft.id)
		return nil
	}

	pbft.status.activeState(&pbft.status.isNewNode, &pbft.status.inAddingNode)

	// new node store the new node state in db, since it may crash after connecting
	// with other nodes but before it truly participating in consensus
	pbft.persistNewNode(uint64(1))

	// the key of new node should be stored, since it will be use in processUpdateN
	key := string(msg.Payload)
	pbft.nodeMgr.localKey = key
	pbft.persistLocalKey(msg.Payload)

	return nil
}

// recvLocalAddNode handles the local event about AddNode, which announces the
// consentor that a new node want to participate in the consensus as a VP node.
func (pbft *pbftImpl) recvLocalAddNode(msg *protos.AddNodeMessage) error {

	if pbft.status.getState(&pbft.status.isNewNode) {
		pbft.logger.Warningf("New replica received local addNode message, there may be something wrong")
		return nil
	}

	if len(msg.Payload) == 0 {
		pbft.logger.Warningf("New replica %d received nil local addNode message", pbft.id)
		return nil
	}

	key := string(msg.Payload)
	pbft.logger.Debugf("Replica %d received local addNode message for new node %v", pbft.id, key)

	pbft.status.activeState(&pbft.status.inAddingNode)
	pbft.sendAgreeAddNode(key)

	return nil
}

// recvLocalAddNode handles the local event about DelNode, which announces the
// consentor that a VP node want to leave the consensus. Notice: the node
// that want to leave away from consensus should also have this message
func (pbft *pbftImpl) recvLocalDelNode(msg *protos.DelNodeMessage) error {

	key := string(msg.DelPayload)
	pbft.logger.Debugf("Replica %d received local delnode message for newId: %d, del node: %d", pbft.id, msg.Id, msg.Del)

	// We only support deleting when the nodes >= 5
	if pbft.N == 4 {
		pbft.logger.Criticalf("Replica %d receive del msg, but we don't support delete as there're only 4 nodes", pbft.id)
		return nil
	}

	if len(msg.DelPayload) == 0 {
		pbft.logger.Warningf("Replica %d received invalid local delNode message", pbft.id)
		return nil
	}

	pbft.status.activeState(&pbft.status.inDeletingNode)
	pbft.sendAgreeDelNode(key, msg.RouterHash, msg.Id, msg.Del)

	return nil
}

// recvLocalRouters handles the local event about the change of local routers.
func (pbft *pbftImpl) recvLocalRouters(routers []byte) {

	pbft.logger.Debugf("Replica %d received local routers: %v", pbft.id, routers)
	pbft.nodeMgr.routers = routers

}

// sendAgreeAddNode broadcasts AgreeAddNode message to other replicas to notify
// that itself had received the add-in request from new node.
func (pbft *pbftImpl) sendAgreeAddNode(key string) {

	pbft.logger.Debugf("Replica %d try to send addnode message for new node", pbft.id)

	// AgreeAddNode message can only sent by original nodes
	if pbft.status.getState(&pbft.status.isNewNode) {
		pbft.logger.Warningf("New replica try to send addnode message, there may be something wrong")
		return
	}

	add := &AddNode{
		ReplicaId: pbft.id,
		Key:       key,
	}

	payload, err := proto.Marshal(add)
	if err != nil {
		pbft.logger.Errorf("Marshal AddNode Error!")
		return
	}
	msg := &ConsensusMessage{
		Type:    ConsensusMessage_ADD_NODE,
		Payload: payload,
	}

	broadcast := cMsgToPbMsg(msg, pbft.id)
	pbft.helper.BroadcastAddNode(broadcast)
	pbft.recvAgreeAddNode(add)

}

// sendAgreeDelNode broadcasts AgreeDelNode message to other replicas to notify
// that itself had received the leave-away request from one node.
func (pbft *pbftImpl) sendAgreeDelNode(key string, routerHash string, newId uint64, delId uint64) {

	pbft.logger.Debugf("Replica %d try to send delnode message for quit node", pbft.id)

	cert := pbft.getDelNodeCert(key)
	cert.newId = newId
	cert.delId = delId

	del := &DelNode{
		ReplicaId: pbft.id,
		Key:       key,
	}

	payload, err := proto.Marshal(del)
	if err != nil {
		pbft.logger.Errorf("Marshal DelNode Error!")
		return
	}
	msg := &ConsensusMessage{
		Type:    ConsensusMessage_DEL_NODE,
		Payload: payload,
	}

	broadcast := cMsgToPbMsg(msg, pbft.id)
	pbft.helper.BroadcastDelNode(broadcast)
	pbft.recvAgreeDelNode(del)

}

// recvAgreeAddNode handles the AgreeAddNode message sent by others
func (pbft *pbftImpl) recvAgreeAddNode(add *AddNode) error {

	pbft.logger.Debugf("Replica %d received addnode from replica %d", pbft.id, add.ReplicaId)

	// Cast the vote of AgreeAddNode into an existing or new tally
	cert := pbft.getAddNodeCert(add.Key)
	ok := cert.addNodes[*add]
	if ok {
		pbft.logger.Warningf("Replica %d ignored duplicate addnode from %d", pbft.id, add.ReplicaId)
		return nil
	}
	cert.addNodes[*add] = true

	return pbft.maybeUpdateTableForAdd(add.Key)
}

// recvAgreeDelNode handles the AgreeDelNode message sent by others
func (pbft *pbftImpl) recvAgreeDelNode(del *DelNode) error {

	pbft.logger.Debugf("Replica %d received agree delnode from replica %d",
		pbft.id, del.ReplicaId)

	// Cast the vote of AgreeDelNode into an existing or new tally
	cert := pbft.getDelNodeCert(del.Key)
	ok := cert.delNodes[*del]
	if ok {
		pbft.logger.Warningf("Replica %d ignored duplicate agree delnode from %d", pbft.id, del.ReplicaId)
		return nil
	}
	cert.delNodes[*del] = true

	return pbft.maybeUpdateTableForDel(del.Key)
}

// maybeUpdateTableForAdd checks if the AgreeAddNode messages have reach
// the quorum or not. If yes, update the routing table to add the connect with
// the new node
func (pbft *pbftImpl) maybeUpdateTableForAdd(key string) error {

	// New node needn't to update the routing table since it already has the newest one
	if pbft.status.getState(&pbft.status.isNewNode) {
		pbft.logger.Debugf("Replica %d is new node, reject update routingTable", pbft.id)
		return nil
	}

	// Adding node stipulates all nodes should come into agreement
	cert := pbft.getAddNodeCert(key)
	if len(cert.addNodes) < pbft.allCorrectQuorum() {
		return nil
	}

	if !pbft.status.getState(&pbft.status.inAddingNode) {
		if cert.finishAdd {
			// This indicates a byzantine behavior that
			// some nodes repeatedly send AgreeAddNode messages.
			pbft.logger.Warningf("Replica %d has already finished adding node, but still recevice add msg from someone else", pbft.id)
			return nil
		} else {
			// This replica hasn't be connected by new node.
			// Since we don't have the key, we cannot complete the updating, it returns.
			pbft.logger.Warningf("Replica %d has not locally ready for adding, have not received connect from new replica", pbft.id)
			return nil
		}
	}

	cert.finishAdd = true
	payload := []byte(key)
	pbft.logger.Debugf("Replica %d update routingTable for %v", pbft.id, key)
	pbft.helper.UpdateTable(payload, true)
	pbft.status.inActiveState(&pbft.status.inAddingNode)

	return nil
}


// maybeUpdateTableForDel checks if the AgreeDelNode messages have reach
// the quorum or not. If yes, update the routing table to cut down the connect with
// the node request to exit.
func (pbft *pbftImpl) maybeUpdateTableForDel(key string) error {


	// Deleting node stipulates all nodes should come into agreement
	cert := pbft.getDelNodeCert(key)
	if len(cert.delNodes) < pbft.allCorrectQuorum() {
		return nil
	}

	if !pbft.status.getState(&pbft.status.inDeletingNode) {
		if cert.finishDel {
			// This indicates a byzantine behavior that
			// some nodes repeatedly send AgreeDelNode messages.
			pbft.logger.Warningf("Replica %d has already finished deleting node, but still recevice del msg from someone else", pbft.id)
			return nil
		} else {
			// This replica hasn't be connected by new node.
			// As we don't need the key, we can complete the updating.
			pbft.logger.Debugf("Replica %d has not locally ready but still accept deleting", pbft.id)
		}
	}

	cert.finishDel = true
	payload := []byte(key)

	pbft.logger.Debugf("Replica %d update routingTable for %v", pbft.id, key)
	pbft.helper.UpdateTable(payload, false)

	// Wait 20ms, since it takes p2p module some time to update the routing table.
	// If we return too immediately, we may fail at broadcasting to new node.
	time.Sleep(20 * time.Millisecond)
	pbft.status.inActiveState(&pbft.status.inDeletingNode)
	atomic.StoreUint32(&pbft.nodeMgr.inUpdatingN, 1)

	// As for deleting node, replicas broadcast AgreeUpdateN just after
	// updating routing table
	pbft.sendAgreeUpdateNforDel(key)

	return nil
}

// sendReadyForN broadcasts the ReadyForN message to others.
// Only new node will call this after finished recovery, ReadyForN message means that
// it has already caught up with others and wants to truly participate in the consensus.
func (pbft *pbftImpl) sendReadyForN() error {

	if !pbft.status.getState(&pbft.status.isNewNode) {
		pbft.logger.Errorf("Replica %d is not new one, but try to send ready_for_n", pbft.id)
		return nil
	}

	// If new node loses the local key, there may be something wrong
	if pbft.nodeMgr.localKey == "" {
		pbft.logger.Errorf("New replica %d doesn't have local key for ready_for_n", pbft.id)
		return nil
	}

	if atomic.LoadUint32(&pbft.activeView) == 0 {
		pbft.logger.Errorf("New replica %d finds itself in viewchange, not sending ready_for_n", pbft.id)
		return nil
	}

	pbft.logger.Noticef("Replica %d send ready_for_n as it already finished recovery", pbft.id)
	atomic.StoreUint32(&pbft.nodeMgr.inUpdatingN, 1)

	ready := &ReadyForN{
		ReplicaId: pbft.id,
		Key:       pbft.nodeMgr.localKey,
	}

	payload, err := proto.Marshal(ready)
	if err != nil {
		pbft.logger.Errorf("Marshal ReadyForN Error!")
		return nil
	}
	msg := &ConsensusMessage{
		Type:    ConsensusMessage_READY_FOR_N,
		Payload: payload,
	}

	// Broadcast to all VP nodes except itself
	broadcast := cMsgToPbMsg(msg, pbft.id)
	pbft.helper.InnerBroadcast(broadcast)

	return nil
}

// recvReadyforNforAdd handles the ReadyForN message sent by new node.
func (pbft *pbftImpl) recvReadyforNforAdd(ready *ReadyForN) events.Event {

	pbft.logger.Debugf("Replica %d received ready_for_n from replica %d", pbft.id, ready.ReplicaId)

	if atomic.LoadUint32(&pbft.activeView) == 0 {
		pbft.logger.Warningf("Replica %d is in view change, reject the ready_for_n message", pbft.id)
		return nil
	}

	cert := pbft.getAddNodeCert(ready.Key)

	if !cert.finishAdd {
		pbft.logger.Debugf("Replica %d has not done with addnode for key=%s", pbft.id, ready.Key)
		return nil
	}

	// Calculate the new N and view
	n, view := pbft.getAddNV()

	// Broadcast the AgreeUpdateN message
	agree := &AgreeUpdateN{
		Flag:      true,
		ReplicaId: pbft.id,
		Key:       ready.Key,
		N:         n,
		View:      view,
		H:         pbft.h,
	}

	return pbft.sendAgreeUpdateNForAdd(agree)
}

// sendAgreeUpdateNForAdd broadcasts the AgreeUpdateN message to others.
// This will be only called after receiving the ReadyForN message sent by new node.
func (pbft *pbftImpl) sendAgreeUpdateNForAdd(agree *AgreeUpdateN) events.Event {

	if atomic.LoadUint32(&pbft.nodeMgr.inUpdatingN) == 1 {
		pbft.logger.Debugf("Replica %d already in updatingN, ignore send agree-update-n again")
		return nil
	}

	// Replica may receive ReadyForN after it has already finished updatingN
	// (it happens in bad network environment)
	if int(agree.N) == pbft.N && agree.View == pbft.view {
		pbft.logger.Debugf("Replica %d alreay finish update for N=%d/view=%d", pbft.id, pbft.N, pbft.view)
		return nil
	}

	delete(pbft.nodeMgr.updateStore, pbft.nodeMgr.updateTarget)
	pbft.stopNewViewTimer()
	atomic.StoreUint32(&pbft.nodeMgr.inUpdatingN, 1)

	if pbft.status.getState(&pbft.status.isNewNode) {
		pbft.logger.Debugf("Replica %d does not need to send agree-update-n", pbft.id)
		return nil
	}

	// Generate the AgreeUpdateN message and broadcast it to others
	pbft.agreeUpdateHelper(agree)
	pbft.logger.Infof("Replica %d sending agree-update-N, v:%d, h:%d, |C|:%d, |P|:%d, |Q|:%d",
		pbft.id, agree.View, agree.H, len(agree.Cset), len(agree.Pset), len(agree.Qset))

	payload, err := proto.Marshal(agree)
	if err != nil {
		pbft.logger.Errorf("ConsensusMessage_AGREE_UPDATE_N Marshal Error", err)
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_AGREE_UPDATE_N,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, pbft.id)
	pbft.helper.InnerBroadcast(msg)
	return pbft.recvAgreeUpdateN(agree)
}

// sendAgreeUpdateNforDel broadcasts the AgreeUpdateN message to other notify
// that it agree update the View & N as deleting a node.
func (pbft *pbftImpl) sendAgreeUpdateNforDel(key string) error {

	pbft.logger.Debugf("Replica %d try to send update_n after finish del node", pbft.id)

	if atomic.LoadUint32(&pbft.activeView) == 0 {
		pbft.logger.Warningf("Primary %d is in view change, reject the ready_for_n message", pbft.id)
		return nil
	}

	cert := pbft.getDelNodeCert(key)

	if !cert.finishDel {
		pbft.logger.Errorf("Primary %d has not done with delnode for key=%s", pbft.id, key)
		return nil
	}
	delete(pbft.nodeMgr.updateStore, pbft.nodeMgr.updateTarget)
	pbft.stopNewViewTimer()
	atomic.StoreUint32(&pbft.nodeMgr.inUpdatingN, 1)

	// Calculate the new N and view
	n, view := pbft.getDelNV(cert.delId)

	agree := &AgreeUpdateN{
		Flag:      false,
		ReplicaId: pbft.id,
		Key:       key,
		N:         n,
		View:      view,
		H:         pbft.h,
	}

	pbft.agreeUpdateHelper(agree)
	pbft.logger.Infof("Replica %d sending agree-update-n, v:%d, h:%d, |C|:%d, |P|:%d, |Q|:%d",
		pbft.id, agree.View, agree.H, len(agree.Cset), len(agree.Pset), len(agree.Qset))

	payload, err := proto.Marshal(agree)
	if err != nil {
		pbft.logger.Errorf("ConsensusMessage_AGREE_UPDATE_N Marshal Error", err)
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_AGREE_UPDATE_N,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, pbft.id)
	pbft.helper.InnerBroadcast(msg)
	pbft.recvAgreeUpdateN(agree)
	return nil
}

// recvAgreeUpdateN handles the AgreeUpdateN message sent by others,
func (pbft *pbftImpl) recvAgreeUpdateN(agree *AgreeUpdateN) events.Event {

	pbft.logger.Debugf("Replica %d received agree-update-n from replica %d, v:%d, n:%d, flag:%v, h:%d, |C|:%d, |P|:%d, |Q|:%d",
		pbft.id, agree.ReplicaId, agree.View, agree.N, agree.Flag, agree.H, len(agree.Cset), len(agree.Pset), len(agree.Qset))

	// Reject response to updating N as replica is in viewChange, negoView or recovery
	if atomic.LoadUint32(&pbft.activeView) == 0 {
		pbft.logger.Warningf("Replica %d try to recvAgreeUpdateN, but it's in view-change", pbft.id)
		return nil
	}
	if pbft.status.getState(&pbft.status.inNegoView) {
		pbft.logger.Warningf("Replica %d try to recvAgreeUpdateN, but it's in nego-view", pbft.id)
		return nil
	}
	if pbft.status.getState(&pbft.status.inRecovery) {
		pbft.logger.Warningf("Replica %d try to recvAgreeUpdateN, but it's in recovery", pbft.id)
		return nil
	}

	// Check if the AgreeUpdateN message is valid or not
	if !pbft.checkAgreeUpdateN(agree) {
		pbft.logger.Debugf("Replica %d found agree-update-n message incorrect", pbft.id)
		return nil
	}

	key := aidx{
		v:    agree.View,
		n:    agree.N,
		id:   agree.ReplicaId,
		flag: agree.Flag,
	}

	// Cast the vote of AgreeUpdateN into an existing or new tally
	if _, ok := pbft.nodeMgr.agreeUpdateStore[key]; ok {
		pbft.logger.Warningf("Replica %d already has a agree-update-n message"+
			" for view=%d/n=%d from replica %d", pbft.id, agree.View, agree.N, agree.ReplicaId)
		return nil
	}
	pbft.nodeMgr.agreeUpdateStore[key] = agree

	// Count of the amount of AgreeUpdateN message for the same key
	replicas := make(map[uint64]bool)
	for idx := range pbft.nodeMgr.agreeUpdateStore {
		if !(idx.v == agree.View && idx.n == agree.N && idx.flag == agree.Flag) {
			continue
		}
		replicas[idx.id] = true
	}
	quorum := len(replicas)

	// We only enter this if there are enough agree-update-n messages but locally not inUpdateN
	if agree.Flag && quorum > pbft.oneCorrectQuorum() && atomic.LoadUint32(&pbft.nodeMgr.inUpdatingN) == 0 {
		pbft.logger.Warningf("Replica %d received f+1 agree-update-n messages, triggering sendAgreeUpdateNForAdd",
			pbft.id)
		pbft.timerMgr.stopTimer(FIRST_REQUEST_TIMER)
		agree.ReplicaId = pbft.id
		return pbft.sendAgreeUpdateNForAdd(agree)
	}

	// We only enter this if there are enough agree-update-n messages but locally not inUpdateN
	if !agree.Flag && quorum >= pbft.oneCorrectQuorum() && atomic.LoadUint32(&pbft.nodeMgr.inUpdatingN) == 0 {
		pbft.logger.Warningf("Replica %d received f+1 agree-update-n messages, triggering sendAgreeUpdateNForDel",
			pbft.id)
		pbft.timerMgr.stopTimer(FIRST_REQUEST_TIMER)
		agree.ReplicaId = pbft.id
		return pbft.sendAgreeUpdateNforDel(agree.Key)
	}

	pbft.logger.Debugf("Replica %d now has %d agree-update requests for view=%d/n=%d", pbft.id, quorum, agree.View, agree.N)

	// Quorum of AgreeUpdateN reach the N, replica can jump to NODE_MGR_AGREE_UPDATEN_QUORUM_EVENT,
	// which mean all nodes agree in updating N
	if quorum >= pbft.allCorrectReplicasQuorum() {
		pbft.nodeMgr.updateTarget = uidx{v: agree.View, n: agree.N, flag: agree.Flag, key: agree.Key}
		return &LocalEvent{
			Service:   NODE_MGR_SERVICE,
			EventType: NODE_MGR_AGREE_UPDATEN_QUORUM_EVENT,
		}
	}

	return nil
}

// sendUpdateN broadcasts the UpdateN message to other, it will only be called
// by primary like the NewView.
func (pbft *pbftImpl) sendUpdateN() events.Event {

	if pbft.status.getState(&pbft.status.inNegoView) {
		pbft.logger.Debugf("Replica %d try to sendUpdateN, but it's in nego-view", pbft.id)
		return nil
	}

	// Reject repeatedly broadcasting of UpdateN message
	if _, ok := pbft.nodeMgr.updateStore[pbft.nodeMgr.updateTarget]; ok {
		pbft.logger.Debugf("Replica %d already has update in store for n=%d/view=%d, skipping", pbft.id, pbft.nodeMgr.updateTarget.n, pbft.nodeMgr.updateTarget.v)
		return nil
	}

	aset := pbft.getAgreeUpdates()

	// Check if primary can find the initial checkpoint for updating
	cp, ok, replicas := pbft.selectInitialCheckpointForUpdate(aset)
	if !ok {
		pbft.logger.Infof("Replica %d could not find consistent checkpoint: %+v", pbft.id, pbft.vcMgr.viewChangeStore)
		return nil
	}

	// Assign the seqNo according to the set of AgreeUpdateN and the initial checkpoint
	msgList := pbft.assignSequenceNumbersForUpdate(aset, cp.SequenceNumber)
	if msgList == nil {
		pbft.logger.Infof("Replica %d could not assign sequence numbers for new view", pbft.id)
		return nil
	}

	update := &UpdateN{
		Flag:      pbft.nodeMgr.updateTarget.flag,
		N:         pbft.nodeMgr.updateTarget.n,
		View:      pbft.nodeMgr.updateTarget.v,
		Key:       pbft.nodeMgr.updateTarget.key,
		Aset:      aset,
		Xset:      msgList,
		ReplicaId: pbft.id,
	}

	pbft.logger.Infof("Replica %d is new primary, sending update-n, v:%d, X:%+v",
		pbft.id, update.View, update.Xset)
	payload, err := proto.Marshal(update)
	if err != nil {
		pbft.logger.Errorf("ConsensusMessage_UPDATE_N Marshal Error", err)
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_UPDATE_N,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, pbft.id)
	pbft.helper.InnerBroadcast(msg)
	pbft.nodeMgr.updateStore[pbft.nodeMgr.updateTarget] = update
	return pbft.primaryProcessUpdateN(cp, replicas, update)
}

// recvUpdateN handles the UpdateN message sent by primary.
func (pbft *pbftImpl) recvUpdateN(update *UpdateN) events.Event {

	pbft.logger.Infof("Replica %d received update-n from replica %d",
		pbft.id, update.ReplicaId)

	// Reject response to updating N as replica is in viewChange, negoView or recovery
	if atomic.LoadUint32(&pbft.activeView) == 0 {
		pbft.logger.Warningf("Replica %d try to recvUpdateN, but it's in view-change", pbft.id)
		return nil
	}
	if pbft.status.getState(&pbft.status.inNegoView) {
		pbft.logger.Debugf("Replica %d try to recvUpdateN, but it's in nego-view", pbft.id)
		return nil
	}
	if pbft.status.getState(&pbft.status.inRecovery) {
		pbft.logger.Noticef("Replica %d try to recvUpdateN, but it's in recovery", pbft.id)
		pbft.recoveryMgr.recvNewViewInRecovery = true
		return nil
	}

	// UpdateN can only be sent by primary
	if !(update.View >= 0 && pbft.primary(pbft.view) == update.ReplicaId) {
		pbft.logger.Infof("Replica %d rejecting invalid UpdateN from %d, v:%d",
			pbft.id, update.ReplicaId, update.View)
		return nil
	}

	key := uidx{
		flag: update.Flag,
		key:  update.Key,
		n:    update.N,
		v:    update.View,
	}
	pbft.nodeMgr.updateStore[key] = update

	// Count of the amount of AgreeUpdateN message for the same key
	quorum := 0
	for idx := range pbft.nodeMgr.agreeUpdateStore {
		if idx.v == update.View {
			quorum++
		}
	}
	// Reject to process UpdateN if replica has not reach allCorrectReplicasQuorum
	if quorum < pbft.allCorrectReplicasQuorum() {
		pbft.logger.Warningf("Replica %d has not meet agreeUpdateNQuorum", pbft.id)
		return nil
	}

	return pbft.processUpdateN()
}

// primaryProcessUpdateN processes the UpdateN message after it has already reached
// updateN-quorum
func (pbft *pbftImpl) primaryProcessUpdateN(initialCp ViewChange_C, replicas []replicaInfo, update *UpdateN) events.Event {
	var newReqBatchMissing bool

	speculativeLastExec := pbft.exec.lastExec
	if pbft.exec.currentExec != nil {
		speculativeLastExec = *pbft.exec.currentExec
	}

	if pbft.h < initialCp.SequenceNumber {
		pbft.moveWatermarks(initialCp.SequenceNumber)
	}

	// true means we can not execToTarget need state transfer
	if speculativeLastExec < initialCp.SequenceNumber {
		pbft.logger.Warningf("Replica %d missing base checkpoint %d (%s), our most recent execution %d", pbft.id, initialCp.SequenceNumber, initialCp.Id, speculativeLastExec)

		snapshotID, err := base64.StdEncoding.DecodeString(initialCp.Id)
		if nil != err {
			err = fmt.Errorf("Replica %d received a agree-update-n whose hash could not be decoded (%s)", pbft.id, initialCp.Id)
			pbft.logger.Error(err.Error())
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

	if len(pbft.storeMgr.missingReqBatches) == 0 {
		return pbft.processReqInUpdate(update)
	} else if newReqBatchMissing {
		pbft.fetchRequestBatches()
	}

	return nil
}

func (pbft *pbftImpl) processUpdateN() events.Event {

	update, ok := pbft.nodeMgr.updateStore[pbft.nodeMgr.updateTarget]
	if !ok {
		pbft.logger.Debugf("Replica %d ignoring processUpdateN as it could not find n=%d/view=%d in its updateStore", pbft.id, pbft.nodeMgr.updateTarget.n, pbft.nodeMgr.updateTarget.v)
		return nil
	}

	if atomic.LoadUint32(&pbft.activeView) == 0 {
		pbft.logger.Infof("Replica %d ignoring update-n from %d, v:%d: we are in view change to view=%d",
			pbft.id, update.ReplicaId, update.View, pbft.view)
		return nil
	}

	if atomic.LoadUint32(&pbft.nodeMgr.inUpdatingN) == 0 {
		pbft.logger.Infof("Replica %d ignoring update-n from %d, v:%d: we are not in updatingN",
			pbft.id, update.ReplicaId, update.View)
		return nil
	}

	cp, ok, replicas := pbft.selectInitialCheckpointForUpdate(update.Aset)
	if !ok {
		pbft.logger.Warningf("Replica %d could not determine initial checkpoint: %+v",
			pbft.id, pbft.vcMgr.viewChangeStore)
		return pbft.sendViewChange()
	}
	// 以上 primary 不必做
	speculativeLastExec := pbft.exec.lastExec
	if pbft.exec.currentExec != nil {
		speculativeLastExec = *pbft.exec.currentExec
	}

	// --
	msgList := pbft.assignSequenceNumbersForUpdate(update.Aset, cp.SequenceNumber)

	if msgList == nil {
		pbft.logger.Warningf("Replica %d could not assign sequence numbers: %+v",
			pbft.id, pbft.vcMgr.viewChangeStore)
		return pbft.sendViewChange()
	}

	if !(len(msgList) == 0 && len(update.Xset) == 0) && !reflect.DeepEqual(msgList, update.Xset) {
		pbft.logger.Warningf("Replica %d failed to verify new-view Xset: computed %+v, received %+v",
			pbft.id, msgList, update.Xset)
		return pbft.sendViewChange()
	}
	// -- primary 不必做
	if pbft.h < cp.SequenceNumber {
		pbft.moveWatermarks(cp.SequenceNumber)
	}

	if speculativeLastExec < cp.SequenceNumber {
		pbft.logger.Warningf("Replica %d missing base checkpoint %d (%s), our most recent execution %d", pbft.id, cp.SequenceNumber, cp.Id, speculativeLastExec)

		snapshotID, err := base64.StdEncoding.DecodeString(cp.Id)
		if nil != err {
			err = fmt.Errorf("Replica %d received a view change whose hash could not be decoded (%s)", pbft.id, cp.Id)
			pbft.logger.Error(err.Error())
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

func (pbft *pbftImpl) processReqInUpdate(update *UpdateN) events.Event {
	pbft.logger.Debugf("Replica %d accepting update-n to target %v", pbft.id, pbft.nodeMgr.updateTarget)

	if pbft.status.getState(&pbft.status.updateHandled) {
		pbft.logger.Debugf("Replica %d repeated enter processReqInUpdate, ignore it")
		return nil
	}
	pbft.status.activeState(&pbft.status.updateHandled)

	pbft.stopNewViewTimer()
	pbft.timerMgr.stopTimer(NULL_REQUEST_TIMER)

	backenVid := pbft.exec.lastExec + 1
	pbft.seqNo = pbft.exec.lastExec
	pbft.batchVdr.lastVid = pbft.exec.lastExec

	pbft.view = update.View
	pbft.N = int(update.N)
	pbft.f = (pbft.N - 1) / 3

	if !update.Flag {
		cert := pbft.getDelNodeCert(update.Key)
		pbft.id = cert.newId
	}

	for idx := range pbft.nodeMgr.agreeUpdateStore {
		if idx.v == update.View && idx.n == update.N && idx.flag == update.Flag {
			delete(pbft.nodeMgr.agreeUpdateStore, idx)
		}
	}

	pbft.nodeMgr.addNodeCertStore = make(map[string]*addNodeCert)
	pbft.nodeMgr.delNodeCertStore = make(map[string]*delNodeCert)

	if !pbft.status.getState(&pbft.status.skipInProgress) &&
		!pbft.status.getState(&pbft.status.inVcReset) {
		pbft.helper.VcReset(backenVid)
		pbft.status.activeState(&pbft.status.inVcReset)
	} else if pbft.primary(pbft.view) == pbft.id {
		pbft.logger.Warningf("New primary %d need to catch up other, waiting", pbft.id)
	} else {
		pbft.logger.Warningf("Replica %d cannot process local vcReset, but also send finishVcReset", pbft.id)
		pbft.sendFinishUpdate()
	}

	return nil
}

func (pbft *pbftImpl) sendFinishUpdate() events.Event {

	finish := &FinishUpdate{
		ReplicaId: pbft.id,
		View:      pbft.view,
		LowH:      pbft.h,
	}

	payload, err := proto.Marshal(finish)
	if err != nil {
		pbft.logger.Errorf("Marshal FinishUpdate Error!")
		return nil
	}
	msg := &ConsensusMessage{
		Type:    ConsensusMessage_FINISH_UPDATE,
		Payload: payload,
	}

	broadcast := cMsgToPbMsg(msg, pbft.id)
	pbft.helper.InnerBroadcast(broadcast)
	pbft.logger.Debugf("Replica %d broadcast FinishUpdate for view=%d, h=%d", pbft.id, pbft.view, pbft.h)

	return pbft.recvFinishUpdate(finish)
}

func (pbft *pbftImpl) recvFinishUpdate(finish *FinishUpdate) events.Event {
	if atomic.LoadUint32(&pbft.nodeMgr.inUpdatingN) == 0 {
		pbft.logger.Debugf("Replica %d is not in updatingN, but received FinishUpdate from replica %d", pbft.id, finish.ReplicaId)
	}
	pbft.logger.Debugf("Replica %d received FinishUpdate from replica %d, view=%d/h=%d", pbft.id, finish.ReplicaId, finish.View, finish.LowH)

	if finish.View != pbft.view {
		pbft.logger.Warningf("Replica %d received FinishUpdate from replica %d, expect view=%d, but get view=%d", pbft.id, finish.ReplicaId, pbft.view, finish.View)
		// We don't ignore it as there may be delay during receiveUpdate from primary between new node and other non-primary old replicas
	}

	ok := pbft.nodeMgr.finishUpdateStore[*finish]
	if ok {
		pbft.logger.Warningf("Replica %d ignored duplicator agree FinishUpdate from %d", pbft.id, finish.ReplicaId)
		return nil
	}
	pbft.nodeMgr.finishUpdateStore[*finish] = true

	return pbft.handleTailAfterUpdate()
}

func (pbft *pbftImpl) handleTailAfterUpdate() events.Event {
	quorum := 0
	hasPrimary := false
	for finish := range pbft.nodeMgr.finishUpdateStore {
		if finish.View == pbft.view {
			quorum++
			if finish.ReplicaId == pbft.primary(pbft.view) {
				hasPrimary = true
			}
		}
	}

	if quorum < pbft.allCorrectReplicasQuorum() {
		pbft.logger.Debugf("Replica %d doesn't has arrive quorum FinishUpdate, expect=%d, got=%d",
			pbft.id, pbft.allCorrectReplicasQuorum(), quorum)
		return nil
	}
	if !hasPrimary {
		pbft.logger.Debugf("Replica %d doesn't receive FinishUpdate from primary %d",
			pbft.id, pbft.primary(pbft.view))
		return nil
	}

	if pbft.status.getState(&pbft.status.inVcReset) {
		pbft.logger.Warningf("Replica %d itself has not finished update", pbft.id)
		return nil
	}

	update, ok := pbft.nodeMgr.updateStore[pbft.nodeMgr.updateTarget]
	if !ok {
		pbft.logger.Debugf("Primary %d ignoring handleTailAfterUpdate as it could not find target %v in its updateStore", pbft.id, pbft.nodeMgr.updateTarget)
		return nil
	}

	pbft.stopNewViewTimer()

	pbft.seqNo = pbft.exec.lastExec
	pbft.batchVdr.lastVid = pbft.exec.lastExec

	if pbft.primary(pbft.view) == pbft.id {
		xSetLen := len(update.Xset)
		upper := uint64(xSetLen) + pbft.h + uint64(1)
		for i := pbft.h + uint64(1); i < upper; i++ {
			d, ok := update.Xset[i]
			if !ok {
				pbft.logger.Critical("update_n Xset miss batch number %d", i)
			} else if d == "" {
				// This should not happen
				pbft.logger.Critical("update_n Xset has null batch, kick it out")
			} else {
				batch, ok := pbft.storeMgr.txBatchStore[d]
				if !ok {
					pbft.logger.Criticalf("In Xset %s exists, but in Replica %d validatedBatchStore there is no such batch digest", d, pbft.id)
				} else if i > pbft.seqNo {
					pbft.primaryValidateBatch(d, batch, i)
				}
			}
		}
	}

	return &LocalEvent{
		Service:   NODE_MGR_SERVICE,
		EventType: NODE_MGR_UPDATEDN_EVENT,
	}
}

func (pbft *pbftImpl) rebuildCertStoreForUpdate() {

	update, ok := pbft.nodeMgr.updateStore[pbft.nodeMgr.updateTarget]
	if !ok {
		pbft.logger.Debugf("Primary %d ignoring rebuildCertStore as it could not find target %v in its updateStore", pbft.id, pbft.nodeMgr.updateTarget)
		return
	}
	pbft.rebuildCertStore(update.Xset)

}

// agreeUpdateHelper helps generate the AgreeUpdateN message.
func (pbft *pbftImpl) agreeUpdateHelper(agree *AgreeUpdateN) {

	pset := pbft.calcPSet()
	qset := pbft.calcQSet()

	pbft.vcMgr.plist = pset
	pbft.vcMgr.qlist = qset

	for idx := range pbft.nodeMgr.agreeUpdateStore {
		if !(idx.v == agree.View && idx.n == agree.N && idx.flag == agree.Flag) {
			delete(pbft.nodeMgr.agreeUpdateStore, idx)
		}
	}

	for n, id := range pbft.storeMgr.chkpts {
		agree.Cset = append(agree.Cset, &ViewChange_C{
			SequenceNumber: n,
			Id:             id,
		})
	}

	for _, p := range pset {
		if p.SequenceNumber < pbft.h {
			pbft.logger.Errorf("BUG! Replica %d should not have anything in our pset less than h, found %+v", pbft.id, p)
		}
		agree.Pset = append(agree.Pset, p)
	}

	for _, q := range qset {
		if q.SequenceNumber < pbft.h {
			pbft.logger.Errorf("BUG! Replica %d should not have anything in our qset less than h, found %+v", pbft.id, q)
		}
		agree.Qset = append(agree.Qset, q)
	}
}

// checkAgreeUpdateN checks the AgreeUpdateN message if it's valid or not.
func (pbft *pbftImpl) checkAgreeUpdateN(agree *AgreeUpdateN) bool {

	if agree.Flag {
		cert := pbft.getAddNodeCert(agree.Key)
		if !pbft.status.getState(&pbft.status.isNewNode) && !cert.finishAdd {
			pbft.logger.Debugf("Replica %d has not complete add node", pbft.id)
			return false
		}
		n, view := pbft.getAddNV()
		if n != agree.N || view != agree.View {
			pbft.logger.Debugf("Replica %d invalid p entry in agree-update: expected n=%d/view=%d, get n=%d/view=%d", pbft.id, n, view, agree.N, agree.View)
			return false
		}

		for _, p := range append(agree.Pset, agree.Qset...) {
			if !(p.View <= agree.View && p.SequenceNumber > agree.H && p.SequenceNumber <= agree.H+pbft.L) {
				pbft.logger.Debugf("Replica %d invalid p entry in agree-update: agree(v:%d h:%d) p(v:%d n:%d)", pbft.id, agree.View, agree.H, p.View, p.SequenceNumber)
				return false
			}
		}

	} else {
		cert := pbft.getDelNodeCert(agree.Key)
		if !cert.finishDel {
			pbft.logger.Debugf("Replica %d has not complete del node", pbft.id)
			return false
		}
		n, view := pbft.getDelNV(cert.delId)
		if n != agree.N || view != agree.View {
			pbft.logger.Debugf("Replica %d invalid p entry in agree-update: expected n=%d/view=%d, get n=%d/view=%d", pbft.id, n, view, agree.N, agree.View)
			return false
		}

		for _, p := range append(agree.Pset, agree.Qset...) {
			if !(p.View <= agree.View+1 && p.SequenceNumber > agree.H && p.SequenceNumber <= agree.H+pbft.L) {
				pbft.logger.Debugf("Replica %d invalid p entry in agree-update: agree(v:%d h:%d) p(v:%d n:%d)", pbft.id, agree.View, agree.H, p.View, p.SequenceNumber)
				return false
			}
		}

	}

	for _, c := range agree.Cset {
		if !(c.SequenceNumber >= agree.H && c.SequenceNumber <= agree.H+pbft.L) {
			pbft.logger.Warningf("Replica %d invalid c entry in agree-update: agree(v:%d h:%d) c(n:%d)", pbft.id, agree.View, agree.H, c.SequenceNumber)
			return false
		}
	}

	return true
}

// getAgreeUpdates gets all the AgreeUpdateN messages the replica received.
func (pbft *pbftImpl) getAgreeUpdates() (agrees []*AgreeUpdateN) {
	for _, agree := range pbft.nodeMgr.agreeUpdateStore {
		agrees = append(agrees, agree)
	}
	return
}

func (pbft *pbftImpl) selectInitialCheckpointForUpdate(aset []*AgreeUpdateN) (checkpoint ViewChange_C, ok bool, replicas []replicaInfo) {
	checkpoints := make(map[ViewChange_C][]*AgreeUpdateN)
	for _, agree := range aset {
		for _, c := range agree.Cset { // TODO, verify that we strip duplicate checkpoints from this set
			checkpoints[*c] = append(checkpoints[*c], agree)
			pbft.logger.Debugf("Replica %d appending checkpoint from replica %d with seqNo=%d, h=%d, and checkpoint digest %s", pbft.id, agree.ReplicaId, agree.H, c.SequenceNumber, c.Id)
		}
	}

	if len(checkpoints) == 0 {
		pbft.logger.Debugf("Replica %d has no checkpoints to select from: %d %s",
			pbft.id, len(pbft.vcMgr.viewChangeStore), checkpoints)
		return
	}

	for idx, vcList := range checkpoints {
		// need weak certificate for the checkpoint
		if len(vcList) < pbft.oneCorrectQuorum() { // type casting necessary to match types
			pbft.logger.Debugf("Replica %d has no weak certificate for n:%d, vcList was %d long",
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

		if quorum < pbft.commonCaseQuorum() {
			pbft.logger.Debugf("Replica %d has no quorum for n:%d", pbft.id, idx.SequenceNumber)
			continue
		}

		if checkpoint.SequenceNumber <= idx.SequenceNumber {
			replicas = make([]replicaInfo, len(vcList))
			for i, vc := range vcList {
				replicas[i] = replicaInfo{
					id:      vc.ReplicaId,
					height:  vc.H,
					genesis: vc.Genesis,
				}
			}

			checkpoint = idx
			ok = true
		}
	}

	return
}

func (pbft *pbftImpl) assignSequenceNumbersForUpdate(aset []*AgreeUpdateN, h uint64) map[uint64]string {
	msgList := make(map[uint64]string)

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

				if quorum < pbft.commonCaseQuorum() {
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

				if quorum < pbft.oneCorrectQuorum() {
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

		if quorum >= pbft.commonCaseQuorum() {
			// "then select the null request for number n"
			msgList[n] = ""

			continue nLoop
		}

		pbft.logger.Warningf("Replica %d could not assign value to contents of seqNo %d, found only %d missing P entries", pbft.id, n, quorum)
		return nil
	}

	// Prune top null requests
	for n, msg := range msgList {
		if n > maxN || msg == "" {
			delete(msgList, n)
		}
	}

	keys := make([]uint64, len(msgList))
	i := 0
	for n := range msgList {
		keys[i] = n
		i++
	}
	sort.Sort(sortableUint64Slice(keys))
	x := h + 1
	list := make(map[uint64]string)
	for _, n := range keys {
		list[x] = msgList[n]
		x++
	}

	return list
}
