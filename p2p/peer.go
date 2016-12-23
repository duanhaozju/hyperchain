//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package p2p

import (
	"errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"hyperchain/p2p/peerComm"
	pb "hyperchain/p2p/peermessage"
	"hyperchain/p2p/transport"
	"strconv"
	"sync"
	"time"
	//"hyperchain/membersrvc"
	"hyperchain/membersrvc"
)

// init the package-level logger system,
// after this declare and init function,
// you can use the `log` whole the package scope

type Peer struct {
	RemoteAddr *pb.PeerAddress
	Connection *grpc.ClientConn
	localAddr  *pb.PeerAddress
	Client     pb.ChatClient
	TEM        transport.TransportEncryptManager
	Status     int
	ID         uint64
	chatMux    sync.Mutex
	IsPrimary  bool
	PeerPool   *PeersPool
	Addr pb.PeerAddress
	Certificate string

}

// NewPeerByIpAndPort to create a Peer which with a connection,
// this connection ip string format is like '192.168.1.1'
// the peer will auto store into the peer pool.
// when creating a peer, the client instance will create a message whose type is HELLO
// if get a response, save the peer into singleton peer pool instance

func NewPeerByIpAndPort(ip string, port int64,rpcPort int64, nid uint64, TEM transport.TransportEncryptManager, localAddr *pb.PeerAddress, peerPool *PeersPool) (*Peer, error) {
	var peer Peer
	peer.TEM = TEM
	peer.ID = nid
	peer.PeerPool = peerPool
	peerAddr := peerComm.ExtractAddress(ip, port, nid)
	peerAddr.RpcPort = rpcPort;
	opts:=membersrvc.GetGrpcClientOpts()
	conn, err := grpc.Dial(ip+":"+strconv.Itoa(int(port)), opts...)
	peer.localAddr = localAddr
	peer.RemoteAddr = peerComm.ExtractAddress(ip, port, nid)
	if err != nil {
		errors.New("Cannot establish a connection!")
		log.Error("err:", err)
		return nil, err
	}
	peer.Connection = conn
	peer.Client = pb.NewChatClient(peer.Connection)
	peer.IsPrimary = false
	//review handshake operation
	handShakeErr := peer.handShake()
	if handShakeErr != nil{
		return nil, handShakeErr
	}
	return &peer, nil
}

func NewPeerByAddress(address *pb.PeerAddress, nid uint64, TEM transport.TransportEncryptManager, localAddr *pb.PeerAddress) (*Peer, error) {
	var peer Peer
	peer.TEM = TEM
	peer.localAddr = localAddr
	peer.RemoteAddr = address
	opts := membersrvc.GetGrpcClientOpts()
	conn, err := grpc.Dial(address.IP + ":" + strconv.Itoa(int(address.Port)), opts...)
	if err != nil {
		errors.New("Cannot establish a connection!")
		log.Error("err:", err)
		return nil, err
	}
	peer.Connection = conn
	peer.Client = pb.NewChatClient(peer.Connection)
	peer.IsPrimary = false
	//review handshake operation
	handShakeErr := peer.handShake()
	if handShakeErr != nil{
		return nil, handShakeErr
	}
	return &peer, nil
}

func (peer *Peer) handShake() (err error) {
	//review start exchange the secret
	helloMessage := pb.Message{
		MessageType:  pb.Message_HELLO,
		Payload:      peer.TEM.GetLocalPublicKey(),
		From:         peer.localAddr,
		MsgTimeStamp: time.Now().UnixNano(),
	}
	retMessage, err := peer.Client.Chat(context.Background(), &helloMessage)
	if err != nil {
		log.Error("cannot establish a connection",err)
		return
	} else {
		//review get the remote peer secret
		if retMessage.MessageType == pb.Message_HELLO_RESPONSE {
			remotePublicKey := retMessage.Payload
			err := peer.TEM.GenerateSecret(remotePublicKey, peer.RemoteAddr.Hash)
			if err != nil {
				log.Error("genErr", err)
				return
			}

			log.Debug("secret", len(peer.TEM.GetSecret(peer.RemoteAddr.Hash)))
			peer.ID = retMessage.From.ID
			return  nil
		}
		err =  errors.New("ret message is not Hello Response!")
		return
	}
}

func NewPeerByIpAndPortReconnect(ip string, port int64,rpcPort int64, nid uint64, TEM transport.TransportEncryptManager, localAddr *pb.PeerAddress, peerPool *PeersPool) (*Peer, error) {
	var peer Peer
	peer.TEM = TEM
	peer.ID = nid
	peer.PeerPool = peerPool
	peerAddr := peerComm.ExtractAddress(ip, port, nid)
	peerAddr.RpcPort = rpcPort

	opts := membersrvc.GetGrpcClientOpts()
	conn, err := grpc.Dial(ip + ":" + strconv.Itoa(int(port)), opts...)
	if err != nil {
		errors.New("Cannot establish a connection!")
		log.Error("err:", err)
		return nil, err
	}
	peer.Connection = conn
	peer.Client = pb.NewChatClient(peer.Connection)
	peer.Addr = *peerAddr
	peer.IsPrimary = false
	// review handshake operation
	// review start exchange the secret
	helloMessage := pb.Message{
		MessageType:  pb.Message_RECONNECT,
		Payload:      peer.TEM.GetLocalPublicKey(),
		From:         localAddr,
		MsgTimeStamp: time.Now().UnixNano(),
	}
	retMessage, err2 := peer.Client.Chat(context.Background(), &helloMessage)
	if err2 != nil {
		log.Error("cannot establish a connection", err2)
		return nil, err2
	} else {
		//review get the remote peer secrets
		if retMessage.MessageType == pb.Message_RECONNECT_RESPONSE {
			remotePublicKey := retMessage.Payload
			genErr := peer.TEM.GenerateSecret(remotePublicKey, peer.Addr.Hash)
			if genErr != nil {
				log.Error("genErr", err)
			}

			log.Notice("secret", len(peer.TEM.GetSecret(peer.Addr.Hash)))
			peer.ID = retMessage.From.ID
			if err != nil {
				log.Error("cannot decrypt the nodeidinfo!")
				errors.New("Decrypt ERROR")
			}
			return &peer, nil
		}
	}
	return nil, errors.New("cannot establish a connection")
}

// Chat is a function to send a message to peer,
// this function invokes the remote function peer-to-peer,
// which implements the service that prototype file declares
//
func (this *Peer) Chat(msg pb.Message) (*pb.Message, error) {
	log.Debug("Invoke the broadcast method", msg.From.ID, ">>>", this.Addr.ID)

	msg.Payload = this.TEM.EncWithSecret(msg.Payload, this.Addr.Hash)
	r, err := this.Client.Chat(context.Background(), &msg)
	if err != nil {
		this.Status = 2;
		log.Error("err:", err)
		return nil,err
	} else {
		this.Status = 1;
	}
	// decode the return message
	if r != nil {
		if r.MessageType != pb.Message_HELLO && r.MessageType != pb.Message_HELLO_RESPONSE {
			r.Payload = this.TEM.DecWithSecret(r.Payload, r.From.Hash)
		}

	}
	return r, err
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
