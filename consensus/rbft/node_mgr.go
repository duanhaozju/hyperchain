package rbft

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
func (rbft *rbftImpl) dispatchNodeMgrMsg(e events.Event) events.Event {
	switch et := e.(type) {
	case *AddNode:
		return rbft.recvAgreeAddNode(et)
	case *DelNode:
		return rbft.recvAgreeDelNode(et)
	case *ReadyForN:
		return rbft.recvReadyforNforAdd(et)
	case *UpdateN:
		return rbft.recvUpdateN(et)
	case *AgreeUpdateN:
		return rbft.recvAgreeUpdateN(et)
	case *FinishUpdate:
		return rbft.recvFinishUpdate(et)
	}
	return nil
}

// recvLocalNewNode handles the local event about NewNode, which announces the
// consentor that it's a new node.
func (rbft *rbftImpl) recvLocalNewNode(msg *protos.NewNodeMessage) error {

	rbft.logger.Debugf("New replica %d received local newNode message", rbft.id)

	if rbft.status.getState(&rbft.status.isNewNode) {
		rbft.logger.Warningf("New replica %d received duplicate local newNode message", rbft.id)
		return nil
	}

	// the key about new node cannot be nil, it will results failure of updateN
	if len(msg.Payload) == 0 {
		rbft.logger.Warningf("New replica %d received nil local newNode message", rbft.id)
		return nil
	}

	rbft.status.activeState(&rbft.status.isNewNode, &rbft.status.inAddingNode)

	// new node store the new node state in db, since it may crash after connecting
	// with other nodes but before it truly participating in consensus
	rbft.persistNewNode(uint64(1))

	// the key of new node should be stored, since it will be use in processUpdateN
	key := string(msg.Payload)
	rbft.nodeMgr.localKey = key
	rbft.persistLocalKey(msg.Payload)

	return nil
}

// recvLocalAddNode handles the local event about AddNode, which announces the
// consentor that a new node want to participate in the consensus as a VP node.
func (rbft *rbftImpl) recvLocalAddNode(msg *protos.AddNodeMessage) error {

	if rbft.status.getState(&rbft.status.isNewNode) {
		rbft.logger.Warningf("New replica received local addNode message, there may be something wrong")
		return nil
	}

	if len(msg.Payload) == 0 {
		rbft.logger.Warningf("New replica %d received nil local addNode message", rbft.id)
		return nil
	}

	key := string(msg.Payload)
	rbft.logger.Debugf("Replica %d received local addNode message for new node %v", rbft.id, key)

	rbft.status.activeState(&rbft.status.inAddingNode)
	rbft.sendAgreeAddNode(key)

	return nil
}

// recvLocalAddNode handles the local event about DelNode, which announces the
// consentor that a VP node want to leave the consensus. Notice: the node
// that want to leave away from consensus should also have this message
func (rbft *rbftImpl) recvLocalDelNode(msg *protos.DelNodeMessage) error {

	key := string(msg.DelPayload)
	rbft.logger.Debugf("Replica %d received local delnode message for newId: %d, del node: %d", rbft.id, msg.Id, msg.Del)

	// We only support deleting when the nodes >= 5
	if rbft.N == 4 {
		rbft.logger.Criticalf("Replica %d receive del msg, but we don't support delete as there're only 4 nodes", rbft.id)
		return nil
	}

	if len(msg.DelPayload) == 0 {
		rbft.logger.Warningf("Replica %d received invalid local delNode message", rbft.id)
		return nil
	}

	rbft.status.activeState(&rbft.status.inDeletingNode)
	rbft.sendAgreeDelNode(key, msg.RouterHash, msg.Id, msg.Del)

	return nil
}

// recvLocalRouters handles the local event about the change of local routers.
func (rbft *rbftImpl) recvLocalRouters(routers []byte) {

	rbft.logger.Debugf("Replica %d received local routers: %v", rbft.id, routers)
	rbft.nodeMgr.routers = routers

}

// sendAgreeAddNode broadcasts AgreeAddNode message to other replicas to notify
// that itself had received the add-in request from new node.
func (rbft *rbftImpl) sendAgreeAddNode(key string) {

	rbft.logger.Debugf("Replica %d try to send addnode message for new node", rbft.id)

	// AgreeAddNode message can only sent by original nodes
	if rbft.status.getState(&rbft.status.isNewNode) {
		rbft.logger.Warningf("New replica try to send addnode message, there may be something wrong")
		return
	}

	add := &AddNode{
		ReplicaId: rbft.id,
		Key:       key,
	}

	payload, err := proto.Marshal(add)
	if err != nil {
		rbft.logger.Errorf("Marshal AddNode Error!")
		return
	}
	msg := &ConsensusMessage{
		Type:    ConsensusMessage_ADD_NODE,
		Payload: payload,
	}

	broadcast := cMsgToPbMsg(msg, rbft.id)
	rbft.helper.BroadcastAddNode(broadcast)
	rbft.recvAgreeAddNode(add)

}

// sendAgreeDelNode broadcasts AgreeDelNode message to other replicas to notify
// that itself had received the leave-away request from one node.
func (rbft *rbftImpl) sendAgreeDelNode(key string, routerHash string, newId uint64, delId uint64) {

	rbft.logger.Debugf("Replica %d try to send delnode message for quit node", rbft.id)

	cert := rbft.getDelNodeCert(key)
	cert.newId = newId
	cert.delId = delId

	del := &DelNode{
		ReplicaId: rbft.id,
		Key:       key,
	}

	payload, err := proto.Marshal(del)
	if err != nil {
		rbft.logger.Errorf("Marshal DelNode Error!")
		return
	}
	msg := &ConsensusMessage{
		Type:    ConsensusMessage_DEL_NODE,
		Payload: payload,
	}

	broadcast := cMsgToPbMsg(msg, rbft.id)
	rbft.helper.BroadcastDelNode(broadcast)
	rbft.recvAgreeDelNode(del)

}

// recvAgreeAddNode handles the AgreeAddNode message sent by others
func (rbft *rbftImpl) recvAgreeAddNode(add *AddNode) error {

	rbft.logger.Debugf("Replica %d received addnode from replica %d", rbft.id, add.ReplicaId)

	// Cast the vote of AgreeAddNode into an existing or new tally
	cert := rbft.getAddNodeCert(add.Key)
	ok := cert.addNodes[*add]
	if ok {
		rbft.logger.Warningf("Replica %d ignored duplicate addnode from %d", rbft.id, add.ReplicaId)
		return nil
	}
	cert.addNodes[*add] = true

	return rbft.maybeUpdateTableForAdd(add.Key)
}

// recvAgreeDelNode handles the AgreeDelNode message sent by others
func (rbft *rbftImpl) recvAgreeDelNode(del *DelNode) error {

	rbft.logger.Debugf("Replica %d received agree delnode from replica %d",
		rbft.id, del.ReplicaId)

	// Cast the vote of AgreeDelNode into an existing or new tally
	cert := rbft.getDelNodeCert(del.Key)
	ok := cert.delNodes[*del]
	if ok {
		rbft.logger.Warningf("Replica %d ignored duplicate agree delnode from %d", rbft.id, del.ReplicaId)
		return nil
	}
	cert.delNodes[*del] = true

	return rbft.maybeUpdateTableForDel(del.Key)
}

// maybeUpdateTableForAdd checks if the AgreeAddNode messages have reach
// the quorum or not. If yes, update the routing table to add the connect with
// the new node
func (rbft *rbftImpl) maybeUpdateTableForAdd(key string) error {

	// New node needn't to update the routing table since it already has the newest one
	if rbft.status.getState(&rbft.status.isNewNode) {
		rbft.logger.Debugf("Replica %d is new node, reject update routingTable", rbft.id)
		return nil
	}

	// Adding node stipulates all nodes should come into agreement
	cert := rbft.getAddNodeCert(key)
	if len(cert.addNodes) < rbft.allCorrectQuorum() {
		return nil
	}

	if !rbft.status.getState(&rbft.status.inAddingNode) {
		if cert.finishAdd {
			// This indicates a byzantine behavior that
			// some nodes repeatedly send AgreeAddNode messages.
			rbft.logger.Warningf("Replica %d has already finished adding node, but still recevice add msg from someone else", rbft.id)
			return nil
		} else {
			// This replica hasn't be connected by new node.
			// Since we don't have the key, we cannot complete the updating, it returns.
			rbft.logger.Warningf("Replica %d has not locally ready for adding, have not received connect from new replica", rbft.id)
			return nil
		}
	}

	cert.finishAdd = true
	payload := []byte(key)
	rbft.logger.Debugf("Replica %d update routingTable for %v", rbft.id, key)
	rbft.helper.UpdateTable(payload, true)
	rbft.status.inActiveState(&rbft.status.inAddingNode)

	return nil
}

// maybeUpdateTableForDel checks if the AgreeDelNode messages have reach
// the quorum or not. If yes, update the routing table to cut down the connect with
// the node request to exit.
func (rbft *rbftImpl) maybeUpdateTableForDel(key string) error {

	// Deleting node stipulates all nodes should come into agreement
	cert := rbft.getDelNodeCert(key)
	if len(cert.delNodes) < rbft.allCorrectQuorum() {
		return nil
	}

	if !rbft.status.getState(&rbft.status.inDeletingNode) {
		if cert.finishDel {
			// This indicates a byzantine behavior that
			// some nodes repeatedly send AgreeDelNode messages.
			rbft.logger.Warningf("Replica %d has already finished deleting node, but still recevice del msg from someone else", rbft.id)
			return nil
		} else {
			// This replica hasn't be connected by new node.
			// As we don't need the key, we can complete the updating.
			rbft.logger.Debugf("Replica %d has not locally ready but still accept deleting", rbft.id)
		}
	}

	cert.finishDel = true
	payload := []byte(key)

	rbft.logger.Debugf("Replica %d update routingTable for %v", rbft.id, key)
	rbft.helper.UpdateTable(payload, false)

	// Wait 20ms, since it takes p2p module some time to update the routing table.
	// If we return too immediately, we may fail at broadcasting to new node.
	time.Sleep(20 * time.Millisecond)
	rbft.status.inActiveState(&rbft.status.inDeletingNode)
	atomic.StoreUint32(&rbft.nodeMgr.inUpdatingN, 1)

	// As for deleting node, replicas broadcast AgreeUpdateN just after
	// updating routing table
	rbft.sendAgreeUpdateNforDel(key)

	return nil
}

// sendReadyForN broadcasts the ReadyForN message to others.
// Only new node will call this after finished recovery, ReadyForN message means that
// it has already caught up with others and wants to truly participate in the consensus.
func (rbft *rbftImpl) sendReadyForN() error {

	if !rbft.status.getState(&rbft.status.isNewNode) {
		rbft.logger.Errorf("Replica %d is not new one, but try to send ready_for_n", rbft.id)
		return nil
	}

	// If new node loses the local key, there may be something wrong
	if rbft.nodeMgr.localKey == "" {
		rbft.logger.Errorf("New replica %d doesn't have local key for ready_for_n", rbft.id)
		return nil
	}

	if atomic.LoadUint32(&rbft.activeView) == 0 {
		rbft.logger.Errorf("New replica %d finds itself in viewchange, not sending ready_for_n", rbft.id)
		return nil
	}

	rbft.logger.Noticef("Replica %d send ready_for_n as it already finished recovery", rbft.id)
	atomic.StoreUint32(&rbft.nodeMgr.inUpdatingN, 1)

	ready := &ReadyForN{
		ReplicaId: rbft.id,
		Key:       rbft.nodeMgr.localKey,
	}

	payload, err := proto.Marshal(ready)
	if err != nil {
		rbft.logger.Errorf("Marshal ReadyForN Error!")
		return nil
	}
	msg := &ConsensusMessage{
		Type:    ConsensusMessage_READY_FOR_N,
		Payload: payload,
	}

	// Broadcast to all VP nodes except itself
	broadcast := cMsgToPbMsg(msg, rbft.id)
	rbft.helper.InnerBroadcast(broadcast)

	return nil
}

// recvReadyforNforAdd handles the ReadyForN message sent by new node.
func (rbft *rbftImpl) recvReadyforNforAdd(ready *ReadyForN) events.Event {

	rbft.logger.Debugf("Replica %d received ready_for_n from replica %d", rbft.id, ready.ReplicaId)

	if atomic.LoadUint32(&rbft.activeView) == 0 {
		rbft.logger.Warningf("Replica %d is in view change, reject the ready_for_n message", rbft.id)
		return nil
	}

	cert := rbft.getAddNodeCert(ready.Key)

	if !cert.finishAdd {
		rbft.logger.Debugf("Replica %d has not done with addnode for key=%s", rbft.id, ready.Key)
		return nil
	}

	// Calculate the new N and view
	n, view := rbft.getAddNV()

	// Broadcast the AgreeUpdateN message
	agree := &AgreeUpdateN{
		Flag:      true,
		ReplicaId: rbft.id,
		Key:       ready.Key,
		N:         n,
		View:      view,
		H:         rbft.h,
	}

	return rbft.sendAgreeUpdateNForAdd(agree)
}

// sendAgreeUpdateNForAdd broadcasts the AgreeUpdateN message to others.
// This will be only called after receiving the ReadyForN message sent by new node.
func (rbft *rbftImpl) sendAgreeUpdateNForAdd(agree *AgreeUpdateN) events.Event {

	if atomic.LoadUint32(&rbft.nodeMgr.inUpdatingN) == 1 {
		rbft.logger.Debugf("Replica %d already in updatingN, ignore send agree-update-n again")
		return nil
	}

	// Replica may receive ReadyForN after it has already finished updatingN
	// (it happens in bad network environment)
	if int(agree.N) == rbft.N && agree.View == rbft.view {
		rbft.logger.Debugf("Replica %d alreay finish update for N=%d/view=%d", rbft.id, rbft.N, rbft.view)
		return nil
	}

	delete(rbft.nodeMgr.updateStore, rbft.nodeMgr.updateTarget)
	rbft.stopNewViewTimer()
	atomic.StoreUint32(&rbft.nodeMgr.inUpdatingN, 1)

	if rbft.status.getState(&rbft.status.isNewNode) {
		rbft.logger.Debugf("Replica %d does not need to send agree-update-n", rbft.id)
		return nil
	}

	// Generate the AgreeUpdateN message and broadcast it to others
	rbft.agreeUpdateHelper(agree)
	rbft.logger.Infof("Replica %d sending agree-update-N, v:%d, h:%d, |C|:%d, |P|:%d, |Q|:%d",
		rbft.id, agree.View, agree.H, len(agree.Cset), len(agree.Pset), len(agree.Qset))

	payload, err := proto.Marshal(agree)
	if err != nil {
		rbft.logger.Errorf("ConsensusMessage_AGREE_UPDATE_N Marshal Error", err)
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_AGREE_UPDATE_N,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, rbft.id)
	rbft.helper.InnerBroadcast(msg)
	return rbft.recvAgreeUpdateN(agree)
}

// sendAgreeUpdateNforDel broadcasts the AgreeUpdateN message to other notify
// that it agree update the View & N as deleting a node.
func (rbft *rbftImpl) sendAgreeUpdateNforDel(key string) error {

	rbft.logger.Debugf("Replica %d try to send update_n after finish del node", rbft.id)

	if atomic.LoadUint32(&rbft.activeView) == 0 {
		rbft.logger.Warningf("Primary %d is in view change, reject the ready_for_n message", rbft.id)
		return nil
	}

	cert := rbft.getDelNodeCert(key)

	if !cert.finishDel {
		rbft.logger.Errorf("Primary %d has not done with delnode for key=%s", rbft.id, key)
		return nil
	}
	delete(rbft.nodeMgr.updateStore, rbft.nodeMgr.updateTarget)
	rbft.stopNewViewTimer()
	atomic.StoreUint32(&rbft.nodeMgr.inUpdatingN, 1)

	// Calculate the new N and view
	n, view := rbft.getDelNV(cert.delId)

	agree := &AgreeUpdateN{
		Flag:      false,
		ReplicaId: rbft.id,
		Key:       key,
		N:         n,
		View:      view,
		H:         rbft.h,
	}

	rbft.agreeUpdateHelper(agree)
	rbft.logger.Infof("Replica %d sending agree-update-n, v:%d, h:%d, |C|:%d, |P|:%d, |Q|:%d",
		rbft.id, agree.View, agree.H, len(agree.Cset), len(agree.Pset), len(agree.Qset))

	payload, err := proto.Marshal(agree)
	if err != nil {
		rbft.logger.Errorf("ConsensusMessage_AGREE_UPDATE_N Marshal Error", err)
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_AGREE_UPDATE_N,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, rbft.id)
	rbft.helper.InnerBroadcast(msg)
	rbft.recvAgreeUpdateN(agree)
	return nil
}

// recvAgreeUpdateN handles the AgreeUpdateN message sent by others,
// checks the correctness and judges if it can move on to QuorumEvent.
func (rbft *rbftImpl) recvAgreeUpdateN(agree *AgreeUpdateN) events.Event {

	rbft.logger.Debugf("Replica %d received agree-update-n from replica %d, v:%d, n:%d, flag:%v, h:%d, |C|:%d, |P|:%d, |Q|:%d",
		rbft.id, agree.ReplicaId, agree.View, agree.N, agree.Flag, agree.H, len(agree.Cset), len(agree.Pset), len(agree.Qset))

	// Reject response to updating N as replica is in viewChange, negoView or recovery
	if atomic.LoadUint32(&rbft.activeView) == 0 {
		rbft.logger.Warningf("Replica %d try to recvAgreeUpdateN, but it's in view-change", rbft.id)
		return nil
	}
	if rbft.status.getState(&rbft.status.inNegoView) {
		rbft.logger.Warningf("Replica %d try to recvAgreeUpdateN, but it's in nego-view", rbft.id)
		return nil
	}
	if rbft.status.getState(&rbft.status.inRecovery) {
		rbft.logger.Warningf("Replica %d try to recvAgreeUpdateN, but it's in recovery", rbft.id)
		return nil
	}

	// Check if the AgreeUpdateN message is valid or not
	if !rbft.checkAgreeUpdateN(agree) {
		rbft.logger.Debugf("Replica %d found agree-update-n message incorrect", rbft.id)
		return nil
	}

	key := aidx{
		v:    agree.View,
		n:    agree.N,
		id:   agree.ReplicaId,
		flag: agree.Flag,
	}

	// Cast the vote of AgreeUpdateN into an existing or new tally
	if _, ok := rbft.nodeMgr.agreeUpdateStore[key]; ok {
		rbft.logger.Warningf("Replica %d already has a agree-update-n message"+
			" for view=%d/n=%d from replica %d", rbft.id, agree.View, agree.N, agree.ReplicaId)
		return nil
	}
	rbft.nodeMgr.agreeUpdateStore[key] = agree

	// Count of the amount of AgreeUpdateN message for the same key
	replicas := make(map[uint64]bool)
	for idx := range rbft.nodeMgr.agreeUpdateStore {
		if !(idx.v == agree.View && idx.n == agree.N && idx.flag == agree.Flag) {
			continue
		}
		replicas[idx.id] = true
	}
	quorum := len(replicas)

	// We only enter this if there are enough agree-update-n messages but locally not inUpdateN
	if agree.Flag && quorum > rbft.oneCorrectQuorum() && atomic.LoadUint32(&rbft.nodeMgr.inUpdatingN) == 0 {
		rbft.logger.Warningf("Replica %d received f+1 agree-update-n messages, triggering sendAgreeUpdateNForAdd",
			rbft.id)
		rbft.timerMgr.stopTimer(FIRST_REQUEST_TIMER)
		agree.ReplicaId = rbft.id
		return rbft.sendAgreeUpdateNForAdd(agree)
	}

	// We only enter this if there are enough agree-update-n messages but locally not inUpdateN
	if !agree.Flag && quorum >= rbft.oneCorrectQuorum() && atomic.LoadUint32(&rbft.nodeMgr.inUpdatingN) == 0 {
		rbft.logger.Warningf("Replica %d received f+1 agree-update-n messages, triggering sendAgreeUpdateNForDel",
			rbft.id)
		rbft.timerMgr.stopTimer(FIRST_REQUEST_TIMER)
		agree.ReplicaId = rbft.id
		return rbft.sendAgreeUpdateNforDel(agree.Key)
	}

	rbft.logger.Debugf("Replica %d now has %d agree-update requests for view=%d/n=%d", rbft.id, quorum, agree.View, agree.N)

	// Quorum of AgreeUpdateN reach the N, replica can jump to NODE_MGR_AGREE_UPDATEN_QUORUM_EVENT,
	// which mean all nodes agree in updating N
	if quorum >= rbft.allCorrectReplicasQuorum() {
		rbft.nodeMgr.updateTarget = uidx{v: agree.View, n: agree.N, flag: agree.Flag, key: agree.Key}
		return &LocalEvent{
			Service:   NODE_MGR_SERVICE,
			EventType: NODE_MGR_AGREE_UPDATEN_QUORUM_EVENT,
		}
	}

	return nil
}

// sendUpdateN broadcasts the UpdateN message to other, it will only be called
// by primary like the NewView.
func (rbft *rbftImpl) sendUpdateN() events.Event {

	if rbft.status.getState(&rbft.status.inNegoView) {
		rbft.logger.Debugf("Replica %d try to sendUpdateN, but it's in nego-view", rbft.id)
		return nil
	}

	// Reject repeatedly broadcasting of UpdateN message
	if _, ok := rbft.nodeMgr.updateStore[rbft.nodeMgr.updateTarget]; ok {
		rbft.logger.Debugf("Replica %d already has update in store for n=%d/view=%d, skipping", rbft.id, rbft.nodeMgr.updateTarget.n, rbft.nodeMgr.updateTarget.v)
		return nil
	}

	aset := rbft.getAgreeUpdates()

	// Check if primary can find the initial checkpoint for updating
	cp, ok, replicas := rbft.selectInitialCheckpointForUpdate(aset)
	if !ok {
		rbft.logger.Infof("Replica %d could not find consistent checkpoint: %+v", rbft.id, rbft.vcMgr.viewChangeStore)
		return nil
	}

	// Assign the seqNo according to the set of AgreeUpdateN and the initial checkpoint
	msgList := rbft.assignSequenceNumbersForUpdate(aset, cp.SequenceNumber)
	if msgList == nil {
		rbft.logger.Infof("Replica %d could not assign sequence numbers for new view", rbft.id)
		return nil
	}

	update := &UpdateN{
		Flag:      rbft.nodeMgr.updateTarget.flag,
		N:         rbft.nodeMgr.updateTarget.n,
		View:      rbft.nodeMgr.updateTarget.v,
		Key:       rbft.nodeMgr.updateTarget.key,
		Aset:      aset,
		Xset:      msgList,
		ReplicaId: rbft.id,
	}

	rbft.logger.Infof("Replica %d is new primary, sending update-n, v:%d, X:%+v",
		rbft.id, update.View, update.Xset)
	payload, err := proto.Marshal(update)
	if err != nil {
		rbft.logger.Errorf("ConsensusMessage_UPDATE_N Marshal Error", err)
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_UPDATE_N,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, rbft.id)
	rbft.helper.InnerBroadcast(msg)
	rbft.nodeMgr.updateStore[rbft.nodeMgr.updateTarget] = update
	return rbft.primaryProcessUpdateN(cp, replicas, update)
}

// recvUpdateN handles the UpdateN message sent by primary.
func (rbft *rbftImpl) recvUpdateN(update *UpdateN) events.Event {

	rbft.logger.Infof("Replica %d received update-n from replica %d",
		rbft.id, update.ReplicaId)

	// Reject response to updating N as replica is in viewChange, negoView or recovery
	if atomic.LoadUint32(&rbft.activeView) == 0 {
		rbft.logger.Warningf("Replica %d try to recvUpdateN, but it's in view-change", rbft.id)
		return nil
	}
	if rbft.status.getState(&rbft.status.inNegoView) {
		rbft.logger.Debugf("Replica %d try to recvUpdateN, but it's in nego-view", rbft.id)
		return nil
	}
	if rbft.status.getState(&rbft.status.inRecovery) {
		rbft.logger.Noticef("Replica %d try to recvUpdateN, but it's in recovery", rbft.id)
		rbft.recoveryMgr.recvNewViewInRecovery = true
		return nil
	}

	// UpdateN can only be sent by primary
	if !(update.View >= 0 && rbft.isPrimary(update.ReplicaId)) {
		rbft.logger.Infof("Replica %d rejecting invalid UpdateN from %d, v:%d",
			rbft.id, update.ReplicaId, update.View)
		return nil
	}

	key := uidx{
		flag: update.Flag,
		key:  update.Key,
		n:    update.N,
		v:    update.View,
	}
	rbft.nodeMgr.updateStore[key] = update

	// Count of the amount of AgreeUpdateN message for the same key
	quorum := 0
	for idx := range rbft.nodeMgr.agreeUpdateStore {
		if idx.v == update.View {
			quorum++
		}
	}
	// Reject to process UpdateN if replica has not reach allCorrectReplicasQuorum
	if quorum < rbft.allCorrectReplicasQuorum() {
		rbft.logger.Warningf("Replica %d has not meet agreeUpdateNQuorum", rbft.id)
		return nil
	}

	return rbft.processUpdateN()
}

// primaryProcessUpdateN processes the UpdateN message after it has already reached
// updateN-quorum.
func (rbft *rbftImpl) primaryProcessUpdateN(initialCp ViewChange_C, replicas []replicaInfo, update *UpdateN) events.Event {

	// Check if primary need state update
	err := rbft.checkIfNeedStateUpdate(initialCp, replicas)
	if err != nil {
		rbft.logger.Error(err.Error())
		return nil
	}

	// Check if primary need fetch missing requests
	newReqBatchMissing := rbft.feedMissingReqBatchIfNeeded(update.Xset)
	if len(rbft.storeMgr.missingReqBatches) == 0 {
		return rbft.processReqInUpdate(update)
	} else if newReqBatchMissing {
		rbft.fetchRequestBatches()
	}

	return nil
}

// processUpdateN handles the UpdateN message sent from primary, it can only be called
// once replica has reached update-quorum.
func (rbft *rbftImpl) processUpdateN() events.Event {

	// Get the UpdateN from the local cache
	update, ok := rbft.nodeMgr.updateStore[rbft.nodeMgr.updateTarget]
	if !ok {
		rbft.logger.Debugf("Replica %d ignoring processUpdateN as it could not find n=%d/view=%d in its updateStore", rbft.id, rbft.nodeMgr.updateTarget.n, rbft.nodeMgr.updateTarget.v)
		return nil
	}

	if atomic.LoadUint32(&rbft.activeView) == 0 {
		rbft.logger.Infof("Replica %d ignoring update-n from %d, v:%d: we are in view change to view=%d",
			rbft.id, update.ReplicaId, update.View, rbft.view)
		return nil
	}

	if atomic.LoadUint32(&rbft.nodeMgr.inUpdatingN) == 0 {
		rbft.logger.Infof("Replica %d ignoring update-n from %d, v:%d: we are not in updatingN",
			rbft.id, update.ReplicaId, update.View)
		return nil
	}

	// Find the initial checkpoint
	// TODO: use the aset sent from primary is not suitable, use aset itself has received.
	cp, ok, replicas := rbft.selectInitialCheckpointForUpdate(update.Aset)
	if !ok {
		rbft.logger.Warningf("Replica %d could not determine initial checkpoint: %+v",
			rbft.id, rbft.vcMgr.viewChangeStore)
		return rbft.sendViewChange()
	}

	// Check if the xset sent by new primary is built correctly by the aset
	msgList := rbft.assignSequenceNumbersForUpdate(update.Aset, cp.SequenceNumber)
	if msgList == nil {
		rbft.logger.Warningf("Replica %d could not assign sequence numbers: %+v",
			rbft.id, rbft.vcMgr.viewChangeStore)
		return rbft.sendViewChange()
	}
	if !(len(msgList) == 0 && len(update.Xset) == 0) && !reflect.DeepEqual(msgList, update.Xset) {
		rbft.logger.Warningf("Replica %d failed to verify new-view Xset: computed %+v, received %+v",
			rbft.id, msgList, update.Xset)
		return rbft.sendViewChange()
	}

	// Check if primary need state update
	err := rbft.checkIfNeedStateUpdate(cp, replicas)
	if err != nil {
		rbft.logger.Error(err.Error())
		return nil
	}

	return rbft.processReqInUpdate(update)

}

// processReqInUpdate resets all the variables that need to be updated after updating n
func (rbft *rbftImpl) processReqInUpdate(update *UpdateN) events.Event {
	rbft.logger.Debugf("Replica %d accepting update-n to target %v", rbft.id, rbft.nodeMgr.updateTarget)

	if rbft.status.getState(&rbft.status.updateHandled) {
		rbft.logger.Debugf("Replica %d repeated enter processReqInUpdate, ignore it")
		return nil
	}
	rbft.status.activeState(&rbft.status.updateHandled)

	rbft.stopNewViewTimer()
	rbft.timerMgr.stopTimer(NULL_REQUEST_TIMER)

	// Update the view, N and f
	rbft.view = update.View
	rbft.N = int(update.N)
	rbft.f = (rbft.N - 1) / 3

	if !update.Flag {
		cert := rbft.getDelNodeCert(update.Key)
		rbft.id = cert.newId
	}

	// Clean the AgreeUpdateN messages in this turn
	for idx := range rbft.nodeMgr.agreeUpdateStore {
		if idx.v == update.View && idx.n == update.N && idx.flag == update.Flag {
			delete(rbft.nodeMgr.agreeUpdateStore, idx)
		}
	}

	// Clean the AddNode/DelNode messages in this turn
	rbft.nodeMgr.addNodeCertStore = make(map[string]*addNodeCert)
	rbft.nodeMgr.delNodeCertStore = make(map[string]*delNodeCert)

	resetTarget := rbft.exec.lastExec + 1
	if !rbft.status.getState(&rbft.status.skipInProgress) &&
		!rbft.status.getState(&rbft.status.inVcReset) {
		// In most common case, do VcReset
		rbft.helper.VcReset(resetTarget)
		rbft.status.activeState(&rbft.status.inVcReset)
	} else if rbft.isPrimary(rbft.id) {
		// Primary cannot do VcReset then we just let others choose next primary
		// after update timer expired
		rbft.logger.Warningf("New primary %d need to catch up other, waiting", rbft.id)
	} else {
		// Replica can just respone the update no matter what happened
		rbft.logger.Warningf("Replica %d cannot process local vcReset, but also send finishVcReset", rbft.id)
		rbft.sendFinishUpdate()
	}

	return nil
}

// sendFinishUpdate broadcasts the FinisheUpdate message to others to inform that
// it has finished VcReset of updating
func (rbft *rbftImpl) sendFinishUpdate() events.Event {

	finish := &FinishUpdate{
		ReplicaId: rbft.id,
		View:      rbft.view,
		LowH:      rbft.h,
	}

	payload, err := proto.Marshal(finish)
	if err != nil {
		rbft.logger.Errorf("Marshal FinishUpdate Error!")
		return nil
	}
	msg := &ConsensusMessage{
		Type:    ConsensusMessage_FINISH_UPDATE,
		Payload: payload,
	}

	broadcast := cMsgToPbMsg(msg, rbft.id)
	rbft.helper.InnerBroadcast(broadcast)
	rbft.logger.Debugf("Replica %d broadcast FinishUpdate for view=%d, h=%d", rbft.id, rbft.view, rbft.h)

	return rbft.recvFinishUpdate(finish)
}

// recvFinishUpdate handles the FinishUpdate messages sent from others
func (rbft *rbftImpl) recvFinishUpdate(finish *FinishUpdate) events.Event {

	if atomic.LoadUint32(&rbft.nodeMgr.inUpdatingN) == 0 {
		rbft.logger.Debugf("Replica %d is not in updatingN, but received FinishUpdate from replica %d", rbft.id, finish.ReplicaId)
	}
	rbft.logger.Debugf("Replica %d received FinishUpdate from replica %d, view=%d/h=%d", rbft.id, finish.ReplicaId, finish.View, finish.LowH)

	if finish.View != rbft.view {
		rbft.logger.Warningf("Replica %d received FinishUpdate from replica %d, expect view=%d, but get view=%d", rbft.id, finish.ReplicaId, rbft.view, finish.View)
		// We don't ignore it as there may be delay during receiveUpdate from primary between new node and other non-primary old replicas
	}

	ok := rbft.nodeMgr.finishUpdateStore[*finish]
	if ok {
		rbft.logger.Warningf("Replica %d ignored duplicator agree FinishUpdate from %d", rbft.id, finish.ReplicaId)
		return nil
	}
	rbft.nodeMgr.finishUpdateStore[*finish] = true

	return rbft.handleTailAfterUpdate()
}

// handleTailAfterUpdate handles the tail after stable checkpoint
func (rbft *rbftImpl) handleTailAfterUpdate() events.Event {

	// Get the quorum of FinishUpdate messages, and if has new primary's
	quorum := 0
	hasPrimary := false
	for finish := range rbft.nodeMgr.finishUpdateStore {
		if finish.View == rbft.view {
			quorum++
			if rbft.isPrimary(finish.ReplicaId) {
				hasPrimary = true
			}
		}
	}

	// We can only go on if we have arrived quorum, and we have primary's message
	if quorum < rbft.allCorrectReplicasQuorum() {
		rbft.logger.Debugf("Replica %d doesn't has arrive quorum FinishUpdate, expect=%d, got=%d",
			rbft.id, rbft.allCorrectReplicasQuorum(), quorum)
		return nil
	}
	if !hasPrimary {
		rbft.logger.Debugf("Replica %d doesn't receive FinishUpdate from primary %d",
			rbft.id, rbft.primary(rbft.view))
		return nil
	}

	if rbft.status.getState(&rbft.status.inVcReset) {
		rbft.logger.Warningf("Replica %d itself has not finished update", rbft.id)
		return nil
	}

	update, ok := rbft.nodeMgr.updateStore[rbft.nodeMgr.updateTarget]
	if !ok {
		rbft.logger.Debugf("Primary %d ignoring handleTailAfterUpdate as it could not find target %v in its updateStore", rbft.id, rbft.nodeMgr.updateTarget)
		return nil
	}

	rbft.stopNewViewTimer()

	// Reset the seqNo and vid
	rbft.seqNo = rbft.exec.lastExec
	rbft.batchVdr.lastVid = rbft.exec.lastExec

	// New Primary validate the batch built by xset
	if rbft.isPrimary(rbft.id) {
		xSetLen := len(update.Xset)
		upper := uint64(xSetLen) + rbft.h + uint64(1)
		for i := rbft.h + uint64(1); i < upper; i++ {
			d, ok := update.Xset[i]
			if !ok {
				rbft.logger.Critical("update_n Xset miss batch number %d", i)
			} else if d == "" {
				// This should not happen
				rbft.logger.Critical("update_n Xset has null batch, kick it out")
			} else {
				batch, ok := rbft.storeMgr.txBatchStore[d]
				if !ok {
					rbft.logger.Criticalf("In Xset %s exists, but in Replica %d validatedBatchStore there is no such batch digest", d, rbft.id)
				} else if i > rbft.seqNo {
					rbft.primaryValidateBatch(d, batch, i)
				}
			}
		}
	}

	return &LocalEvent{
		Service:   NODE_MGR_SERVICE,
		EventType: NODE_MGR_UPDATEDN_EVENT,
	}
}

// rebuildCertStoreForUpdate rebuilds the certStore after finished updating,
// according to the Xset
func (rbft *rbftImpl) rebuildCertStoreForUpdate() {

	update, ok := rbft.nodeMgr.updateStore[rbft.nodeMgr.updateTarget]
	if !ok {
		rbft.logger.Debugf("Primary %d ignoring rebuildCertStore as it could not find target %v in its updateStore", rbft.id, rbft.nodeMgr.updateTarget)
		return
	}
	rbft.rebuildCertStore(update.Xset)

}

// agreeUpdateHelper helps generate the AgreeUpdateN message.
func (rbft *rbftImpl) agreeUpdateHelper(agree *AgreeUpdateN) {

	pset := rbft.calcPSet()
	qset := rbft.calcQSet()

	rbft.vcMgr.plist = pset
	rbft.vcMgr.qlist = qset

	for idx := range rbft.nodeMgr.agreeUpdateStore {
		if !(idx.v == agree.View && idx.n == agree.N && idx.flag == agree.Flag) {
			delete(rbft.nodeMgr.agreeUpdateStore, idx)
		}
	}

	// Gather all the checkpoints
	for n, id := range rbft.storeMgr.chkpts {
		agree.Cset = append(agree.Cset, &ViewChange_C{
			SequenceNumber: n,
			Id:             id,
		})
	}

	// Gather all the p entries
	for _, p := range pset {
		if p.SequenceNumber < rbft.h {
			rbft.logger.Errorf("BUG! Replica %d should not have anything in our pset less than h, found %+v", rbft.id, p)
		}
		agree.Pset = append(agree.Pset, p)
	}

	// Gather all the q entries
	for _, q := range qset {
		if q.SequenceNumber < rbft.h {
			rbft.logger.Errorf("BUG! Replica %d should not have anything in our qset less than h, found %+v", rbft.id, q)
		}
		agree.Qset = append(agree.Qset, q)
	}
}

// checkAgreeUpdateN checks the AgreeUpdateN message if it's valid or not.
func (rbft *rbftImpl) checkAgreeUpdateN(agree *AgreeUpdateN) bool {

	if agree.Flag {
		// Get the add-node cert
		cert := rbft.getAddNodeCert(agree.Key)
		if !rbft.status.getState(&rbft.status.isNewNode) && !cert.finishAdd {
			rbft.logger.Debugf("Replica %d has not complete add node", rbft.id)
			return false
		}

		// Check the N and view after updating
		n, view := rbft.getAddNV()
		if n != agree.N || view != agree.View {
			rbft.logger.Debugf("Replica %d invalid p entry in agree-update: expected n=%d/view=%d, get n=%d/view=%d", rbft.id, n, view, agree.N, agree.View)
			return false
		}

		// Check if there's any invalid p or q entry
		for _, p := range append(agree.Pset, agree.Qset...) {
			if !(p.View <= agree.View && p.SequenceNumber > agree.H && p.SequenceNumber <= agree.H+rbft.L) {
				rbft.logger.Debugf("Replica %d invalid p entry in agree-update: agree(v:%d h:%d) p(v:%d n:%d)", rbft.id, agree.View, agree.H, p.View, p.SequenceNumber)
				return false
			}
		}

	} else {
		// Get the del-node cert
		cert := rbft.getDelNodeCert(agree.Key)
		if !cert.finishDel {
			rbft.logger.Debugf("Replica %d has not complete del node", rbft.id)
			return false
		}

		// Check the N and view after updating
		n, view := rbft.getDelNV(cert.delId)
		if n != agree.N || view != agree.View {
			rbft.logger.Debugf("Replica %d invalid p entry in agree-update: expected n=%d/view=%d, get n=%d/view=%d", rbft.id, n, view, agree.N, agree.View)
			return false
		}

		// Check if there's any invalid p or q entry
		for _, p := range append(agree.Pset, agree.Qset...) {
			if !(p.View <= agree.View+1 && p.SequenceNumber > agree.H && p.SequenceNumber <= agree.H+rbft.L) {
				rbft.logger.Debugf("Replica %d invalid p entry in agree-update: agree(v:%d h:%d) p(v:%d n:%d)", rbft.id, agree.View, agree.H, p.View, p.SequenceNumber)
				return false
			}
		}

	}

	// Check if there's invalid checkpoint
	for _, c := range agree.Cset {
		if !(c.SequenceNumber >= agree.H && c.SequenceNumber <= agree.H+rbft.L) {
			rbft.logger.Warningf("Replica %d invalid c entry in agree-update: agree(v:%d h:%d) c(n:%d)", rbft.id, agree.View, agree.H, c.SequenceNumber)
			return false
		}
	}

	return true
}

// checkIfNeedStateUpdate checks if a replica need to do state update
func (rbft *rbftImpl) checkIfNeedStateUpdate(initialCp ViewChange_C, replicas []replicaInfo) error {

	speculativeLastExec := rbft.exec.lastExec
	if rbft.exec.currentExec != nil {
		speculativeLastExec = *rbft.exec.currentExec
	}

	if rbft.h < initialCp.SequenceNumber {
		rbft.moveWatermarks(initialCp.SequenceNumber)
	}

	// If replica's lastExec < initial checkpoint, replica is out of data
	if speculativeLastExec < initialCp.SequenceNumber {
		rbft.logger.Warningf("Replica %d missing base checkpoint %d (%s), our most recent execution %d", rbft.id, initialCp.SequenceNumber, initialCp.Id, speculativeLastExec)

		snapshotID, err := base64.StdEncoding.DecodeString(initialCp.Id)
		if nil != err {
			err = fmt.Errorf("Replica %d received a agree-update-n whose hash could not be decoded (%s)", rbft.id, initialCp.Id)
			return err
		}

		// Build the new target of state update
		target := &stateUpdateTarget{
			checkpointMessage: checkpointMessage{
				seqNo: initialCp.SequenceNumber,
				id:    snapshotID,
			},
			replicas: replicas,
		}

		rbft.updateHighStateTarget(target)
		rbft.stateTransfer(target)
	}

	return nil
}

// getAgreeUpdates gets all the AgreeUpdateN messages the replica received.
func (rbft *rbftImpl) getAgreeUpdates() (agrees []*AgreeUpdateN) {
	for _, agree := range rbft.nodeMgr.agreeUpdateStore {
		agrees = append(agrees, agree)
	}
	return
}

// selectInitialCheckpointForUpdate selects the highest checkpoint
// that reached f+1 quorum.
func (rbft *rbftImpl) selectInitialCheckpointForUpdate(aset []*AgreeUpdateN) (checkpoint ViewChange_C, ok bool, replicas []replicaInfo) {

	// For the checkpoint as key, find the corresponding AgreeUpdateN messages
	checkpoints := make(map[ViewChange_C][]*AgreeUpdateN)
	for _, agree := range aset {
		// Verify that we strip duplicate checkpoints from this Cset
		set := make(map[ViewChange_C]bool)
		for _, c := range agree.Cset {
			if ok := set[*c]; ok {
				continue
			}
			checkpoints[*c] = append(checkpoints[*c], agree)
			set[*c] = true
			rbft.logger.Debugf("Replica %d appending checkpoint from replica %d with seqNo=%d, h=%d, and checkpoint digest %s", rbft.id, agree.ReplicaId, agree.H, c.SequenceNumber, c.Id)
		}
	}

	// Indicate that replica cannot find any checkpoint
	if len(checkpoints) == 0 {
		rbft.logger.Debugf("Replica %d has no checkpoints to select from: %d %s",
			rbft.id, len(rbft.vcMgr.viewChangeStore), checkpoints)
		return
	}

	for idx, vcList := range checkpoints {
		// Need weak certificate for the checkpoint
		if len(vcList) < rbft.oneCorrectQuorum() { // type casting necessary to match types
			rbft.logger.Debugf("Replica %d has no weak certificate for n:%d, vcList was %d long",
				rbft.id, idx.SequenceNumber, len(vcList))
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

		if quorum < rbft.commonCaseQuorum() {
			rbft.logger.Debugf("Replica %d has no quorum for n:%d", rbft.id, idx.SequenceNumber)
			continue
		}

		// Find the highest checkpoint
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

// assignSequenceNumbersForUpdate assigns the new seqNo of each valid batch in new view:
// 1. finds all the batch that both prepared and pre-prepared;
// 2. assigns the seqNo to batches ignoring the invalid ones.
// Example: we have 1, 3, 4, 5, 6 prepared and pre-prepared, we get {1, 2, 3, 4, 5}
func (rbft *rbftImpl) assignSequenceNumbersForUpdate(aset []*AgreeUpdateN, h uint64) map[uint64]string {
	msgList := make(map[uint64]string)

	maxN := h + 1

	// "for all n such that h < n <= h + L"
nLoop:
	for n := h + 1; n <= h+rbft.L; n++ {
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

				if quorum < rbft.commonCaseQuorum() {
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

				if quorum < rbft.oneCorrectQuorum() {
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

		if quorum >= rbft.commonCaseQuorum() {
			// "then select the null request for number n"
			msgList[n] = ""

			continue nLoop
		}

		rbft.logger.Warningf("Replica %d could not assign value to contents of seqNo %d, found only %d missing P entries", rbft.id, n, quorum)
		return nil
	}

	// Prune top null requests
	for n, msg := range msgList {
		if n > maxN || msg == "" {
			delete(msgList, n)
		}
	}

	// Sort the msgList, as it's random ranged
	keys := make([]uint64, len(msgList))
	i := 0
	for n := range msgList {
		keys[i] = n
		i++
	}
	sort.Sort(sortableUint64Slice(keys))

	// Sequentially assign the seqNo
	x := h + 1
	list := make(map[uint64]string)
	for _, n := range keys {
		list[x] = msgList[n]
		x++
	}

	return list
}