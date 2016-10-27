package pbft

import (
	"hyperchain/protos"

	"github.com/golang/protobuf/proto"
)

func (pbft *pbftProtocal) recvLocalNewNode(msg protos.NewNodeMessage) error {

	logger.Debugf("New replica received local newNode message for ip=%s/N=%d", msg.Ip, msg.N)

	if pbft.isNewNode {
		logger.Warningf("New replica received duplicate local newNode message for ip=%s/N=%d", msg.Ip, msg.N)
		return nil
	}

	pbft.isNewNode = true
	pbft.inAddingNode = true
	pbft.inGettingTable = true
	pbft.newIP = msg.Ip
	pbft.N = msg.N
	pbft.f = (msg.N-1)/3

	logger.Debug("New replica initially update N=%d/f=$d", pbft.N, pbft.f)

	return nil
}

func (pbft *pbftProtocal) recvLocalAddNode(msg protos.AddNodeMessage) error {

	if pbft.isNewNode {
		logger.Warningf("New replica received local addNode message, there may be something wrong")
		return nil
	}

	logger.Debugf("Replica %d received local addNode message for new node ip=%s/id=%d", pbft.id, msg.Ip, msg.NewId)

	if pbft.inAddingNode {
		logger.Warningf("Replica %d already in addingNode, reject this message", pbft.id)
		return nil
	}

	// store the info about new replica
	pbft.newIP = msg.Ip
	pbft.newID = msg.NewId
	pbft.routingTable = msg.RoutingTable
	pbft.tableDigest = hashString(pbft.routingTable)

	pbft.inAddingNode = true
	pbft.inTableError = false

	if pbft.primary(pbft.view, pbft.N) == pbft.id {
		pbft.sendAddNode()
	} else {
		cert := pbft.getAddNodeCert(pbft.newIP, pbft.tableDigest)
		cert.table = pbft.routingTable
	}

	return nil
}

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
		return nil
	}
	unicastMsg := &ConsensusMessage{
		Type: ConsensusMessage_ROUTING_TABLE,
		Payload: unicastPayload,
	}
	unicast := consensusMsgHelper(unicastMsg, pbft.id)
	pbft.helper.InnerUnicast(unicast)

	// broadcast the new routing table
	addNodeMsg := &AddNode{
		ReplicaId:	pbft.id,
		Ip:			pbft.newIP,
		TableDigest:pbft.tableDigest,
		NewId:		pbft.newID,
	}

	cert := pbft.getAddNodeCert(pbft.newIP, pbft.tableDigest)
	cert.table = pbft.routingTable
	cert.addNode = addNodeMsg

	broadcastPayload, err := proto.Marshal(addNodeMsg)
	if err != nil {
		logger.Errorf("Marshal AddNode Error!")
		return nil
	}
	broadcastMsg := &ConsensusMessage{
		Type: ConsensusMessage_ADD_NODE,
		Payload: broadcastPayload,
	}
	broadcast := consensusMsgHelper(broadcastMsg, pbft.id)
	pbft.helper.InnerBroadcast(broadcast)

	pbft.maybeUpdateTable(pbft.newIP, pbft.tableDigest)
}

func (pbft *pbftProtocal) recvRoutingTable(table *RoutingTable) error {

	if pbft.isNewNode {
		logger.Debugf("New replica %d received routing table from primary %d, table=%s",
			table.NewId, table.ReplicaId, table.Table)
		pbft.id = table.NewId
		pbft.routingTable = table.Table
		pbft.tableDigest = hashString(pbft.routingTable)
	} else if pbft.inTableError {
		logger.Debugf("Replica %d received new routing table from primary %d", table.NewId, table.ReplicaId)
		pbft.id = table.NewId
		// TODO: self find table error
	}
	return nil
}

func (pbft *pbftProtocal) recvAddNode(addnode *AddNode) error {

	logger.Debugf("Replica %d received addnode from replica %d for newIP=%s/newID=%d",
		pbft.id, addnode.ReplicaId, addnode.Ip, addnode.NewId)

	if pbft.isNewNode && pbft.inAddingNode {
		// TODO
	}

	if !pbft.isNewNode && pbft.inAddingNode && !pbft.inTableError && pbft.tableDigest != addnode.TableDigest{
		logger.Debugf("Replica %d find local routing table different from primary's", pbft.id)
		pbft.inTableError = true
		// TODO recovery
		return nil
	}

	if pbft.inAddingNode {

	}

	return nil
}

func (pbft *pbftProtocal) recvAgreeAddNode(agree *AgreeAddNode) error {

	logger.Debugf("Replica %d received addnode from replica %d for newIP=%s/newID=%d",
		pbft.id, agree.ReplicaId, agree.Ip, agree.NewId)

	if pbft.isNewNode && pbft.inAddingNode {

	}

	if !pbft.isNewNode && pbft.inAddingNode && !pbft.inTableError && pbft.tableDigest != agree.TableDigest{
		logger.Debugf("Replica %d find local routing table different from primary's", pbft.id)
		pbft.inTableError = true
		// TODO recovery
		return nil
	}

	if pbft.inAddingNode {

	}

	cert := pbft.getAddNodeCert(agree.Ip, agree.TableDigest)
	if cert == nil {

	}


	pbft.maybeUpdateTable(agree.Ip, agree.TableDigest)

	return nil
}

func (pbft *pbftProtocal) maybeUpdateTable(ip string, digest string) {

	cert := pbft.getAddNodeCert(ip, digest)

	if cert == nil {
		logger.Errorf("Replica %d can't get the cert for ip=%s/digest=%s", pbft.id, ip, digest)
		return
	}

	if cert.count < pbft.preparedReplicasQuorum() {
		return
	}

	if !pbft.inAddingNode {
		logger.Debugf("Replica %d hasn't received local addnode message, but already")
	}

	if pbft.inTableError {
		logger.Debugf("Replica %d update routing table, but local one is wrong", pbft.id)
		// TODO
		//pbft.helper.UpdateTable(cert.table, false)
	} else {
		logger.Debugf("Replica %d update routing table, and local one is right", pbft.id)
		pbft.helper.UpdateTable(pbft.routingTable, true)
	}
}