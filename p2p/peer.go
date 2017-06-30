//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package p2p

import (
	"hyperchain/p2p/transport"
	"hyperchain/p2p/network"
	"hyperchain/p2p/info"
	pb "hyperchain/p2p/message"
	"hyperchain/manager/event"
	"encoding/json"
)

// init the package-level logger system,
// after this declare and init function,
// you can use the `log` whole the package scope
type Peer struct {
	hostname string
	info *info.Info
	local *info.Info
	namespace string
	net *network.HyperNet
	TM         *transport.TransportManager
	p2pHub event.TypeMux
}
//NewPeer get a new peer which chat/greeting/whisper functions
func NewPeer(namespace string,hostname string,id int,localInfo *info.Info,net *network.HyperNet)*Peer {
	return &Peer{
		info:info.NewInfo(id,hostname,namespace),
		namespace:namespace,
		hostname:hostname,
		net:net,
		local:localInfo,
	}
}

// release 1.3 feature
//Chat send a stream message to remote peer
func (peer *Peer)Chat(in *pb.Message) (*pb.Message, error){
	//here will wrapper the message
	in.From = &pb.Endpoint{
		Field : []byte(peer.namespace),
		Hostname : []byte(peer.local.Hostname),
		UUID : []byte(peer.local.GetHash()),
		Version : P2P_MODULE_DEV_VERSION,
	}

	//TODO here should change to Chat method
	//TODO change as bidi stream transfer method
	resp,err := peer.net.Whisper(peer.hostname,in)
	if err != nil{
		return nil,err
	}
	return resp,nil
}

//Greeting send a single message to remote peer
func (peer *Peer)Greeting(in *pb.Message) (*pb.Message,error){
	//here will wrapper the message
	in.From = &pb.Endpoint{
		
		Field : []byte(peer.local.GetNameSpace()),
		Hostname : []byte(peer.local.Hostname),
		UUID : []byte(peer.local.GetHash()),
		Version : P2P_MODULE_DEV_VERSION,
	}
	response,err := peer.net.Greeting(peer.hostname,in)
	if err != nil{
		return nil,err
	}
	return response,nil
}

//Whisper send a whisper message to remote peer
func (peer *Peer)Whisper(in *pb.Message) (*pb.Message,error){
	in.From = &pb.Endpoint{
		Field : []byte(peer.local.GetNameSpace()),
		Hostname : []byte(peer.local.Hostname),
		UUID : []byte(peer.local.GetHash()),
		Version : P2P_MODULE_DEV_VERSION,
	}
	response,err := peer.net.Greeting(peer.hostname,in)
	if err != nil{
		return nil,err
	}
	return response,nil
}


func (peer *Peer)Serialize()[]byte{
	ps := struct {
		Hostname string `json:"hostname"`
		Namespace string `json:"namespace"`
		Hash string `json:"hash"`
	}{}
	ps.Hash = peer.info.GetHash()
	ps.Hostname = peer.hostname
	ps.Namespace = peer.namespace
	b,err := json.Marshal(ps)
	if err != nil{
		return nil
	}
	return b
}

func PeerUnSerialize(raw []byte)(hostname string,namespace string,hash string,err error){
	ps := &struct {
		Hostname string `json:"hostname"`
		Namespace string `json:"namespace"`
		Hash string `json:"hash"`
	}{}
	err = json.Unmarshal(raw,ps)
	if err != nil{
		return "","","",err
	}
	return ps.Hostname,ps.Namespace,ps.Hash,nil
}
