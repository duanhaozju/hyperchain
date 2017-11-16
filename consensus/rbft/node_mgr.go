//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import (
	"encoding/base64"
	"reflect"
	"sort"
	"time"

	"github.com/hyperchain/hyperchain/manager/protos"

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

	routers           []byte                 // track the vp replicas' routers
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

	return nm
}

// dispatchNodeMgrMsg dispatches node manager service messages from other peers
// and uses corresponding function to handle them.
func (rbft *rbftImpl) dispatchNodeMgrMsg(e consensusEvent) consensusEvent {
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

// recvLocalNewNode handles the local consensusEvent about NewNode, which announces the
// consentor that it's a new node.
func (rbft *rbftImpl) recvLocalNewNode(msg *protos.NewNodeMessage) error {

	rbft.logger.Debugf("New replica %d received local newNode message", rbft.id)

	if rbft.in(isNewNode) {
		rbft.logger.Warningf("New replica %d received duplicate local newNode message", rbft.id)
		return nil
	}

	// the key about new node cannot be nil, it will results failure of updateN
	if len(msg.Payload) == 0 {
		rbft.logger.Warningf("New replica %d received nil local newNode message", rbft.id)
		return nil
	}

	rbft.on(isNewNode, inAddingNode)

	// new node store the new node state in db, since it may crash after connecting
	// with other nodes but before it truly participating in consensus
	rbft.persistNewNode(uint64(1))

	// the key of new node should be stored, since it will be use in processUpdateN
	key := string(msg.Payload)
	rbft.nodeMgr.localKey = key
	rbft.persistLocalKey(msg.Payload)

	return nil
}

// recvLocalAddNode handles the local consensusEvent about AddNode, which announces the
// consentor that a new node want to participate in the consensus as a VP node.
func (rbft *rbftImpl) recvLocalAddNode(msg *protos.AddNodeMessage) error {

	key := string(msg.Payload)
	rbft.logger.Debugf("Replica %d received local addNode message for new node %v", rbft.id, key)

	if rbft.in(isNewNode) {
		rbft.logger.Warningf("New replica received local addNode message, there may be something wrong")
		return nil
	}

	if len(msg.Payload) == 0 {
		rbft.logger.Warningf("Replica %d received nil local addNode message", rbft.id)
		return nil
	}

	rbft.on(inAddingNode)
	rbft.sendAgreeAddNode(key)

	return nil
}

// recvLocalAddNode handles the local consensusEvent about DelNode, which announces the
// consentor that a VP node wants to leave the consensus. If there only exists less than
// 5 VP nodes, we don't allow delete node.
// Notice: the node that wants to leave away from consensus should also handle this message.
func (rbft *rbftImpl) recvLocalDelNode(msg *protos.DelNodeMessage) error {

	key := string(msg.DelPayload)
	rbft.logger.Debugf("Replica %d received local delNode message for newId: %d, del node ID: %d", rbft.id, msg.Id, msg.Del)

	// We only support deleting when number of all VP nodes >= 5
	if rbft.N == 4 {
		rbft.logger.Warningf("Replica %d received delNode message, but we don't support delete as there're only 4 nodes", rbft.id)
		return nil
	}

	if len(msg.DelPayload) == 0 {
		rbft.logger.Warningf("Replica %d received nil local delNode message", rbft.id)
		return nil
	}

	rbft.on(inDeletingNode)
	rbft.sendAgreeDelNode(key, msg.RouterHash, msg.Id, msg.Del)

	return nil
}

// sendAgreeAddNode broadcasts AgreeAddNode message to other replicas to notify
// that itself had received the add-in request from new node.
func (rbft *rbftImpl) sendAgreeAddNode(key string) {

	rbft.logger.Debugf("Replica %d try to send addNode message for new node %s", rbft.id, key)

	// AgreeAddNode message can only sent by original nodes
	if rbft.in(isNewNode) {
		rbft.logger.Warningf("New replica try to send addNode message, there may be something wrong")
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

	rbft.logger.Debugf("Replica %d try to send delNode message for quit node %d", rbft.id, delId)

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

	rbft.logger.Debugf("Replica %d received agree addNode from replica %d", rbft.id, add.ReplicaId)

	// Cast the vote of AgreeAddNode into an existing or new tally
	cert := rbft.getAddNodeCert(add.Key)
	ok := cert.addNodes[*add]
	if ok {
		rbft.logger.Warningf("Replica %d received duplicate agree addNode from replica %d, replace it", rbft.id, add.ReplicaId)
	}
	cert.addNodes[*add] = true

	return rbft.maybeUpdateTableForAdd(add.Key)
}

// recvAgreeDelNode handles the AgreeDelNode message sent by others
func (rbft *rbftImpl) recvAgreeDelNode(del *DelNode) error {

	rbft.logger.Debugf("Replica %d received agree delNode from replica %d", rbft.id, del.ReplicaId)

	// Cast the vote of AgreeDelNode into an existing or new tally
	cert := rbft.getDelNodeCert(del.Key)
	ok := cert.delNodes[*del]
	if ok {
		rbft.logger.Warningf("Replica %d received duplicate agree delNode from replica %d, replace it", rbft.id, del.ReplicaId)
	}
	cert.delNodes[*del] = true

	return rbft.maybeUpdateTableForDel(del.Key)
}

// maybeUpdateTableForAdd checks if the AgreeAddNode messages have reach
// the quorum or not. If yes, update the routing table to add the connect with
// the new node
func (rbft *rbftImpl) maybeUpdateTableForAdd(key string) error {

	// New node needn't to update the routing table since it already has the newest one
	if rbft.in(isNewNode) {
		rbft.logger.Debugf("Replica %d is a new node, reject update routingTable", rbft.id)
		return nil
	}

	// Adding node stipulates all nodes should come into agreement
	cert := rbft.getAddNodeCert(key)
	if len(cert.addNodes) < rbft.allCorrectQuorum() {
		return nil
	}

	if !rbft.in(inAddingNode) {
		if cert.finishAdd {
			// This indicates a byzantine behavior that
			// some nodes repeatedly send AgreeAddNode messages.
			rbft.logger.Warningf("Replica %d has already finished addingNode, but still received addNode msg from someone else", rbft.id)
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
	rbft.off(inAddingNode)

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

	if !rbft.in(inDeletingNode) {
		if cert.finishDel {
			// This indicates a byzantine behavior that
			// some nodes repeatedly send AgreeDelNode messages.
			rbft.logger.Warningf("Replica %d has already finished deletingNode, but still recevice del msg from someone else", rbft.id)
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
	rbft.off(inDeletingNode)
	rbft.on(inUpdatingN)

	// As for deleting node, replicas broadcast AgreeUpdateN just after
	// updating routing table
	rbft.sendAgreeUpdateNforDel(key)

	return nil
}

// sendReadyForN broadcasts the ReadyForN message to others.
// Only new node will call this after finished recovery, ReadyForN message means that
// it has already caught up with others and wants to truly participate in the consensus.
func (rbft *rbftImpl) sendReadyForN() error {

	if !rbft.in(isNewNode) {
		rbft.logger.Errorf("Replica %d isn't a new replica, but try to send readyForN", rbft.id)
		return nil
	}

	// If new node loses the local key, there may be something wrong
	if rbft.nodeMgr.localKey == "" {
		rbft.logger.Errorf("New replica %d doesn't have local key for readyForN", rbft.id)
		return nil
	}

	if rbft.in(inViewChange) {
		rbft.logger.Errorf("New replica %d finds itself in viewChange, not sending readyForN", rbft.id)
		return nil
	}

	rbft.logger.Infof("Replica %d sending readyForN as it already finished recovery", rbft.id)
	rbft.on(inUpdatingN)

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
func (rbft *rbftImpl) recvReadyforNforAdd(ready *ReadyForN) consensusEvent {

	rbft.logger.Debugf("Replica %d received readyForN from new node %d", rbft.id, ready.ReplicaId)

	if rbft.in(inViewChange) {
		rbft.logger.Warningf("Replica %d is in viewChange, reject the readyForN message", rbft.id)
		return nil
	}

	cert := rbft.getAddNodeCert(ready.Key)

	if !cert.finishAdd {
		rbft.logger.Debugf("Replica %d has not done with addNode for key=%s", rbft.id, ready.Key)
		return nil
	}

	// Calculate the new N and view
	n, view := rbft.getAddNV()

	// Broadcast the AgreeUpdateN message
	agree := &AgreeUpdateN{
		Basis: &VcBasis{
			ReplicaId: rbft.id,
			View:      view,
			H:         rbft.h,
		},
		Flag: true,
		Key:  ready.Key,
		N:    n,
	}

	return rbft.sendAgreeUpdateNForAdd(agree)
}

// sendAgreeUpdateNForAdd broadcasts the AgreeUpdateN message to others.
// This will be only called after receiving the ReadyForN message sent by new node.
func (rbft *rbftImpl) sendAgreeUpdateNForAdd(agree *AgreeUpdateN) consensusEvent {

	rbft.logger.Debugf("Replica %d try to send agree updateN for add", rbft.id)

	if rbft.in(inUpdatingN) {
		rbft.logger.Debugf("Replica %d already in updatingN, don't send agreeUpdateN again")
		return nil
	}

	if rbft.in(isNewNode) {
		rbft.logger.Warningf("New replica %d does not need to send agreeUpdateN", rbft.id)
		return nil
	}

	// Replica may receive ReadyForN after it has already finished updatingN
	// (it happens in bad network environment)
	if int(agree.N) == rbft.N && agree.Basis.View == rbft.view {
		rbft.logger.Debugf("Replica %d already finished updateN for N=%d/view=%d", rbft.id, rbft.N, rbft.view)
		return nil
	}

	delete(rbft.nodeMgr.updateStore, rbft.nodeMgr.updateTarget)
	rbft.stopNewViewTimer()
	rbft.on(inUpdatingN)

	// Generate the AgreeUpdateN message and broadcast it to others
	rbft.agreeUpdateHelper(agree)
	rbft.logger.Debugf("Replica %d sending agreeUpdateN, v:%d, h:%d, |C|:%d, |P|:%d, |Q|:%d",
		rbft.id, agree.Basis.View, agree.Basis.H, len(agree.Basis.Cset), len(agree.Basis.Pset), len(agree.Basis.Qset))

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

	rbft.logger.Debugf("Replica %d try to send agree updateN for delete after finished delNode", rbft.id)

	if rbft.in(inViewChange) {
		rbft.logger.Warningf("Replica %d is in viewChange, reject sending agreeUpdateN", rbft.id)
		return nil
	}

	cert := rbft.getDelNodeCert(key)

	if !cert.finishDel {
		rbft.logger.Warningf("Replica %d has not done with delNode for key=%s", rbft.id, key)
		return nil
	}
	delete(rbft.nodeMgr.updateStore, rbft.nodeMgr.updateTarget)
	rbft.stopNewViewTimer()
	rbft.on(inUpdatingN)

	// Calculate the new N and view
	n, view := rbft.getDelNV(cert.delId)

	agree := &AgreeUpdateN{
		Basis: &VcBasis{
			ReplicaId: rbft.id,
			View:      view,
			H:         rbft.h,
		},
		Flag: false,
		Key:  key,
		N:    n,
	}

	rbft.agreeUpdateHelper(agree)
	rbft.logger.Debugf("Replica %d sending agreeUpdateN, v:%d, h:%d, |C|:%d, |P|:%d, |Q|:%d",
		rbft.id, agree.Basis.View, agree.Basis.H, len(agree.Basis.Cset), len(agree.Basis.Pset), len(agree.Basis.Qset))

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
// checks the correctness and judges if it can move on to QuorumconsensusEvent.
func (rbft *rbftImpl) recvAgreeUpdateN(agree *AgreeUpdateN) consensusEvent {

	rbft.logger.Debugf("Replica %d received agreeUpdateN from replica %d, v:%d, n:%d, flag:%v, h:%d, |C|:%d, |P|:%d, |Q|:%d",
		rbft.id, agree.Basis.ReplicaId, agree.Basis.View, agree.N, agree.Flag, agree.Basis.H, len(agree.Basis.Cset), len(agree.Basis.Pset), len(agree.Basis.Qset))

	// Reject response to updating N as replica is in viewChange, negoView or recovery
	if rbft.in(inViewChange) {
		rbft.logger.Infof("Replica %d try to receive AgreeUpdateN, but it's in viewChange", rbft.id)
		return nil
	}
	if rbft.in(inNegotiateView) {
		rbft.logger.Infof("Replica %d try to receive AgreeUpdateN, but it's in negotiateView", rbft.id)
		return nil
	}
	if rbft.in(inRecovery) {
		rbft.logger.Infof("Replica %d try to receive AgreeUpdateN, but it's in recovery", rbft.id)
		return nil
	}

	// Check if the AgreeUpdateN message is valid or not
	if !rbft.checkAgreeUpdateN(agree) {
		rbft.logger.Debugf("Replica %d found agreeUpdateN message incorrect", rbft.id)
		return nil
	}

	key := aidx{
		v:    agree.Basis.View,
		n:    agree.N,
		id:   agree.Basis.ReplicaId,
		flag: agree.Flag,
	}

	// Cast the vote of AgreeUpdateN into an existing or new tally
	if _, ok := rbft.nodeMgr.agreeUpdateStore[key]; ok {
		rbft.logger.Warningf("Replica %d already has a agreeUpdateN message"+
			" for view=%d/n=%d from replica %d", rbft.id, agree.Basis.View, agree.N, agree.Basis.ReplicaId)
		return nil
	}
	rbft.nodeMgr.agreeUpdateStore[key] = agree

	// Count of the amount of AgreeUpdateN message for the same key
	replicas := make(map[uint64]bool)
	for idx := range rbft.nodeMgr.agreeUpdateStore {
		if !(idx.v == agree.Basis.View && idx.n == agree.N && idx.flag == agree.Flag) {
			continue
		}
		replicas[idx.id] = true
	}
	quorum := len(replicas)

	// We only enter this if there are enough agree-update-n messages but locally not inUpdateN
	if agree.Flag && quorum > rbft.oneCorrectQuorum() && !rbft.in(inUpdatingN) {
		rbft.logger.Debugf("Replica %d received f+1 agreeUpdateN messages, triggering sendAgreeUpdateNForAdd",
			rbft.id)
		rbft.timerMgr.stopTimer(FIRST_REQUEST_TIMER)
		agree.Basis.ReplicaId = rbft.id
		return rbft.sendAgreeUpdateNForAdd(agree)
	}

	// We only enter this if there are enough agree-update-n messages but locally not inUpdateN
	if !agree.Flag && quorum >= rbft.oneCorrectQuorum() && !rbft.in(inUpdatingN) {
		rbft.logger.Debugf("Replica %d received f+1 agreeUpdateN messages, triggering sendAgreeUpdateNForDel",
			rbft.id)
		rbft.timerMgr.stopTimer(FIRST_REQUEST_TIMER)
		agree.Basis.ReplicaId = rbft.id
		return rbft.sendAgreeUpdateNforDel(agree.Key)
	}

	rbft.logger.Debugf("Replica %d now has %d agreeUpdate requests for view=%d/n=%d", rbft.id, quorum, agree.Basis.View, agree.N)

	// Quorum of AgreeUpdateN reach the N, replica can jump to NODE_MGR_AGREE_UPDATEN_QUORUM_consensusEvent,
	// which mean all nodes agree in updating N
	if quorum >= rbft.allCorrectReplicasQuorum() {
		rbft.nodeMgr.updateTarget = uidx{v: agree.Basis.View, n: agree.N, flag: agree.Flag, key: agree.Key}
		return &LocalEvent{
			Service:   NODE_MGR_SERVICE,
			EventType: NODE_MGR_AGREE_UPDATE_QUORUM_EVENT,
		}
	}

	return nil
}

// sendUpdateN broadcasts the UpdateN message to other, it will only be called
// by primary like the NewView.
func (rbft *rbftImpl) sendUpdateN() consensusEvent {

	if rbft.in(inNegotiateView) {
		rbft.logger.Warningf("Primary %d try to sendUpdateN, but it's in negotiateView", rbft.id)
		return nil
	}

	// Reject repeatedly broadcasting of UpdateN message
	if _, ok := rbft.nodeMgr.updateStore[rbft.nodeMgr.updateTarget]; ok {
		rbft.logger.Debugf("Primary %d already has updateN in store for n=%d/view=%d, skipping", rbft.id, rbft.nodeMgr.updateTarget.n, rbft.nodeMgr.updateTarget.v)
		return nil
	}

	aset := rbft.getAgreeUpdates()

	// Check if primary can find the initial checkpoint for updating
	cp, ok, replicas := rbft.selectInitialCheckpoint(aset)
	if !ok {
		rbft.logger.Infof("Primary %d could not find consistent checkpoint: %+v", rbft.id, rbft.vcMgr.viewChangeStore)
		return nil
	}

	// Assign the seqNo according to the set of AgreeUpdateN and the initial checkpoint
	msgList := rbft.assignSequenceNumbers(aset, cp.SequenceNumber)
	if msgList == nil {
		rbft.logger.Infof("Primary %d could not assign sequence numbers for updateN", rbft.id)
		return nil
	}

	update := &UpdateN{
		Flag:      rbft.nodeMgr.updateTarget.flag,
		N:         rbft.nodeMgr.updateTarget.n,
		View:      rbft.nodeMgr.updateTarget.v,
		Key:       rbft.nodeMgr.updateTarget.key,
		Xset:      msgList,
		ReplicaId: rbft.id,
	}

	rbft.logger.Debugf("Replica %d is new primary, sending updateN, v:%d, X:%+v",
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
	return rbft.primaryCheckUpdateN(cp, replicas, update)
}

// recvUpdateN handles the UpdateN message sent by primary.
func (rbft *rbftImpl) recvUpdateN(update *UpdateN) consensusEvent {

	rbft.logger.Debugf("Replica %d received updateN from replica %d",
		rbft.id, update.ReplicaId)

	// Reject response to updating N as replica is in viewChange, negoView or recovery
	if rbft.in(inViewChange) {
		rbft.logger.Infof("Replica %d try to receive UpdateN, but it's in viewChange", rbft.id)
		return nil
	}
	if rbft.in(inNegotiateView) {
		rbft.logger.Infof("Replica %d try to receive UpdateN, but it's in negotiateView", rbft.id)
		return nil
	}
	if rbft.in(inRecovery) {
		rbft.logger.Infof("Replica %d try to receive UpdateN, but it's in recovery", rbft.id)
		rbft.recoveryMgr.recvNewViewInRecovery = true
		return nil
	}

	// UpdateN can only be sent by primary
	if !(update.View >= 0 && rbft.isPrimary(update.ReplicaId)) {
		rbft.logger.Warningf("Replica %d rejecting invalid updateN from %d, v:%d",
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
	for idx, agree := range rbft.nodeMgr.agreeUpdateStore {
		if idx.v == update.View && idx.n == update.N && idx.flag == update.Flag && agree.Key == update.Key {
			quorum++
		}
	}
	// Reject to process UpdateN if replica has not reach allCorrectReplicasQuorum
	if quorum < rbft.allCorrectReplicasQuorum() {
		rbft.logger.Debugf("Replica %d has not meet agreeUpdateNQuorum", rbft.id)
		return nil
	}

	return rbft.replicaCheckUpdateN()
}

// primaryProcessUpdateN processes the UpdateN message after it has already reached
// updateN-quorum.
func (rbft *rbftImpl) primaryCheckUpdateN(initialCp Vc_C, replicas []replicaInfo, update *UpdateN) consensusEvent {

	// Check if primary need state update
	err := rbft.checkIfNeedStateUpdate(initialCp, replicas)
	if err != nil {
		return nil
	}

	// Check if primary need fetch missing requests
	newReqBatchMissing := rbft.feedMissingReqBatchIfNeeded(update.Xset)
	if len(rbft.storeMgr.missingReqBatches) == 0 {
		return rbft.resetStateForUpdate(update)
	} else if newReqBatchMissing {
		rbft.fetchRequestBatches()
	}

	return nil
}

// processUpdateN handles the UpdateN message sent from primary, it can only be called
// once replica has reached update-quorum.
func (rbft *rbftImpl) replicaCheckUpdateN() consensusEvent {

	// Get the UpdateN from the local cache
	update, ok := rbft.nodeMgr.updateStore[rbft.nodeMgr.updateTarget]
	if !ok {
		rbft.logger.Debugf("Replica %d ignore processUpdateN as it could not find n=%d/view=%d in its updateStore", rbft.id, rbft.nodeMgr.updateTarget.n, rbft.nodeMgr.updateTarget.v)
		return nil
	}

	if rbft.in(inViewChange) {
		rbft.logger.Infof("Replica %d ignore updateN from replica %d, v:%d: we are in viewChange to view=%d",
			rbft.id, update.ReplicaId, update.View, rbft.view)
		return nil
	}

	if !rbft.in(inUpdatingN) {
		rbft.logger.Infof("Replica %d ignore updateN from %replica d, v:%d: we are not in updatingN",
			rbft.id, update.ReplicaId, update.View)
		return nil
	}

	// Find the initial checkpoint
	aset := rbft.getAgreeUpdates()
	cp, ok, replicas := rbft.selectInitialCheckpoint(aset)
	if !ok {
		rbft.logger.Infof("Replica %d could not determine initial checkpoint: %+v",
			rbft.id, rbft.vcMgr.viewChangeStore)
		return rbft.sendViewChange()
	}

	// Check if the xset sent by new primary is built correctly by the aset
	msgList := rbft.assignSequenceNumbers(aset, cp.SequenceNumber)
	if msgList == nil {
		rbft.logger.Infof("Replica %d could not assign sequence numbers: %+v",
			rbft.id, rbft.vcMgr.viewChangeStore)
		return rbft.sendViewChange()
	}
	if !(len(msgList) == 0 && len(update.Xset) == 0) && !reflect.DeepEqual(msgList, update.Xset) {
		rbft.logger.Warningf("Replica %d failed to verify updateN Xset: computed %+v, received %+v",
			rbft.id, msgList, update.Xset)
		return rbft.sendViewChange()
	}

	// Check if primary need state update
	err := rbft.checkIfNeedStateUpdate(cp, replicas)
	if err != nil {
		return nil
	}

	return rbft.resetStateForUpdate(update)

}

// processReqInUpdate resets all the variables that need to be updated after updating n
func (rbft *rbftImpl) resetStateForUpdate(update *UpdateN) consensusEvent {

	rbft.logger.Debugf("Replica %d accept updateN to target %v", rbft.id, rbft.nodeMgr.updateTarget)

	if rbft.in(updateHandled) {
		rbft.logger.Debugf("Replica %d enter processReqInUpdate again, ignore it", rbft.id)
		return nil
	}
	rbft.on(updateHandled)

	rbft.stopNewViewTimer()
	rbft.timerMgr.stopTimer(NULL_REQUEST_TIMER)

	// Update the view, N and f
	rbft.view = update.View
	rbft.N = int(update.N)
	rbft.f = (rbft.N - 1) / 3

	// in adding node, node id will not change as new node will use N+1 as id, and old nodes
	// don't need to change their id.
	// But in deleting node, if deleting node id is not the largest one, old nodes' id need
	// to be rearranged in an ascending order.
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

	// empty the outstandingReqBatch, it is useless since new primary will resend pre-prepare
	rbft.storeMgr.outstandingReqBatches = make(map[string]*TransactionBatch)
	rbft.batchVdr.preparedCert = make(map[vidx]string)
	rbft.storeMgr.committedCert = make(map[msgID]string)

	resetTarget := rbft.exec.lastExec + 1
	if !rbft.inOne(skipInProgress, inVcReset) {
		// in most common case, do VcReset
		rbft.helper.VcReset(resetTarget, rbft.view)
		rbft.on(inVcReset)
	} else if rbft.isPrimary(rbft.id) {
		// Primary cannot do VcReset then we just let others choose next primary
		// after update timer expired
		rbft.logger.Infof("New primary %d need to catch up other, waiting", rbft.id)
	} else {
		// Replica can just response the update no matter what happened
		rbft.logger.Infof("Replica %d cannot process local vcReset, but also send finishVcReset", rbft.id)
		rbft.sendFinishUpdate()
	}

	return nil
}

// sendFinishUpdate broadcasts the FinisheUpdate message to others to inform that
// it has finished VcReset of updating
func (rbft *rbftImpl) sendFinishUpdate() consensusEvent {

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
	rbft.logger.Debugf("Replica %d sending finishUpdate for view=%d, h=%d", rbft.id, rbft.view, rbft.h)

	return rbft.recvFinishUpdate(finish)
}

// recvFinishUpdate handles the FinishUpdate messages sent from others
func (rbft *rbftImpl) recvFinishUpdate(finish *FinishUpdate) consensusEvent {

	if !rbft.in(inUpdatingN) {
		rbft.logger.Debugf("Replica %d is not in updatingN, but received finishUpdate from replica %d", rbft.id, finish.ReplicaId)
	}
	rbft.logger.Debugf("Replica %d received finishUpdate from replica %d, view=%d/h=%d", rbft.id, finish.ReplicaId, finish.View, finish.LowH)

	ok := rbft.nodeMgr.finishUpdateStore[*finish]
	if ok {
		rbft.logger.Warningf("Replica %d ignored duplicator agree finishUpdate from %d", rbft.id, finish.ReplicaId)
		return nil
	}
	rbft.nodeMgr.finishUpdateStore[*finish] = true

	return rbft.processReqInUpdate()
}

// handleTailAfterUpdate handles the tail after stable checkpoint
func (rbft *rbftImpl) processReqInUpdate() consensusEvent {

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
		rbft.logger.Debugf("Replica %d doesn't has arrive quorum finishUpdate, expect=%d, got=%d",
			rbft.id, rbft.allCorrectReplicasQuorum(), quorum)
		return nil
	}
	if !hasPrimary {
		rbft.logger.Debugf("Replica %d doesn't receive finishUpdate from primary %d",
			rbft.id, rbft.primary(rbft.view))
		return nil
	}

	if rbft.in(inVcReset) {
		rbft.logger.Warningf("Replica %d itself has not finishUpdate", rbft.id)
		return nil
	}

	update, ok := rbft.nodeMgr.updateStore[rbft.nodeMgr.updateTarget]
	if !ok {
		rbft.logger.Debugf("Primary %d ignore processReqInUpdate as it could not find target %v in its updateStore", rbft.id, rbft.nodeMgr.updateTarget)
		return nil
	}

	rbft.stopNewViewTimer()

	// Reset the seqNo and vid
	rbft.seqNo = rbft.exec.lastExec
	rbft.batchVdr.setLastVid(rbft.exec.lastExec)

	rbft.putBackTxBatches(update.Xset)

	// New Primary validate the batch built by xset
	if rbft.isPrimary(rbft.id) {
		rbft.primaryResendBatch(update.Xset)
	}

	return &LocalEvent{
		Service:   NODE_MGR_SERVICE,
		EventType: NODE_MGR_UPDATED_EVENT,
	}
}

//##########################################################################
//           node management auxiliary functions
//##########################################################################

// putBackTxBatches put batches who is not in xset back to txPool.
// For previous primary, there may be extra stored batches which haven't prepared.
func (rbft *rbftImpl) putBackTxBatches(xset Xset) {

	var (
		keys         []uint64
		hashList     []string
		initialChkpt uint64
		deleteList   []string
	)

	// Return if xset is nil
	if len(xset) == 0 {
		return
	}

	// Sort the xset
	for no := range xset {
		keys = append(keys, no)
	}
	sort.Sort(sortableUint64Slice(keys))

	// Remove all the batches that smaller than initial checkpoint.
	// Those batches are the dependency of duplicator,
	// but we can remove since we already have checkpoint after viewchange.
	initialChkpt = keys[0] - 1
	for digest, batch := range rbft.storeMgr.txBatchStore {
		if batch.SeqNo <= initialChkpt {
			delete(rbft.storeMgr.txBatchStore, digest)
			rbft.persistDelTxBatch(digest)
			deleteList = append(deleteList, digest)
		}
	}
	rbft.batchMgr.txPool.RemoveBatches(deleteList)

	// Construct hashList
	for _, key := range keys {
		hash := xset[key]
		hashList = append(hashList, hash)
	}

	rbft.batchMgr.txPool.GetBatchesBackExcept(hashList)
}

// rebuildCertStoreForUpdate rebuilds the certStore after finished updating,
// according to the Xset
func (rbft *rbftImpl) rebuildCertStoreForUpdate() {

	update, ok := rbft.nodeMgr.updateStore[rbft.nodeMgr.updateTarget]
	if !ok {
		rbft.logger.Debugf("Primary %d ignore rebuildCertStore as it could not find target %v in its updateStore", rbft.id, rbft.nodeMgr.updateTarget)
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
		if !(idx.v == agree.Basis.View && idx.n == agree.N && idx.flag == agree.Flag) {
			delete(rbft.nodeMgr.agreeUpdateStore, idx)
		}
	}

	agree.Basis.Cset, agree.Basis.Pset, agree.Basis.Qset = rbft.gatherPQC()
}

// checkAgreeUpdateN checks the AgreeUpdateN message if it's valid or not.
func (rbft *rbftImpl) checkAgreeUpdateN(agree *AgreeUpdateN) bool {

	if agree.Flag {
		// Get the add-node cert
		cert := rbft.getAddNodeCert(agree.Key)
		if !rbft.in(isNewNode) && !cert.finishAdd {
			rbft.logger.Debugf("Replica %d has not complete addNode", rbft.id)
			return false
		}

		// Check the N and view after updating
		n, view := rbft.getAddNV()
		if n != agree.N || view != agree.Basis.View {
			rbft.logger.Debugf("Replica %d received incorrect agreeUpdateN: "+
				"expected n=%d/view=%d, get n=%d/view=%d", rbft.id, n, view, agree.N, agree.Basis.View)
			return false
		}

		// Check if there's any invalid p or q entry
		for _, p := range append(agree.Basis.Pset, agree.Basis.Qset...) {
			if !(p.View <= agree.Basis.View && p.SequenceNumber > agree.Basis.H && p.SequenceNumber <= agree.Basis.H+rbft.L) {
				rbft.logger.Debugf("Replica %d received invalid p entry in agreeUpdateN: "+
					"agree(v:%d h:%d) p(v:%d n:%d)", rbft.id, agree.Basis.View, agree.Basis.H, p.View, p.SequenceNumber)
				return false
			}
		}

	} else {
		// Get the del-node cert
		cert := rbft.getDelNodeCert(agree.Key)
		if !cert.finishDel {
			rbft.logger.Debugf("Replica %d has not complete delNode", rbft.id)
			return false
		}

		// Check the N and view after updating
		n, view := rbft.getDelNV(cert.delId)
		if n != agree.N || view != agree.Basis.View {
			rbft.logger.Debugf("Replica %d received incorrect agreeUpdateN: "+
				"expected n=%d/view=%d, get n=%d/view=%d", rbft.id, n, view, agree.N, agree.Basis.View)
			return false
		}

		// Check if there's any invalid p or q entry
		for _, p := range append(agree.Basis.Pset, agree.Basis.Qset...) {
			if !(p.View <= agree.Basis.View+1 && p.SequenceNumber > agree.Basis.H && p.SequenceNumber <= agree.Basis.H+rbft.L) {
				rbft.logger.Debugf("Replica %d received invalid p entry in agreeUpdateN: "+
					"agree(v:%d h:%d) p(v:%d n:%d)", rbft.id, agree.Basis.View, agree.Basis.H, p.View, p.SequenceNumber)
				return false
			}
		}

	}

	// Check if there's invalid checkpoint
	for _, c := range agree.Basis.Cset {
		if !(c.SequenceNumber >= agree.Basis.H && c.SequenceNumber <= agree.Basis.H+rbft.L) {
			rbft.logger.Warningf("Replica %d received invalid c entry in agreeUpdateN: "+
				"agree(v:%d h:%d) c(n:%d)", rbft.id, agree.Basis.View, agree.Basis.H, c.SequenceNumber)
			return false
		}
	}

	return true
}

// checkIfNeedStateUpdate checks if a replica need to do state update
func (rbft *rbftImpl) checkIfNeedStateUpdate(initialCp Vc_C, replicas []replicaInfo) error {

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
			rbft.logger.Errorf("Replica %d received a agreeUpdateN whose hash could not be decoded (%s)", rbft.id, initialCp.Id)
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
func (rbft *rbftImpl) getAgreeUpdates() (aset []*VcBasis) {
	for _, agree := range rbft.nodeMgr.agreeUpdateStore {
		aset = append(aset, agree.Basis)
	}
	return
}
