//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package p2p

import (
	pb "hyperchain/p2p/peermessage"
	"google.golang.org/grpc"
)

//PeerPool provide the basic peer manager function
type PeersPool interface{
	PutPeer(address pb.PeerAddr, client *Peer) error
	PutPeerToTemp(address pb.PeerAddr, client *Peer) error
	GetPeer(address pb.PeerAddr) *Peer
	GetPeerByHash(hash string) (*Peer,error)
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
	PeerSetter
}

type PeerSetter interface {
	SetConnectionByHash(hash string,conn *grpc.ClientConn) error
	SetClientByHash(hash string, client pb.ChatClient) error
}