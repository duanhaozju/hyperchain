package pbft

import (
	"hyperchain/protos"

	"github.com/golang/protobuf/proto"
)

// New replica receive local NewNode message
func (pbft *pbftProtocal) recvLocalNewNode(msg protos.NewNodeMessage) error {

	logger.Debugf("New replica received local newNode message for ip=%s/N=%d", msg.Ip, msg.N)

	if pbft.isNewNode {
		logger.Warningf("New replica received duplicate local newNode message for ip=%s/N=%d", msg.Ip, msg.N)
		return nil
	}

	pbft.isNewNode = true
	pbft.inAddingNode = true
	pbft.newIP = msg.Ip
	pbft.N = msg.N
	pbft.f = (msg.N-1)/3

	logger.Debug("New replica initially update N=%d/f=$d", pbft.N, pbft.f)

	return nil
}

// Replica receive local message about new node and routing table
func (pbft *pbftProtocal) recvLocalAddNode(msg protos.AddNodeMessage) error {

	if pbft.isNewNode {
		logger.Warningf("New replica received local addNode message, there may be something wrong")
		return nil
	}

	logger.Debugf("Replica %d received local addNode message for new node ip=%s/id=%d", pbft.id, msg.Ip, msg.NewId)

	if pbft.tableReceived {
		logger.Warningf("Replica %d already received local message, reject this message", pbft.id)
		return nil
	}

	// store the info about new replica
	pbft.newIP = msg.Ip
	pbft.newID = msg.NewId
	pbft.routingTable = msg.RoutingTable
	pbft.tableDigest = hashString(pbft.routingTable)

	pbft.inTableError = false
	pbft.tableReceived = true

	if pbft.primary(pbft.view, pbft.N) == pbft.id && pbft.activeView{
		pbft.inAddingNode = true
		pbft.sendAddNode()
		return nil
	}

	cert := pbft.getAddNodeCert(pbft.tableDigest)
	cert.table = pbft.routingTable
	if pbft.inAddingNode {
		if cert.addNode.TableDigest == pbft.tableDigest {
			pbft.sendAgreeAddNode()
		} else {
			pbft.inTableError = true
		}
	}

	return nil
}

// Primary broadcast addnode to all replicas(include new one), and unicast routing table to new node
func (pbft *pbftProtocal) sendAddNode() {

	logger.Debugf("Replica %d is primary, send the addnode message to other replicas", pbft.id)

	// primary send the routing table to the new replica
	routingTable := &RoutingTable{
		ReplicaId:	pbft.id,
		Table:		pbft.routingTable,
		NewId:		pbft.newID,
	}
	unicastPayload, err := proto.Marshal(routingTable)
	if err != nil {
		logger.Errorf("Marshal AddNode Error!")
		return
	}
	unicastMsg := &ConsensusMessage{
		Type: ConsensusMessage_ROUTING_TABLE,
		Payload: unicastPayload,
	}
	unicast := consensusMsgHelper(unicastMsg, pbft.id)
	pbft.helper.InnerUnicast(unicast, pbft.newID)

	// broadcast the new digest of the routing table
	addNodeMsg := &AddNode{
		ReplicaId:	pbft.id,
		Ip:			pbft.newIP,
		TableDigest:pbft.tableDigest,
		NewId:		pbft.newID,
	}

	cert := pbft.getAddNodeCert(pbft.tableDigest)
	cert.table = pbft.routingTable
	cert.addNode = addNodeMsg

	broadcastPayload, err := proto.Marshal(addNodeMsg)
	if err != nil {
		logger.Errorf("Marshal AddNode Error!")
		return
	}
	broadcastMsg := &ConsensusMessage{
		Type: ConsensusMessage_ADD_NODE,
		Payload: broadcastPayload,
	}
	broadcast := consensusMsgHelper(broadcastMsg, pbft.id)
	pbft.helper.InnerBroadcast(broadcast)

	pbft.maybeUpdateTable(pbft.tableDigest)
}


// New replica receive routing table from primary
// Or old replica need recovery
func (pbft *pbftProtocal) recvRoutingTable(table *RoutingTable) error {

	if pbft.isNewNode {
		logger.Debugf("New replica %d received routing table from primary %d, table=%s",
			table.NewId, table.ReplicaId, table.Table)
		pbft.id = table.NewId
		pbft.routingTable = table.Table
		pbft.tableDigest = hashString(pbft.routingTable)
		pbft.tableReceived = true
		cert := pbft.getAddNodeCert(pbft.tableDigest)
		cert.table = pbft.routingTable
	} else if pbft.inTableError {
		logger.Debugf("Replica %d received new routing table from primary %d", table.NewId, table.ReplicaId)
		pbft.id = table.NewId
		// TODO: self find table error
	}
	return nil
}

// Replica receive addnode message from primary
func (pbft *pbftProtocal) recvAddNode(addnode *AddNode) error {

	logger.Debugf("Replica %d received addnode from replica %d for newIP=%s/newID=%d",
		pbft.id, addnode.ReplicaId, addnode.Ip, addnode.NewId)

	cert := pbft.getAddNodeCert(addnode.TableDigest)

	if pbft.tableReceived {
		if pbft.tableDigest == addnode.TableDigest {
			cert.addNode = addnode
			pbft.sendAgreeAddNode()
		} else {
			logger.Warningf("Replica %d find tableErr itself", pbft.id)
			pbft.inTableError = true
		}
	} else {
		logger.Debugf("Replica %d received addnode message from primary, but hasn't get local routing table yet", pbft.id)
	}

	return nil
}

// Repica broadcast agree message for addnode
func (pbft *pbftProtocal) sendAgreeAddNode() {

	logger.Debugf("Replica %d try to send agree message for addnode", pbft.id)

	if pbft.isNewNode {
		logger.Debugf("New replica does not need to send agree message")
		return
	}

	agree := &AgreeAddNode{
		ReplicaId:	pbft.id,
		Ip:			pbft.newIP,
		NewId:		pbft.newID,
		TableDigest:pbft.tableDigest,
	}

	payload, err := proto.Marshal(agree)
	if err != nil {
		logger.Errorf("Marshal AgreeAddNode Error!")
		return
	}
	msg := &ConsensusMessage{
		Type: ConsensusMessage_AGREE_ADD_NODE,
		Payload: payload,
	}

	broadcast := consensusMsgHelper(msg, pbft.id)
	pbft.helper.InnerBroadcast(broadcast)
	pbft.recvAgreeAddNode(agree)

	return
}

// Replica received agree for addnode
func (pbft *pbftProtocal) recvAgreeAddNode(agree *AgreeAddNode) error {

	logger.Debugf("Replica %d received agree addnode from replica %d for newIP=%s/newID=%d",
		pbft.id, agree.ReplicaId, agree.Ip, agree.NewId)

	if pbft.primary(pbft.view, pbft.N) == agree.ReplicaId {
		logger.Warningf("Replica %d received agree addnode from primary, ignoring", pbft.id)
	}

	// TODO: check if in recovery

	cert := pbft.getAddNodeCert(agree.TableDigest)

	ok := cert.agreeAdd[*agree]
	if ok {
		logger.Warningf("Replica %d ignored duplicate agree addnode from %d", pbft.id, agree.ReplicaId)
		return nil
	}

	cert.agreeAdd[*agree] = true
	cert.addCount++

	return pbft.maybeUpdateTable(agree.TableDigest)
}

// Check if replica prepared for update routing table
func (pbft *pbftProtocal) maybeUpdateTable(digest string) error {

	cert := pbft.getAddNodeCert(digest)

	if cert == nil {
		logger.Errorf("Replica %d can't get the cert for digest=%s", pbft.id, digest)
		return nil
	}

	if cert.addCount < pbft.preparedReplicasQuorum() {
		return nil
	}

	if !pbft.tableReceived {
		logger.Errorf("Replica %d hasn't received local addnode message, but already prepared for update routing table", pbft.id)
		return nil
	}

	if pbft.inTableError {
		logger.Debugf("Replica %d update routing table, but local one is wrong", pbft.id)
		// TODO : old replica need recovery
		//pbft.helper.UpdateTable(cert.table, false)
		return nil
	}

	if pbft.isNewNode {
		logger.Debugf("New replica %d finish get routing table", pbft.id)
		pbft.helper.UpdateTable(cert.table, false)
		// TODO: new replica start recovery
	} else {
		pbft.helper.UpdateTable(cert.table, true)
	}

	return nil
}

// New replica send ready_for_n to primary after recovery
func (pbft *pbftProtocal) sendReadyforN() {

	if !pbft.isNewNode {
		logger.Errorf("Replica %d is not new one, but try to send ready_for_n", pbft.id)
		return
	}

	ready := &ReadyForN{
		ReplicaId:	pbft.id,
		TableDigest:pbft.tableDigest,
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
}

// Primary receive ready_for_n from new replica
func (pbft *pbftProtocal) recvReadyforN(ready *ReadyForN) error {

	if pbft.primary(pbft.view, pbft.N) == pbft.id {
		logger.Debugf("Primary %d received ready_for_n from %d", pbft.id, ready.ReplicaId)
	} else {
		logger.Errorf("Replica %d received ready_for_n from %d", pbft.id, ready.ReplicaId)
		return nil
	}

	if !pbft.activeView {
		logger.Warningf("Primary %d is in view change, reject the ready_for_n message", pbft.id)
		return nil
	}

	if !pbft.inAddingNode {
		logger.Errorf("Primary %d is not in adding node, but received ready_for_n", pbft.id)
		return nil
	}

	if ready.ReplicaId != pbft.newID || ready.TableDigest != pbft.tableDigest {
		logger.Errorf("Primary %d found wrong info in ready_for_n, reject it", pbft.id)
		return nil
	}

	// calculate the new N and view
	n, view := pbft.getUpdatedNf()

	// broadcast the updateN message
	updateN := &UpdateN{
		ReplicaId:		pbft.id,
		TableDigest:	pbft.tableDigest,
		N:				n,
		SeqNo:			pbft.seqNo,
		View: 			view,
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

	cert := pbft.getAddNodeCert(ready.TableDigest)
	cert.update = updateN
	broadcast := consensusMsgHelper(msg, pbft.id)
	pbft.helper.InnerBroadcast(broadcast)

	return pbft.maybeStartUpdateN(pbft.tableDigest)
}


func (pbft *pbftProtocal) recvUpdateN(update *UpdateN) error {

	logger.Debugf("Replica %d received updateN message from %d", pbft.id, update.ReplicaId)

	if !pbft.activeView {
		logger.Warningf("Replica %d is in view change, reject the ready_for_n message", pbft.id)
		return nil
	}

	if !pbft.inAddingNode {
		logger.Errorf("Replica %d is not in adding node, but received ready_for_n", pbft.id)
		return nil
	}

	if pbft.primary(pbft.view, pbft.N) != update.ReplicaId {
		logger.Errorf("Replica %d received updateN from other than primary: got %d, should be %d",
		pbft.id, update.ReplicaId, pbft.primary(pbft.view, pbft.N))
		return nil
	}

	if !pbft.inW(update.SeqNo) {
		logger.Errorf("Replica %d thinks the seqNo=%d from primary is out of watermark", pbft.id, update.SeqNo)
		return nil
	}

	if update.TableDigest != update.TableDigest {
		logger.Errorf("Replica %d has different table", pbft.id)
		return nil
	}

	n, view := pbft.getUpdatedNf()
	if n != update.N || view != update.View {
		logger.Errorf("Replica %d has different idea: got n=%d/view=%d, should be n=%d/view=%d",
		pbft.id, update.N, update.View, n, view)
		return nil
	}

	agree := &AgreeUpdateN{
		ReplicaId:	pbft.id,
		TableDigest:	pbft.tableDigest,
		N:				n,
		View:			view,
		SeqNo:			update.SeqNo,
	}

	payload, err := proto.Marshal(agree)
	if err != nil {
		logger.Errorf("Marshal updateN Error!")
		return
	}
	msg := &ConsensusMessage{
		Type: ConsensusMessage_AGREE_UPDATE_N,
		Payload: payload,
	}

	broadcast := consensusMsgHelper(msg, pbft.id)
	pbft.helper.InnerBroadcast(broadcast)

	return pbft.recvAgreeUpdateN(agree)

}

func (pbft *pbftProtocal) recvAgreeUpdateN(agree *UpdateN) error {

	logger.Debugf("Replica %d received agree updateN from replica %d for n=%d/view=%d/seqNo=%d",
		pbft.id, agree.ReplicaId, agree.N, agree.View, agree.SeqNo)

	if pbft.primary(pbft.view, pbft.N) == agree.ReplicaId {
		logger.Warningf("Replica %d received agree updateN from primary, ignoring", pbft.id)
	}

	// TODO: check if in recovery

	cert := pbft.getAddNodeCert(agree.TableDigest)

	ok := cert.agreeUpdate[*agree]
	if ok {
		logger.Warningf("Replica %d ignored duplicate agree updateN from %d", pbft.id, agree.ReplicaId)
		return nil
	}

	cert.agreeUpdate[*agree] = true
	cert.updateCount++

	return pbft.maybeStartUpdateN(agree.TableDigest)
}

func (pbft *pbftProtocal) maybeStartUpdateN(digest string) error {


	return nil
}