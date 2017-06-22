//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package p2p

import (
	"hyperchain/p2p/transport"
	"hyperchain/p2p/network"
	"hyperchain/p2p/info"
	pb "hyperchain/p2p/message"
	"hyperchain/manager/event"
)

// init the package-level logger system,
// after this declare and init function,
// you can use the `log` whole the package scope
type Peer struct {
	hostname string
	info *info.Info
	namespace string
	net *network.HyperNet
	TM         *transport.TransportManager
	p2pHub event.TypeMux
}
//NewPeer get a new peer which chat/greeting/whisper functions
func NewPeer(namespace string,hostname string,id int,net *network.HyperNet)*Peer {
	return &Peer{
		info:info.NewInfo(id,hostname,namespace),
		namespace:namespace,
		hostname:hostname,
		net:net,
	}
}

// release 1.3 feature
//Chat send a stream message to remote peer
func (peer *Peer)Chat(in *pb.Message) (*pb.Message, error){
	//here will wrapper the message
	in.From = &pb.Endpoint{
		Filed : []byte(peer.namespace),
		Hostname : []byte(peer.hostname),
		UUID : []byte(peer.info.GetHash()),
		Version : 13,
	}
	response,err := peer.net.Greeting(peer.hostname,in)
	if err != nil{
		return nil,err
	}
	return response,nil
}

//Greeting send a single message to remote peer
func (peer *Peer)Greeting(in *pb.Message) (*pb.Message,error){
	response,err := peer.net.Greeting(peer.hostname,in)
	if err != nil{
		return nil,err
	}
	return response,nil
}

//Whisper send a whisper message to remote peer
func (peer *Peer)Whisper(in *pb.Message) (*pb.Message,error){
	response,err := peer.net.Greeting(peer.hostname,in)
	if err != nil{
		return nil,err
	}
	return response,nil
}


