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
	"github.com/pkg/errors"
	"hyperchain/crypto/csprng"
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

//implements the WeightItem interface
func(peer *Peer)Weight()int{
	return peer.info.Id
}
func(peer *Peer)Value()interface{}{
	return peer
}

//Chat send a stream message to remote peer
func (peer *Peer) Chat(in *pb.Message) (*pb.Message, error) {
	fmt.Println("Chat msg to ",peer.info.Hostname)
	//here will wrapper the message
	in.From = &pb.Endpoint{
		Field:    []byte(peer.namespace),
		Hostname: []byte(peer.local.Hostname),
		UUID:     []byte(peer.local.GetHash()),
		Version:  P2P_MODULE_DEV_VERSION,
	}
	//encrypt
	encPayload,err := peer.chts.Encrypt(in.Payload)
	if err != nil{
		return nil,err
	}
	in.Payload = encPayload

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
	//encrypt
	encPayload,err := peer.chts.Encrypt(in.Payload)
	if err != nil{
		return nil,err
	}
	in.Payload = encPayload
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
	/*
	 ^ClientHello
	  *ClientCertificate ==> e cert r cert
	  *ClientSignature ==> e sign r sign
	  *ClientCipher ==> rand
	  *ClientKeyExchange ==> ignored
	  */
	data := []byte("hyperchain")
	esign, err := peer.chts.CG.ESign(data)
	if err !=nil{
		return err
	}
	rsign, err := peer.chts.CG.RSign(data)
	if err !=nil{
		return err
	}
	rand,err := csprng.CSPRNG(32)
	if err !=nil{
		return err
	}
	certpayload,err := payloads.NewCertificate(data,peer.chts.CG.GetECert(),esign,peer.chts.CG.GetRCert(),rsign,rand)
	if err !=nil{
		return err
	}
	// peer should
	identify := payloads.NewIdentify(peer.local.IsVP,isOriginal,peer.namespace, peer.local.Hostname, peer.local.Id,certpayload)
	payload, err := identify.Serialize()
	if err != nil {
		return err
	}
	msg := pb.NewMsg(pb.MsgType_CLIENTHELLO, payload)

	serverHello, err := peer.Greeting(msg)

	fmt.Printf("peer.go 205 got a server hello message %+v \n", serverHello)
	if err != nil {
		fmt.Printf("peer.go 205  err: %s \n",err.Error())
		return err
	}
	// complele the key agree
	if err := peer.negotiateShareKey(serverHello,rand);err != nil{
		fmt.Printf("peer.go 210 error: %s \n",err.Error())
		return peer.clientReject(serverHello)
	}else {
		return peer.clientResponse(serverHello)
	}

}


func (peer *Peer)negotiateShareKey(in *pb.Message,rand []byte) error{
	/*
	   ^ServerReject
                or
            ^ServerHello
             *ServerCertificate
	     *ServerSignature
             *ServerCipherSpec
             *ServerKeyExchange
             */
	if in == nil || in.Payload == nil{
		return errors.New("invalid server return message")
	}

	iden,err := payloads.IdentifyUnSerialize(in.Payload)
	if err !=nil{
		return err
	}
	//TODO Check the identity is legal or not
	if  iden.Payload == nil{
		return errors.New("iden payload is nil")
	}
	cert,err := payloads.CertificateUnMarshal(iden.Payload)
	if err != nil{
		return err
	}
	r := append(cert.Rand,rand...)
	return peer.chts.GenShareKey(r,cert.ECert)
}


// handle the double side handshake process,
// when got a serverhello, this peer should response by clientResponse.
func (peer *Peer) clientResponse(serverHello *pb.Message) error {
	payload := []byte("client accept [msg test]")
	msg := pb.NewMsg(pb.MsgType_CLIENTACCEPT, payload)
	serverdone, err := peer.Greeting(msg)
	fmt.Printf("peer.go162 got a server done message %+v \n", serverdone)
	if err != nil {
		return err
	}
	return nil
}

// handle the double side handshake process,
// when got a serverhello, this peer should response by clientResponse.
func (peer *Peer) clientReject(serverHello *pb.Message) error {
	payload := []byte("client accept [msg test]")
	msg := pb.NewMsg(pb.MsgType_CLIENTREJECT, payload)
	serverdone, err := peer.Greeting(msg)
	fmt.Printf("peer.go162 got a server done message %+v \n", serverdone)
	if err != nil {
		return err
	}
	return nil
}
