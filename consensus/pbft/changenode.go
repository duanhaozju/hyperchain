package pbft

import (
	"hyperchain/protos"

	"github.com/golang/protobuf/proto"
)

// New replica receive local NewNode message
func (pbft *pbftProtocal) recvLocalNewNode(msg *protos.NewNodeMessage) error {

	logger.Debugf("New replica %d received local newNode message", pbft.id)

	if pbft.isNewNode {
		logger.Warningf("New replica %d received duplicate local newNode message", pbft.id)
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

	key := string(msg.Payload)
	logger.Debugf("Replica %d received local addNode message for new node %v", pbft.id, key)

	pbft.inAddingNode = true
	pbft.sendAgreeAddNode(key)

	return nil
}

// Replica receive local message about new node and routing table
func (pbft *pbftProtocal) recvLocalDelNode(msg *protos.DelNodeMessage) error {

	key := string(msg.DelPayload)
	logger.Errorf("Replica %d received local delnode message for del node %s", pbft.id, key)
	if pbft.N == 4 {
		logger.Criticalf("Replica %d receive del msg, but we don't support delete as there're only 4 nodes")
		return nil
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

	logger.Errorf("Replica %d try to send delnode message for quit node", pbft.id)

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
		return nil
	}

	cert.addNodes[*add] = true
	cert.addCount++

	return pbft.maybeUpdateTableForAdd(add.Key)
}

// Replica received delnode for quit node
func (pbft *pbftProtocal) recvAgreeDelNode(del *DelNode) error {

	logger.Errorf("Replica %d received agree delnode from replica %d for %v",
		pbft.id, del.ReplicaId, del.Key)

	cert := pbft.getDelNodeCert(del.Key, del.RouterHash)

	ok := cert.delNodes[*del]
	if ok {
		logger.Warningf("Replica %d ignored duplicate agree addnode from %d", pbft.id, del.ReplicaId)
		return nil
	}

	cert.delNodes[*del] = true
	cert.delCount++

	return pbft.maybeUpdateTableForDel(del.Key, del.RouterHash)
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
			if cert.addCount <= pbft.N {
				logger.Debugf("Replica %d have already finished adding node", pbft.id)
				return nil
			}
			logger.Warningf("Replica %d have already finished adding node, but still recevice add msg", pbft.id)
			return nil
		}
	//else {
	//		// TODO: or just follow others?
	//		logger.Warningf("Replica %d haven't locally prepared for update routing table for %s, but others have agreed", pbft.id, key)
	//		return nil
	//	}
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

	if cert == nil {
		logger.Errorf("Replica %d can't get the delnode cert for key=%v", pbft.id, key)
		return nil
	}

	if cert.delCount < pbft.committedReplicasQuorum() {
		return nil
	}

	if !pbft.inDeletingNode {
		if cert.finishDel {
			if cert.delCount < pbft.N {
				logger.Debugf("Replica %d have already finished deleting node", pbft.id)
				return nil
			}
			logger.Warningf("Replica %d have already finished deleting node, but still recevice del msg", pbft.id)
			return nil
		}
		//} else {
		//	// TODO: or just follow others?
		//	logger.Warningf("Replica %d haven't locally prepared for update routing table for %s, but others have agreed", pbft.id, key)
		//	return nil
		//}
	}

	cert.finishDel = true
	payload := []byte(key)

	logger.Errorf("Replica %d try to update routing table", pbft.id)
	pbft.helper.UpdateTable(payload, false)
	pbft.inDeletingNode = false
	if pbft.primary(pbft.view) == pbft.id {
		pbft.sendUpdateNforDel(key, routerHash)
	}

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

	primary := pbft.primary(pbft.view)
	unicast := consensusMsgHelper(msg, pbft.id)
	pbft.helper.InnerUnicast(unicast, primary)
	logger.Debugf("Replica %d send readyforn to primary %d", pbft.id, primary)
}

// Primary receive ready_for_n from new replica
func (pbft *pbftProtocal) recvReadyforNforAdd(ready *ReadyForN) error {

	if pbft.primary(pbft.view) == pbft.id {
		logger.Debugf("Primary %d received ready_for_n from %d", pbft.id, ready.ReplicaId)
	} else {
		logger.Errorf("Replica %d is not primary but received ready_for_n from %d", pbft.id, ready.ReplicaId)
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
		SeqNo:		pbft.seqNo + 1,
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

	return pbft.maybeUpdateN(ready.Key, "", true)
}

// Primary send update_n after finish del node
func (pbft *pbftProtocal) sendUpdateNforDel(key string, routerHash string) {

	logger.Errorf("Replica %d try to send update_n after finish del node", pbft.id)

	if !pbft.activeView {
		logger.Warningf("Primary %d is in view change, reject the ready_for_n message", pbft.id)
		return
	}

	cert := pbft.getDelNodeCert(key, routerHash)

	if cert == nil {
		logger.Errorf("Primary %d can't get the delnode cert for key=%s", pbft.id, key)
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
		RouterHash:	routerHash,
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

	pbft.maybeUpdateN(key, routerHash, false)
}

func (pbft *pbftProtocal) recvUpdateN(update *UpdateN) error {

	logger.Criticalf("Replica %d received updateN message from %d", pbft.id, update.ReplicaId)

	if !pbft.activeView {
		logger.Warningf("Replica %d is in view change, reject the update_n message", pbft.id)
		return nil
	}

	if pbft.primary(pbft.view) != update.ReplicaId {
		logger.Errorf("Replica %d received updateN from other than primary: got %d, should be %d",
		pbft.id, update.ReplicaId, pbft.primary(pbft.view))
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
		return pbft.recvAgreeUpdateN(agree)
	} else {
		n, view := pbft.getDelNV()
		if n != update.N || view != update.View {
			logger.Errorf("Replica %d has different idea: got n=%d/view=%d, should be n=%d/view=%d",
				pbft.id, update.N, update.View, n, view)
			return nil
		}

		cert := pbft.getDelNodeCert(update.Key, update.RouterHash)
		if cert == nil {
			logger.Errorf("Primary %d can't get the delnode cert for key=%s", pbft.id, update.Key)
			return nil
		}

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

	logger.Criticalf("Replica %d received agree updateN from replica %d for n=%d/view=%d",
		pbft.id, agree.ReplicaId, agree.N, agree.View)

	if pbft.primary(pbft.view) == agree.ReplicaId {
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

		return pbft.maybeUpdateN(agree.Key, "", true)
	} else {
		cert := pbft.getDelNodeCert(agree.Key, agree.RouterHash)

		ok := cert.agrees[*agree]
		if ok {
			logger.Warningf("Replica %d ignored duplicate agree updateN from %d", pbft.id, agree.ReplicaId)
			return nil
		}

		cert.agrees[*agree] = true
		cert.updateCount++

		return pbft.maybeUpdateN(agree.Key, agree.RouterHash, false)
	}
}

func (pbft *pbftProtocal) maybeUpdateN(digest string, routerHash string, flag bool) error {

	if flag {
		cert := pbft.getAddNodeCert(digest)

		if cert == nil {
			logger.Errorf("Replica %d can't get the cert for digest=%s", pbft.id, digest)
			return nil
		}

		if cert.updateCount < pbft.committedReplicasQuorum() {
			return nil
		}

		if cert.update == nil {
			logger.Warningf("Replica %d haven't locally prepared for update_n, but got 2f prepared", pbft.id)
			return nil
		}

		if cert.finishUpdate {
			if cert.updateCount < pbft.N + 1 {
				logger.Debugf("Replica %d already finish update for digest %s", pbft.id, digest)
				return nil
			} else {
				logger.Warningf("Replica %d already finish update but still try to update N for digest %s", pbft.id, digest)
				return nil
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
		logger.Noticef("Replica %d update after adding, N=%d/f=%d/view=%d/keypoint=%d",
			pbft.id, pbft.N, pbft.f, pbft.view, pbft.keypoint)

	} else {
		cert := pbft.getDelNodeCert(digest, routerHash)

		if cert == nil {
			logger.Errorf("Replica %d can't get the cert for digest=%s", pbft.id, digest)
			return nil
		}

		if cert.updateCount < pbft.preparedReplicasQuorum() {
			return nil
		}

		if cert.update == nil {
			logger.Warningf("Replica %d haven't locally prepared for update_n, but got 2f prepared", pbft.id)
			return nil
		}

		if cert.finishUpdate {
			if cert.updateCount < pbft.N {
				logger.Debugf("Replica %d already finish update for digest %s", pbft.id, digest)
				return nil
			} else {
				logger.Warningf("Replica %d already finish update but still try to update N for digest %s", pbft.id, digest)
				return nil
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
		logger.Noticef("Replica %d update after deleting, N=%d/f=%d/view=%d, new local id is %d",
			oldId, pbft.N, pbft.f, pbft.view, pbft.id)
	}

	return nil
}