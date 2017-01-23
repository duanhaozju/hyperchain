//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package p2p

import (
	"google.golang.org/grpc"
	pb "hyperchain/p2p/peermessage"
)

//PeerPool provide the basic peer manager function
type PeersPool interface {
	PutPeer(address pb.PeerAddr, client *Peer) error
	PutPeerToTemp(address pb.PeerAddr, client *Peer) error
	GetPeer(address pb.PeerAddr) *Peer
	GetPeerByHash(hash string) (*Peer, error)
	GetPeers() []*Peer
	GetVPPeers() []*Peer
	GetNVPPeers() []*Peer
	GetPeersAddrMap() map[string]pb.PeerAddr
	GetPeersWithTemp() []*Peer
	GetAliveNodeNum() int
	ToRoutingTable() pb.Routers
	ToRoutingTableWithout(hash string) pb.Routers
	MergeFromRoutersToTemp(routers pb.Routers)
	MergeTempPeers(peer *Peer)
	MergeTempPeersForNewNode()
	RejectTempPeers()
	DeletePeer(peer *Peer) map[string]pb.PeerAddr
	PeerSetter
}

type PeerSetter interface {
	SetConnectionByHash(hash string, conn *grpc.ClientConn) error
	SetClientByHash(hash string, client pb.ChatClient) error
}
