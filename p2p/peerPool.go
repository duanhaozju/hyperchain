//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package p2p

import (
	pb "hyperchain/p2p/peermessage"
)

//PeerPool provide the basic peer manager function
type PeersPool interface{
	PutPeer(address pb.PeerAddress, client *Peer) error
	PutPeerToTemp(address pb.PeerAddress, client *Peer) error
	GetPeer(address pb.PeerAddress) *Peer
	GetPeers() []*Peer
	GetPeersWithTemp() []*Peer
	GetAliveNodeNum() int
	ToRoutingTable() pb.Routers
	ToRoutingTableWithout(hash string)pb.Routers
	MergeFormRoutersToTemp(routers pb.Routers)
	MergeTempPeers(peer *Peer)
	MergeTempPeersForNewNode()
	RejectTempPeers()
	DeletePeer(peer *Peer)
}