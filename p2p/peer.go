//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package p2p

import (
	"encoding/json"
	"fmt"
	"hyperchain/manager/event"
	"hyperchain/p2p/hts"
	"hyperchain/p2p/info"
	pb "hyperchain/p2p/message"
	"hyperchain/p2p/network"
	"hyperchain/p2p/payloads"
	"hyperchain/p2p/transport"
)

// init the package-level logger system,
// after this declare and init function,
// you can use the `log` whole the package scope
type Peer struct {
	hostname  string
	info      *info.Info
	local     *info.Info
	namespace string
	net       *network.HyperNet
	TM        *transport.TransportManager
	p2pHub    event.TypeMux
	chts      *hts.ClientHTS
}

//NewPeer get a new peer which chat/greeting/whisper functions
func NewPeer(namespace string, hostname string, id int, localInfo *info.Info, net *network.HyperNet, chts *hts.ClientHTS) (*Peer, error) {
	peer := &Peer{
		info:      info.NewInfo(id, hostname, namespace),
		namespace: namespace,
		hostname:  hostname,
		net:       net,
		local:     localInfo,
		chts:      chts,
	}
	if err := peer.clientHello(peer.info.GetOriginal()); err != nil {
		return nil, err
	}
	return peer, nil
}

//Chat send a stream message to remote peer
func (peer *Peer) Chat(in *pb.Message) (*pb.Message, error) {
	//here will wrapper the message
	in.From = &pb.Endpoint{
		Field:    []byte(peer.namespace),
		Hostname: []byte(peer.local.Hostname),
		UUID:     []byte(peer.local.GetHash()),
		Version:  P2P_MODULE_DEV_VERSION,
	}

	//TODO here should change to Chat method
	//TODO change as bidi stream transfer method
	resp, err := peer.net.Whisper(peer.hostname, in)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

//Whisper send a whisper message to remote peer
func (peer *Peer) Whisper(in *pb.Message) (*pb.Message, error) {
	in.From = &pb.Endpoint{
		Field:    []byte(peer.local.GetNameSpace()),
		Hostname: []byte(peer.local.Hostname),
		UUID:     []byte(peer.local.GetHash()),
		Version:  P2P_MODULE_DEV_VERSION,
	}
	response, err := peer.net.Whisper(peer.hostname, in)
	if err != nil {
		return nil, err
	}
	return response, nil
}

//Greeting send a greeting message to remote peer
func (peer *Peer) Greeting(in *pb.Message) (*pb.Message, error) {
	in.From = &pb.Endpoint{
		Field:    []byte(peer.local.GetNameSpace()),
		Hostname: []byte(peer.local.Hostname),
		UUID:     []byte(peer.local.GetHash()),
		Version:  P2P_MODULE_DEV_VERSION,
	}
	response, err := peer.net.Greeting(peer.hostname, in)
	if err != nil {
		return nil, err
	}
	return response, nil
}

//Serialize the peer
func (peer *Peer) Serialize() []byte {
	ps := struct {
		Hostname  string `json:"hostname"`
		Namespace string `json:"namespace"`
		Hash      string `json:"hash"`
	}{}
	ps.Hash = peer.info.GetHash()
	ps.Hostname = peer.hostname
	ps.Namespace = peer.namespace
	b, err := json.Marshal(ps)
	if err != nil {
		return nil
	}
	return b
}

//Unserialize the peer
func PeerUnSerialize(raw []byte) (hostname string, namespace string, hash string, err error) {
	ps := &struct {
		Hostname  string `json:"hostname"`
		Namespace string `json:"namespace"`
		Hash      string `json:"hash"`
	}{}
	err = json.Unmarshal(raw, ps)
	if err != nil {
		return "", "", "", err
	}
	return ps.Hostname, ps.Namespace, ps.Hash, nil
}

/*
	Client                                         Server
	^ClientHello
	  *ClientCertificate
	  *ClientSignature      ------------>        Listening
	  *ClientCipher
	  *ClientKeyExchange

	                                            ^ServerReject
	                                                or
	                                            ^ServerHello
	                                             *ServerCertificate
	                       <------------         *ServerSignature
	                                             *ServerCipherSpec
	                                             *ServerKeyExchange

	  ^ClientAccept
	      or               ------------->
	  ^ClientReject

	                                            ^ServerDone
                               <------------
*/

//this is peer should do things
func (peer *Peer) clientHello(isOriginal bool) error {
	fmt.Println("peer.go 152 send client hello message")
	// if self return nil do not need verify
	if peer.info.Hostname == peer.local.Hostname {
		return nil
	}

	// peer should
	payload := []byte("client hello[msg test]")
	msg := pb.NewMsg(pb.MsgType_CLIENTHELLO, payload)
	identify := payloads.NewIdentify(peer.local.IsVP,isOriginal,peer.namespace, peer.local.Hostname, peer.local.Id)
	payload, err := identify.Serialize()
	if err != nil {
		return err
	}
	msg.Payload = payload
	serverHello, err := peer.Greeting(msg)
	fmt.Printf("peer.go 151 got a server hello message %+v \n", serverHello)
	if err != nil {
		return err
	}
	return peer.clientResponse(serverHello)
}

func (peer *Peer) clientResponse(serverHello *pb.Message) error {
	fmt.Println("peer.go 152 send client accept message")
	payload := []byte("client accept [msg test]")
	msg := pb.NewMsg(pb.MsgType_CLIENTACCEPT, payload)
	serverdone, err := peer.Greeting(msg)
	fmt.Printf("peer.go 162 got a server done message %+v \n", serverdone)
	if err != nil {
		return err
	}

	return nil
}
