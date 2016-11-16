package pbft

import (
	"hyperchain/protos"

	"github.com/golang/protobuf/proto"
)

// New replica receive local NewNode message
func (pbft *pbftProtocal) recvLocalNewNode(msg *protos.NewNodeMessage) error {

	logger.Errorf("New replica %d received local newNode message", pbft.id)

	if pbft.isNewNode {
		logger.Warningf("New replica %d received duplicate local newNode message", pbft.id)
		return nil
	}

	pbft.isNewNode = true
	pbft.inAddingNode = true
	key := byteToString(msg.Payload)
	pbft.localKey = key
	logger.Error("localkey: ", key)

	return nil
}

// Replica receive local message about new node and routing table
func (pbft *pbftProtocal) recvLocalAddNode(msg *protos.AddNodeMessage) error {

	if pbft.isNewNode {
		logger.Warningf("New replica received local addNode message, there may be something wrong")
		return nil
	}

	key := byteToString(msg.Payload)
	logger.Errorf("Replica %d received local addNode message for new node %v", pbft.id, key)

	pbft.inAddingNode = true
	pbft.sendAgreeAddNode(key)

	return nil
}

// Replica receive local message about new node and routing table
func (pbft *pbftProtocal) recvLocalDelNode(msg *protos.DelNodeMessage) error {

	key := byteToString(msg.Payload)
	logger.Debugf("Replica %d received local delnode message for del node %v", pbft.id, key)

	pbft.inDeletingNode = true
	pbft.newid = msg.Id
	pbft.sendAgreeDelNode(key)

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
func (pbft *pbftProtocal) sendAgreeDelNode(key string) {

	logger.Debugf("Replica %d try to send delnode message for quit node", pbft.id)

	del := &DelNode{
		ReplicaId:	pbft.id,
		Key:		key,
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
	pbft.helper.BroadcastAddNode(broadcast)
	pbft.recvAgreeDelNode(del)

}

// Replica received addnode for new node
func (pbft *pbftProtocal) recvAgreeAddNode(add *AddNode) error {

	logger.Errorf("Replica %d received addnode from replica %d for %v",
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

	logger.Debugf("Replica %d received agree addnode from replica %d for %v",
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

	if cert == nil {
		logger.Errorf("Replica %d can't get the addnode cert for key=%s", pbft.id, key)
		return nil
	}

	if cert.addCount < pbft.committedReplicasQuorum() {
		return nil
	}

	if !pbft.inAddingNode {
		if cert.finishAdd {
			logger.Warningf("Replica %d have already finish adding node", pbft.id)
		} else {
			// TODO: or just follow others?
			logger.Warningf("Replica %d haven't locally prepared for update routing table, but others have agreed", pbft.id, key)
			return nil
		}
	}

	cert.finishAdd = true
	payload, err := stringToByte(key)
	if err != nil {
		logger.Errorf("Replica %d parse string to byte error", pbft.id)
		return nil
	}

	pbft.helper.UpdateTable(payload)
	pbft.inAddingNode = false

	return nil
}

// Check if replica prepared for update routing table after del node
func (pbft *pbftProtocal) maybeUpdateTableForDel(key string) error {

	cert := pbft.getDelNodeCert(key)

	if cert == nil {
		logger.Errorf("Replica %d can't get the delnode cert for key=%v", pbft.id, key)
		return nil
	}

	if cert.delCount < pbft.committedReplicasQuorum() {
		return nil
	}

	if !pbft.inDeletingNode {
		// TODO: or just follow others?
		logger.Warningf("Replica %d haven't locally prepared for update routing table", pbft.id, key)
		return nil
	}

	cert.finishDel = true
	payload, err := stringToByte(key)
	if err != nil {
		logger.Errorf("Replica %d parse string to byte error", pbft.id)
		return nil
	}

	logger.Errorf("Replica %d try to update routing table", pbft.id)
	pbft.helper.UpdateTable(payload)
	pbft.inDeletingNode = false
	pbft.sendUpdateN(key)

	return nil
}

// New replica send ready_for_n to primary after recovery
func (pbft *pbftProtocal) sendReadyForN() {

	if !pbft.isNewNode {
		logger.Errorf("Replica %d is not new one, but try to send ready_for_n", pbft.id)
		return
	}

	if pbft.localKey == "" {
		logger.Errorf("Replica %d don't have local key to ready_for_n", pbft.id)
	}

	ready := &ReadyForN{
		ReplicaId:	pbft.id,
		Key:		pbft.localKey,
	}

	payload, err := proto.Marshal(ready)
	if err != nil {
		logger.Errorf("Marshal ReadyForN Error!")
		return
	}
	msg := &ConsensusMessage{
		Type: ConsensusMessage_READY_FOR_N,
		Payload: payload,
	}

	primary := pbft.primary(pbft.view, pbft.N)
	unicast := consensusMsgHelper(msg, pbft.id)
	pbft.helper.InnerUnicast(unicast, primary)
	logger.Errorf("Replica %d send readyforn to primary %d", pbft.id, primary)
}

// Primary receive ready_for_n from new replica
func (pbft *pbftProtocal) recvReadyforN(ready *ReadyForN) error {

	if pbft.primary(pbft.view, pbft.N) == pbft.id {
		logger.Errorf("Primary %d received ready_for_n from %d", pbft.id, ready.ReplicaId)
	} else {
		logger.Errorf("Replica %d received ready_for_n from %d", pbft.id, ready.ReplicaId)
		return nil
	}

	if !pbft.activeView {
		logger.Warningf("Primary %d is in view change, reject the ready_for_n message", pbft.id)
		return nil
	}

	cert := pbft.getAddNodeCert(ready.Key)

	if cert == nil {
		logger.Errorf("Primary %d can't get the addnode cert for key=%s", pbft.id, ready.Key)
		return nil
	}

	if !cert.finishAdd {
		logger.Errorf("Primary %d has not done with addnode for key=%s", pbft.id, ready.Key)
		return nil
	}

	// calculate the new N and view
	n, view := pbft.getAddNV()

	// broadcast the updateN message
	updateN := &UpdateN{
		ReplicaId:	pbft.id,
		Key:		ready.Key,
		N:			n,
		View: 		view,
		SeqNo:		pbft.seqNo,
		Flag:		true,
	}

	payload, err := proto.Marshal(updateN)
	if err != nil {
		logger.Errorf("Marshal updateN Error!")
		return nil
	}
	msg := &ConsensusMessage{
		Type: ConsensusMessage_UPDATE_N,
		Payload: payload,
	}

	cert.update = updateN
	broadcast := consensusMsgHelper(msg, pbft.id)
	pbft.helper.InnerBroadcast(broadcast)

	return pbft.maybeUpdateN(ready.Key, true)
}

// Primary send update_n after finish del node
func (pbft *pbftProtocal) sendUpdateN(key string) {

	logger.Errorf("Replica %d try to send update_n after finish del node", pbft.id)

	if !pbft.activeView {
		logger.Warningf("Primary %d is in view change, reject the ready_for_n message", pbft.id)
		return
	}

	cert := pbft.getDelNodeCert(key)

	if cert == nil {
		logger.Errorf("Primary %d can't get the addnode cert for key=%s", pbft.id, key)
		return
	}

	if !cert.finishDel {
		logger.Errorf("Primary %d has not done with addnode for key=%s", pbft.id, key)
		return
	}

	// calculate the new N and view
	n, view := pbft.getDelNV()

	// broadcast the updateN message
	updateN := &UpdateN{
		ReplicaId:	pbft.id,
		Key:		key,
		N:			n,
		View: 		view,
		Flag:		false,
	}

	payload, err := proto.Marshal(updateN)
	if err != nil {
		logger.Errorf("Marshal updateN Error!")
		return
	}
	msg := &ConsensusMessage{
		Type: ConsensusMessage_UPDATE_N,
		Payload: payload,
	}

	cert.update = updateN
	broadcast := consensusMsgHelper(msg, pbft.id)
	pbft.helper.InnerBroadcast(broadcast)

	pbft.maybeUpdateN(key, false)
}

func (pbft *pbftProtocal) recvUpdateN(update *UpdateN) error {

	logger.Errorf("Replica %d received updateN message from %d", pbft.id, update.ReplicaId)

	if !pbft.activeView {
		logger.Warningf("Replica %d is in view change, reject the update_n message", pbft.id)
		return nil
	}

	if pbft.primary(pbft.view, pbft.N) != update.ReplicaId {
		logger.Errorf("Replica %d received updateN from other than primary: got %d, should be %d",
		pbft.id, update.ReplicaId, pbft.primary(pbft.view, pbft.N))
		return nil
	}

	if update.Flag {
		n, view := pbft.getAddNV()
		if n != update.N || view != update.View {
			logger.Errorf("Replica %d has different idea: got n=%d/view=%d, should be n=%d/view=%d",
				pbft.id, update.N, update.View, n, view)
			return nil
		}

		if !pbft.inW(update.SeqNo) {
			logger.Warningf("Replica %d received updateN but the seqNo not in view", pbft.id)
			return nil
		}

		cert := pbft.getAddNodeCert(update.Key)
		if cert == nil {
			logger.Errorf("Primary %d can't get the addnode cert for key=%s", pbft.id, update.Key)
			return nil
		}

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
			return nil
		}
		msg := &ConsensusMessage{
			Type: ConsensusMessage_AGREE_UPDATE_N,
			Payload: payload,
		}

		broadcast := consensusMsgHelper(msg, pbft.id)
		pbft.helper.InnerBroadcast(broadcast)
		pbft.inUpdatingN = true
		return pbft.recvAgreeUpdateN(agree)
	} else {
		n, view := pbft.getDelNV()
		if n != update.N || view != update.View {
			logger.Errorf("Replica %d has different idea: got n=%d/view=%d, should be n=%d/view=%d",
				pbft.id, update.N, update.View, n, view)
			return nil
		}

		cert := pbft.getDelNodeCert(update.Key)
		if cert == nil {
			logger.Errorf("Primary %d can't get the delnode cert for key=%s", pbft.id, update.Key)
			return nil
		}

		cert.update = update

		agree := &AgreeUpdateN{
			ReplicaId:	pbft.id,
			Key:		update.Key,
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

	if pbft.primary(pbft.view, pbft.N) == agree.ReplicaId {
		logger.Warningf("Replica %d received agree updateN from primary, ignoring", pbft.id)
		return nil
	}

	if agree.Flag {
		cert := pbft.getAddNodeCert(agree.Key)

		ok := cert.agrees[*agree]
		if ok {
			logger.Warningf("Replica %d ignored duplicate agree updateN from %d", pbft.id, agree.ReplicaId)
			return nil
		}

		cert.agrees[*agree] = true
		cert.updateCount++

		return pbft.maybeUpdateN(agree.Key, true)
	} else {
		cert := pbft.getDelNodeCert(agree.Key)

		ok := cert.agrees[*agree]
		if ok {
			logger.Warningf("Replica %d ignored duplicate agree updateN from %d", pbft.id, agree.ReplicaId)
			return nil
		}

		cert.agrees[*agree] = true
		cert.updateCount++

		return pbft.maybeUpdateN(agree.Key, false)
	}
}

func (pbft *pbftProtocal) maybeUpdateN(digest string, flag bool) error {

	if flag {
		cert := pbft.getAddNodeCert(digest)

		if cert == nil {
			logger.Errorf("Replica %d can't get the cert for digest=%s", pbft.id, digest)
			return nil
		}

		if cert.update == nil {
			logger.Warningf("Replica %d has not received updateN yet", pbft.id)
			return nil
		}

		if cert.updateCount < pbft.preparedReplicasQuorum() {
			return nil
		}

		if !pbft.inUpdatingN {
			logger.Warningf("Replica %d haven't locally prepared for update_n, but got 2f prepared", pbft.id)
			return nil
		}

		// update N, f, view
		pbft.mux.Lock()
		defer pbft.mux.Unlock()
		pbft.inUpdatingN = false
		pbft.previousN = pbft.N
		pbft.previousView = pbft.view
		pbft.previousF = pbft.f
		pbft.keypoint = cert.update.SeqNo
		pbft.N = int(cert.update.N)
		pbft.view = cert.update.View
		pbft.f = (pbft.N-1) / 3
		logger.Warningf("Replica %d update N=%d/f=%d/view=%d")

	} else {
		cert := pbft.getAddNodeCert(digest)

		if cert == nil {
			logger.Errorf("Replica %d can't get the cert for digest=%s", pbft.id, digest)
			return nil
		}

		if cert.update == nil {
			logger.Warningf("Replica %d has not received updateN yet", pbft.id)
			return nil
		}

		if cert.updateCount < pbft.preparedReplicasQuorum() {
			return nil
		}

		if !pbft.inUpdatingN {
			logger.Warningf("Replica %d haven't locally prepared for update_n, but got 2f prepared", pbft.id)
			return nil
		}

		// update N, f, view
		pbft.mux.Lock()
		defer pbft.mux.Unlock()
		pbft.N = int(cert.update.N)
		pbft.view = cert.update.View
		pbft.f = (pbft.N-1) / 3
	}

	return nil
}