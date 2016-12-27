//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"hyperchain/protos"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

// New replica receive local NewNode message
func (pbft *pbftProtocal) recvLocalNewNode(msg *protos.NewNodeMessage) error {

	logger.Debugf("New replica %d received local newNode message", pbft.id)

	if pbft.isNewNode {
		logger.Warningf("New replica %d received duplicate local newNode message", pbft.id)
		return errors.New("New replica received duplicate local newNode message")
	}

	if len(msg.Payload) == 0 {
		logger.Warningf("New replica %d received nil local newNode message", pbft.id)
		return errors.New("New replica received nil local newNode message")
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
		return errors.New("New replica received local addNode message")
	}

	if len(msg.Payload) == 0 {
		logger.Warningf("New replica %d received nil local addNode message", pbft.id)
		return errors.New("New replica received nil local addNode message")
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
		return errors.New("Deleting is not supported as there're only 4 nodes")
	}

	if len(msg.DelPayload) == 0 || len(msg.RouterHash) == 0 || msg.Id == 0 {
		logger.Warningf("New replica %d received invalid local delNode message", pbft.id)
		return errors.New("New replica received invalid local delNode message")
	}

	pbft.inDeletingNode = true
	pbft.newid = msg.Id
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

	cert := pbft.getDelNodeCert(key, routerHash)
	cert.newId = newId

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
		return errors.New("Receive duplicate addnode message")
	}

	cert.addNodes[*add] = true
	cert.addCount++

	return pbft.maybeUpdateTableForAdd(add.Key)
}

// Replica received delnode for quit node
func (pbft *pbftProtocal) recvAgreeDelNode(del *DelNode) error {

	logger.Debugf("Replica %d received agree delnode from replica %d for %v",
		pbft.id, del.ReplicaId, del.Key)

	cert := pbft.getDelNodeCert(del.Key, del.RouterHash)

	ok := cert.delNodes[*del]
	if ok {
		logger.Warningf("Replica %d ignored duplicate agree addnode from %d", pbft.id, del.ReplicaId)
		return errors.New("Receive duplicate delnode message")
	}

	cert.delNodes[*del] = true
	cert.delCount++

	return pbft.maybeUpdateTableForDel(del.Key, del.RouterHash)
}

// Check if replica prepared for update routing table after add node
func (pbft *pbftProtocal) maybeUpdateTableForAdd(key string) error {

	cert := pbft.getAddNodeCert(key)

	if cert.addCount < pbft.committedReplicasQuorum() {
		return errors.New("Not enough add message to update table")
	}

	if !pbft.inAddingNode {
		if cert.finishAdd {
			if cert.addCount <= pbft.N {
				logger.Debugf("Replica %d has already finished adding node", pbft.id)
				return errors.New("Replica has already finished adding node")
			} else {
				logger.Warningf("Replica %d has already finished adding node, but still recevice add msg from someone else", pbft.id)
				return errors.New("Replica has already finished adding node, but still recevice add msg from someone else")
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
func (pbft *pbftProtocal) maybeUpdateTableForDel(key string, routerHash string) error {

	cert := pbft.getDelNodeCert(key, routerHash)

	if cert.delCount < pbft.committedReplicasQuorum() {
		return errors.New("Not enough del message to update table")
	}

	if !pbft.inDeletingNode {
		if cert.finishDel {
			if cert.delCount < pbft.N {
				logger.Debugf("Replica %d have already finished deleting node", pbft.id)
				return errors.New("Replica has already finished deleting node")
			} else {
				logger.Warningf("Replica %d has already finished deleting node, but still recevice del msg from someone else", pbft.id)
				return errors.New("Replica has already finished deleting node, but still recevice del msg from someone else")
			}
		}
	}

	cert.finishDel = true
	payload := []byte(key)

	logger.Debugf("Replica %d try to update routing table", pbft.id)
	pbft.helper.UpdateTable(payload, false)
	pbft.inDeletingNode = false
	if pbft.primary(pbft.view) == pbft.id {
		pbft.sendUpdateNforDel(key, routerHash)
	}

	return nil
}

// New replica send ready_for_n to primary after recovery
func (pbft *pbftProtocal) sendReadyForN() error {

	if !pbft.isNewNode {
		logger.Errorf("Replica %d is not new one, but try to send ready_for_n", pbft.id)
		return errors.New("Replica is an old node, but try to send redy_for_n")
	}

	if pbft.localKey == "" {
		logger.Errorf("Replica %d doesn't have local key for ready_for_n", pbft.id)
		return errors.New("Rplica doesn't have local key for ready_for_n")
	}

	ready := &ReadyForN{
		ReplicaId:	pbft.id,
		Key:		pbft.localKey,
	}

	payload, err := proto.Marshal(ready)
	if err != nil {
		logger.Errorf("Marshal ReadyForN Error!")
		return errors.New("Marshal ReadyForN Error!")
	}
	msg := &ConsensusMessage{
		Type: ConsensusMessage_READY_FOR_N,
		Payload: payload,
	}

	primary := pbft.primary(pbft.view)
	unicast := consensusMsgHelper(msg, pbft.id)
	pbft.helper.InnerUnicast(unicast, primary)

	return nil
}

// Primary receive ready_for_n from new replica
func (pbft *pbftProtocal) recvReadyforNforAdd(ready *ReadyForN) error {

	if pbft.primary(pbft.view) == pbft.id {
		logger.Debugf("Primary %d received ready_for_n from %d", pbft.id, ready.ReplicaId)
	} else {
		logger.Errorf("Replica %d is not primary but received ready_for_n from %d", pbft.id, ready.ReplicaId)
		return errors.New("Replica is not primary but received ready_for_n")
	}

	if !pbft.activeView {
		logger.Warningf("Primary %d is in view change, reject the ready_for_n message", pbft.id)
		return errors.New("Primary is in view change, reject the ready_for_n message")
	}

	cert := pbft.getAddNodeCert(ready.Key)

	if !cert.finishAdd {
		logger.Errorf("Primary %d has not done with addnode for key=%s", pbft.id, ready.Key)
		return errors.New("Primary has not done with addnode")
	}

	// calculate the new N and view
	n, view := pbft.getAddNV()

	// broadcast the updateN message
	updateN := &UpdateN{
		ReplicaId:	pbft.id,
		Key:		ready.Key,
		N:			n,
		View: 		view,
		SeqNo:		pbft.seqNo + 1,
		Flag:		true,
	}

	payload, err := proto.Marshal(updateN)
	if err != nil {
		logger.Errorf("Marshal updateN Error!")
		return errors.New("Marshal updateN Error!")
	}
	msg := &ConsensusMessage{
		Type: ConsensusMessage_UPDATE_N,
		Payload: payload,
	}

	cert.update = updateN
	broadcast := consensusMsgHelper(msg, pbft.id)
	pbft.helper.InnerBroadcast(broadcast)
	return pbft.maybeUpdateN(ready.Key, "", true)
}

// Primary send update_n after finish del node
func (pbft *pbftProtocal) sendUpdateNforDel(key string, routerHash string) error {

	logger.Debugf("Replica %d try to send update_n after finish del node", pbft.id)

	if !pbft.activeView {
		logger.Warningf("Primary %d is in view change, reject the ready_for_n message", pbft.id)
		return errors.New("Primary is in view change, choose not send the ready_for_n message")
	}

	cert := pbft.getDelNodeCert(key, routerHash)

	if !cert.finishDel {
		logger.Errorf("Primary %d has not done with delnode for key=%s", pbft.id, key)
		return errors.New("Primary hasn't done with delnode")
	}

	// calculate the new N and view
	n, view := pbft.getDelNV()

	// broadcast the updateN message
	updateN := &UpdateN{
		ReplicaId:	pbft.id,
		Key:		key,
		RouterHash:	routerHash,
		N:			n,
		View: 		view,
		Flag:		false,
	}

	payload, err := proto.Marshal(updateN)
	if err != nil {
		logger.Errorf("Marshal updateN Error!")
		return errors.New("Marshal updateN Error!")
	}
	msg := &ConsensusMessage{
		Type: ConsensusMessage_UPDATE_N,
		Payload: payload,
	}

	cert.update = updateN
	broadcast := consensusMsgHelper(msg, pbft.id)
	pbft.helper.InnerBroadcast(broadcast)

	return pbft.maybeUpdateN(key, routerHash, false)
}

func (pbft *pbftProtocal) recvUpdateN(update *UpdateN) error {

	logger.Debugf("Replica %d received updateN message from %d", pbft.id, update.ReplicaId)

	if !pbft.activeView {
		logger.Warningf("Replica %d is in view change, reject the update_n message", pbft.id)
		return errors.New("Replica reject update_n msg as it's in viewchange")
	}

	if pbft.primary(pbft.view) != update.ReplicaId {
		logger.Errorf("Replica %d received update_n from other than primary: got %d, should be %d",
		pbft.id, update.ReplicaId, pbft.primary(pbft.view))
		return errors.New("Replica reject update_n msg as it's not from primary")
	}

	if update.Flag {
		n, view := pbft.getAddNV()
		if n != update.N || view != update.View {
			logger.Errorf("Replica %d has different idea: got n=%d/view=%d, should be n=%d/view=%d",
				pbft.id, update.N, update.View, n, view)
			return errors.New("Replica has different idea about n and view")
		}

		if !pbft.inW(update.SeqNo) {
			logger.Warningf("Replica %d received updateN but the seqNo not in view", pbft.id)
			return errors.New("Replica reject not-in-view msg")
		}

		cert := pbft.getAddNodeCert(update.Key)
		cert.update = update

		agree := &AgreeUpdateN{
			ReplicaId:	pbft.id,
			Key:		update.Key,
			N:			n,
			View:		view,
			Flag:		true,
		}

		payload, err := proto.Marshal(agree)
		if err != nil {
			logger.Errorf("Marshal AgreeUpdateN Error!")
			return errors.New("Marshal AgreeUpdateN Error!")
		}
		msg := &ConsensusMessage{
			Type: ConsensusMessage_AGREE_UPDATE_N,
			Payload: payload,
		}

		broadcast := consensusMsgHelper(msg, pbft.id)
		pbft.helper.InnerBroadcast(broadcast)
		return pbft.recvAgreeUpdateN(agree)
	} else {
		n, view := pbft.getDelNV()
		if n != update.N || view != update.View {
			logger.Errorf("Replica %d has different idea: got n=%d/view=%d, should be n=%d/view=%d",
				pbft.id, update.N, update.View, n, view)
			return errors.New("Replica has different idea about n and view")
		}

		cert := pbft.getDelNodeCert(update.Key, update.RouterHash)
		cert.update = update

		agree := &AgreeUpdateN{
			ReplicaId:	pbft.id,
			Key:		update.Key,
			RouterHash:	update.RouterHash,
			N:			n,
			View:		view,
			Flag:		false,
		}

		payload, err := proto.Marshal(agree)
		if err != nil {
			logger.Errorf("Marshal AgreeUpdateN Error!")
			return nil
		}
		msg := &ConsensusMessage{
			Type: ConsensusMessage_AGREE_UPDATE_N,
			Payload: payload,
		}

		broadcast := consensusMsgHelper(msg, pbft.id)
		pbft.helper.InnerBroadcast(broadcast)
		return pbft.recvAgreeUpdateN(agree)
	}
}

func (pbft *pbftProtocal) recvAgreeUpdateN(agree *AgreeUpdateN) error {

	logger.Debugf("Replica %d received agree updateN from replica %d for n=%d/view=%d",
		pbft.id, agree.ReplicaId, agree.N, agree.View)

	if pbft.primary(pbft.view) == agree.ReplicaId {
		logger.Warningf("Replica %d received agree updateN from primary, ignoring", pbft.id)
		return errors.New("")
	}

	if agree.Flag {
		cert := pbft.getAddNodeCert(agree.Key)

		ok := cert.agrees[*agree]
		if ok {
			logger.Warningf("Replica %d ignored duplicate agree updateN from %d", pbft.id, agree.ReplicaId)
			return errors.New("Replica ignored duplicate agree updateN msg")
		}

		cert.agrees[*agree] = true
		cert.updateCount++

		return pbft.maybeUpdateN(agree.Key, "", true)
	} else {
		cert := pbft.getDelNodeCert(agree.Key, agree.RouterHash)

		ok := cert.agrees[*agree]
		if ok {
			logger.Warningf("Replica %d ignored duplicate agree updateN from %d", pbft.id, agree.ReplicaId)
			return errors.New("Replica ignored duplicate agree updateN msg")
		}

		cert.agrees[*agree] = true
		cert.updateCount++

		return pbft.maybeUpdateN(agree.Key, agree.RouterHash, false)
	}
}

func (pbft *pbftProtocal) maybeUpdateN(digest string, routerHash string, flag bool) error {

	if flag {
		cert := pbft.getAddNodeCert(digest)

		if cert.updateCount < pbft.committedReplicasQuorum() {
			return errors.New("Not enough agree message to update n")
		}

		if cert.update == nil {
			logger.Warningf("Replica %d haven't locally prepared for update_n, but got 2f prepared", pbft.id)
			return errors.New("Replica hasn't locally prepared for updating n after adding")
		}

		if cert.finishUpdate {
			if cert.updateCount <= pbft.N + 1 {
				logger.Debugf("Replica %d already finish update for digest %s", pbft.id, digest)
				return errors.New("Replica has already finished updating n after adding")
			} else {
				logger.Warningf("Replica %d already finish update but still try to update N for digest %s", pbft.id, digest)
				return errors.New("Replica has already finished updating n after adding, but still recevice agree msg from someone else")
			}
		}

		// update N, f, view
		pbft.mux.Lock()
		defer pbft.mux.Unlock()
		cert.finishUpdate = true
		pbft.inUpdatingN = true
		pbft.previousN = pbft.N
		pbft.previousView = pbft.view
		pbft.previousF = pbft.f
		pbft.keypoint = cert.update.SeqNo
		pbft.N = int(cert.update.N)
		pbft.view = cert.update.View
		pbft.f = (pbft.N-1) / 3
		pbft.persistView(pbft.view)
		logger.Noticef("Replica %d update after adding, N=%d/f=%d/view=%d/keypoint=%d",
			pbft.id, pbft.N, pbft.f, pbft.view, pbft.keypoint)

	} else {
		cert := pbft.getDelNodeCert(digest, routerHash)

		if cert.updateCount < pbft.preparedReplicasQuorum() {
			return errors.New("Not enough agree message to update n after deleting")
		}

		if cert.update == nil {
			logger.Warningf("Replica %d haven't locally prepared for update_n, but got 2f prepared", pbft.id)
			return errors.New("Replica hasn't locally prepared for updating n after deleting")
		}

		if cert.finishUpdate {
			if cert.updateCount <= pbft.N {
				logger.Debugf("Replica %d already finish update for digest %s", pbft.id, digest)
				return errors.New("Replica has already finished updating n after deleting")
			} else {
				logger.Warningf("Replica %d already finish update but still try to update N for digest %s", pbft.id, digest)
				return errors.New("Replica has already finished updating n after deleting, but still recevice agree msg from someone else")
			}
		}

		// update N, f, view
		pbft.mux.Lock()
		defer pbft.mux.Unlock()
		cert.finishUpdate = true
		pbft.N = int(cert.update.N)
		pbft.view = cert.update.View
		pbft.f = (pbft.N-1) / 3
		oldId := pbft.id
		pbft.id = cert.newId
		pbft.persistView(pbft.view)
		logger.Noticef("Replica %d update after deleting, N=%d/f=%d/view=%d, new local id is %d",
			oldId, pbft.N, pbft.f, pbft.view, pbft.id)
	}

	return nil
}