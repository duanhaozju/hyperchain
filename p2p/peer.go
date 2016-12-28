//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package p2p

import (
	"errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "hyperchain/p2p/peermessage"
	"hyperchain/p2p/transport"
	"strconv"
	"sync"
	"time"
	"hyperchain/membersrvc"
	"hyperchain/core/crypto/primitives"
)

// init the package-level logger system,
// after this declare and init function,
// you can use the `log` whole the package scope

type Peer struct {
	PeerAddr *pb.PeerAddr
	Connection *grpc.ClientConn
	LocalAddr  *pb.PeerAddr
	Client     pb.ChatClient
	TEM        transport.TransportEncryptManager
	Status     int
	chatMux    sync.Mutex
	IsPrimary  bool
	//PeerPool   PeersPool
	Certificate string

}

// NewPeer to create a Peer which with a connection
// the peer will auto store into the peer pool.
// when creating a peer, the client instance will create a message whose type is HELLO
// if get a response, save the peer into singleton peer pool instance
// NewPeer 用于返回一个新的NewPeer 用于与远端的peer建立连接，这个peer将会存储在peerPool中
// 如果取得相应的连接返回值，将会将peer存储在单例的PeersPool中进行存储
func NewPeer(peerAddr *pb.PeerAddr,localAddr *pb.PeerAddr,TEM transport.TransportEncryptManager) (*Peer, error){
	var peer Peer
	peer.TEM = TEM
	log.Critical("TEM",TEM)
	peer.LocalAddr = localAddr
	peer.PeerAddr = peerAddr
	//peer.PeerPool = peerspool
	//TODO rewrite the tls options get method
	opts := membersrvc.GetGrpcClientOpts()
	// dial to remote
	conn, err := grpc.Dial(peerAddr.IP + ":" + strconv.Itoa(peerAddr.Port), opts...)
	if err != nil {
		log.Error("err:", errors.New("Cannot establish a connection!"))
		return nil,err
	}
	peer.Connection = conn
	peer.Client = pb.NewChatClient(conn)
	// set primary flag false
	peer.IsPrimary = false
	//review handshake operation
	err = peer.handShake()
	if err != nil{
		return nil, err
	}
	return &peer, nil
}

// handShake connect to remote peer, and negotiate the secret
// handShake 用于与相应的远端peer进行通信，并进行密钥协商
func (peer *Peer) handShake() (err error) {

	eca,getErr1 := primitives.GetConfig("../config/cert/server/eca.cert")
	if getErr1 != nil{
		log.Error("cannot read ecert.",err)
	}
	ecertBtye := []byte(eca)

	rca,getErr2 := primitives.GetConfig("../config/cert/server/rca.cert")
	if getErr2 != nil{
		log.Error("cannot read rcert.",err)
	}
	rcertByte := []byte(rca)

	signature := pb.Signature{
		Ecert:ecertBtye,
		Rcert:rcertByte,
	}

	//review start exchange the secret
	helloMessage := pb.Message{
		MessageType:  pb.Message_HELLO,
		Payload:      peer.TEM.GetLocalPublicKey(),
		From:         peer.LocalAddr.ToPeerAddress(),
		MsgTimeStamp: time.Now().UnixNano(),
		Signature: &signature,
	}
	retMessage, err := peer.Client.Chat(context.Background(), &helloMessage)

	//retMessage.Signature.Ecert

	if err != nil {
		log.Error("cannot establish a connection",err)
		return
	}
	//review get the remote peer secret
	if retMessage.MessageType == pb.Message_HELLO_RESPONSE {
		remotePublicKey := retMessage.Payload
		err = peer.TEM.GenerateSecret(remotePublicKey, peer.PeerAddr.Hash)
		if err != nil {
			log.Error("Local Generate Secret Failed, localAddr:",peer.LocalAddr,err)
			return
		}
		return  nil
	}
	return errors.New("ret message is not Hello Response!")
}

func NewPeerReconnect(peerAddr *pb.PeerAddr,localAddr *pb.PeerAddr,TEM transport.TransportEncryptManager) (peer *Peer, err error) {
	peer.TEM = TEM
	peer.PeerAddr = peerAddr
	peer.LocalAddr = localAddr
	//peer.PeerPool = peerPool

	opts := membersrvc.GetGrpcClientOpts()
	// dial to remote
	conn, err := grpc.Dial(peerAddr.IP + ":" + strconv.Itoa(peerAddr.Port), opts...)
	if err != nil {
		log.Error("err:", errors.New("Cannot establish a connection!"))
		return nil, err
	}

	peer.Connection = conn
	peer.Client = pb.NewChatClient(conn)
	// set the primary flag
	peer.IsPrimary = false
	// review handshake operation
	// review start exchange the secret
	helloMessage := pb.Message{
		MessageType:  pb.Message_RECONNECT,
		Payload:      peer.TEM.GetLocalPublicKey(),
		From:         peer.LocalAddr.ToPeerAddress(),
		MsgTimeStamp: time.Now().UnixNano(),
	}

	retMessage, err := peer.Client.Chat(context.Background(), &helloMessage)
	log.Debug("reconnect return :", retMessage)
	if err != nil {
		log.Error("cannot establish a connection", err)
		return nil, err
	}
	//review get the remote peer secrets
	if retMessage.MessageType == pb.Message_RECONNECT_RESPONSE {
		remotePublicKey := retMessage.Payload
		err = peer.TEM.GenerateSecret(remotePublicKey, peer.PeerAddr.Hash)
		if err != nil {
			log.Error("genErr", err)
			return nil,err
		}
		log.Debug("remote Peer address:", peer.PeerAddr)
		log.Debug(peer.TEM.GetSecret(peer.PeerAddr.Hash))
		return peer, nil
	}
	return nil, errors.New("cannot establish a connection")
}

// Chat is a function to send a message to peer,
// this function invokes the remote function peer-to-peer,
// which implements the service that prototype file declares
//
func (this *Peer) Chat(msg pb.Message) (response *pb.Message, err error) {
	log.Debug("BROADCAST:", msg.From.ID, ">>>", this.PeerAddr.ID)
	msg.Payload,err = this.TEM.EncWithSecret(msg.Payload, this.PeerAddr.Hash)
	if err != nil{
		log.Error("enc with secret ",err)
		return nil,err
	}
	response, err = this.Client.Chat(context.Background(), &msg)
	if err != nil {
		this.Status = 2;
		log.Error("response err:", err)
		//TODO
		//panic(err)
		return nil,err
	}
	this.Status = 1;
	// decode the return message
	if response != nil && response.MessageType != pb.Message_HELLO && response.MessageType != pb.Message_HELLO_RESPONSE{
		response.Payload,err = this.TEM.DecWithSecret(response.Payload, response.From.Hash)
		if err != nil{
			log.Error("decwithSec err:", err)
			return nil,err
		}
	}
	log.Debug("response",response.MessageType)
	return response, err
}

// Close the peer connection
// this function should ensure no thread use this thead
// this is not thread safety
func (this *Peer) Close() (bool, error) {
	err := this.Connection.Close()
	if err != nil {
		log.Error("err:", err)
		return false, err
	} else {
		return true, nil
	}
}
