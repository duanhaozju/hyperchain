//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package p2p

import (
	"encoding/json"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/crypto/csprng"
	"github.com/hyperchain/hyperchain/manager/event"
	"github.com/hyperchain/hyperchain/p2p/hts"
	"github.com/hyperchain/hyperchain/p2p/info"
	pb "github.com/hyperchain/hyperchain/p2p/message"
	"github.com/hyperchain/hyperchain/p2p/network"
	"github.com/hyperchain/hyperchain/p2p/payloads"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
)

type Peer struct {
	hostname  string
	info      *info.Info
	local     *info.Info
	namespace string
	net       *network.HyperNet
	p2pHub    *event.TypeMux
	chts      *hts.ClientHTS
	logger    *logging.Logger
}

// NewPeer create and returns a new Peer instance.
func NewPeer(namespace string, hostname string, id int, localInfo *info.Info, net *network.HyperNet, chts *hts.ClientHTS, evhub *event.TypeMux) (*Peer, error) {
	peer := &Peer{
		info:      info.NewInfo(id, hostname, namespace),
		namespace: namespace,
		hostname:  hostname,
		net:       net,
		local:     localInfo,
		chts:      chts,
		p2pHub:    evhub,
		logger:    common.GetLogger(namespace, "p2p"),
	}
	if err := peer.clientHello(peer.local.IsOrg(), peer.local.IsRec()); err != nil {
		return nil, err
	}
	return peer, nil
}

// Weight returns the node ID.
func (peer *Peer) Weight() int {
	return peer.info.Id
}

func (peer *Peer) Value() interface{} {
	return peer
}

// Chat sends a stream message to remote peer.
func (peer *Peer) Chat(in *pb.Message) (*pb.Message, error) {
	peer.logger.Debug("Chat msg to ", peer.info.Hostname)
	//here will wrapper the message
	in.From = &pb.Endpoint{
		Field:    []byte(peer.namespace),
		Hostname: []byte(peer.local.Hostname),
		UUID:     []byte(peer.local.GetHash()),
		Version:  P2P_MODULE_DEV_VERSION,
	}
	//encrypt
	encPayload, err := peer.chts.Encrypt(in.Payload)
	if err != nil {
		return nil, err
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

// Whisper sends a whisper message to remote peer.
func (peer *Peer) Whisper(in *pb.Message) (*pb.Message, error) {
	in.From = &pb.Endpoint{
		Field:    []byte(peer.local.GetNameSpace()),
		Hostname: []byte(peer.local.Hostname),
		UUID:     []byte(peer.local.GetHash()),
		Version:  P2P_MODULE_DEV_VERSION,
	}
	//encrypt
	encPayload, err := peer.chts.Encrypt(in.Payload)
	if err != nil {
		return nil, err
	}
	in.Payload = encPayload
	response, err := peer.net.Whisper(peer.hostname, in)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// Greeting sends a greeting message to remote peer.
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

// Serialize serializes the peer information.
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

// PeerDeSerialize deserializes the peer information.
func PeerDeSerialize(raw []byte) (hostname string, namespace string, hash string, err error) {
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

// session key negotiation
func (peer *Peer) clientHello(isOrg, isRec bool) error {
	peer.logger.Debug("send client hello message")
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
	if err != nil {
		return err
	}
	rsign, err := peer.chts.CG.RSign(data)
	if err != nil {
		return err
	}
	rand, err := csprng.CSPRNG(32)
	if err != nil {
		return err
	}
	certpayload, err := payloads.NewCertificate(data, peer.chts.CG.GetECert(), esign, peer.chts.CG.GetRCert(), rsign, rand)
	if err != nil {
		return err
	}

	identify := payloads.NewIdentify(peer.local.IsVP, isOrg, isRec, peer.namespace, peer.local.Hostname, peer.local.Id, certpayload)
	payload, err := identify.Serialize()
	if err != nil {
		return err
	}
	msg := pb.NewMsg(pb.MsgType_CLIENTHELLO, payload)
	serverHello, err := peer.Greeting(msg)
	if err != nil {
		return err
	}
	peer.logger.Debugf("got a server hello message msgtype %s  ", serverHello.MessageType)
	// complele the key agree
	if err := peer.negotiateShareKey(serverHello, rand); err != nil {
		return peer.clientReject(serverHello)
	} else {
		return peer.clientResponse(serverHello)
	}

}

func (peer *Peer) negotiateShareKey(in *pb.Message, rand []byte) error {
	/*
			   ^ServerReject
		                or
		            ^ServerHello
		             *ServerCertificate
			     *ServerSignature
		             *ServerCipherSpec
		             *ServerKeyExchange
	*/
	if in == nil || in.Payload == nil {
		return errors.New("invalid server return message")
	}

	iden, err := payloads.IdentifyUnSerialize(in.Payload)
	if err != nil {
		return err
	}
	//TODO Check the identity is legal or not
	if iden.Payload == nil {
		return errors.New("iden payload is nil")
	}
	cert, err := payloads.CertificateUnMarshal(iden.Payload)
	if err != nil {
		return err
	}
	r := append(cert.Rand, rand...)
	err = peer.chts.GenShareKey(r, cert.ECert)
	peer.logger.Debugf(`
Client nego key
Local Hostname: %s
Local Hash %s
Peer hostname %s
Peer hash %s
server Rand %s
client Rand %s
Total rand %s
Sharekey %s
`,
		peer.local.Hostname,
		peer.local.Hash,
		peer.info.Hostname,
		peer.info.Hash,
		common.ToHex(cert.Rand),
		common.ToHex(rand),
		common.ToHex(r),
		common.ToHex(peer.chts.GetSK()))
	return err

}

// handle the double side handshake process,
// when got a serverhello, if accept, this peer should response by clientResponse.
func (peer *Peer) clientResponse(serverHello *pb.Message) error {
	payload := []byte("client accept [msg test]")
	msg := pb.NewMsg(pb.MsgType_CLIENTACCEPT, payload)
	serverdone, err := peer.Greeting(msg)
	peer.logger.Debugf("got a server done message %+v ", serverdone)
	if err != nil {
		return err
	}
	return nil
}

// handle the double side handshake process,
// when got a serverhello, if reject, this peer should response by clientReject.
func (peer *Peer) clientReject(serverHello *pb.Message) error {
	payload := []byte("client accept [msg test]")
	msg := pb.NewMsg(pb.MsgType_CLIENTREJECT, payload)
	serverdone, err := peer.Greeting(msg)
	peer.logger.Debug("got a server done message %+v ", serverdone)
	if err != nil {
		return err
	}
	return nil
}
