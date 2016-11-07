// author: chenquan
// date: 16-8-25
// last modified: 16-8-29 13:58
// last Modified Author: chenquan
// change log: add a comment of this file function
//

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
}

// NewPeerByIpAndPort to create a Peer which with a connection,
// this connection ip string format is like '192.168.1.1'
// the peer will auto store into the peer pool.
// when creating a peer, the client instance will create a message whose type is HELLO
// if get a response, save the peer into singleton peer pool instance
func NewPeerByIpAndPort(ip string, port int64, nid uint64, TEM transport.TransportEncryptManager, localAddr *pb.PeerAddress) (*Peer, error) {
	var peer Peer
	peer.TEM = TEM
	peer.localAddr = localAddr
	peer.ID = nid
	peer.RemoteAddr = peerComm.ExtractAddress(ip, port, nid)
	opts := membersrvc.GetGrpcClientOpts()
	conn, err := grpc.Dial(ip + ":" + strconv.Itoa(int(port)), opts...)
	//conn, err := grpc.Dial(ip+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	if err != nil {
		errors.New("Cannot establish a connection!")
		log.Error("err:", err)
		return nil, err
	}
	peer.Connection = conn
	peer.Client = pb.NewChatClient(peer.Connection)
	peer.IsPrimary = false
	//TODO handshake operation
	peer.handShake()
	return &peer, nil
}

func NewPeerByAddress(address *pb.PeerAddress, nid uint64, TEM transport.TransportEncryptManager, localAddr *pb.PeerAddress) (*Peer, error) {
	var peer Peer
	peer.TEM = TEM
	peer.localAddr = localAddr
	peer.RemoteAddr = address
	opts := membersrvc.GetGrpcClientOpts()
	conn, err := grpc.Dial(address.IP + ":" + strconv.Itoa(int(address.Port)), opts...)
	//conn, err := grpc.Dial(ip+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	if err != nil {
		errors.New("Cannot establish a connection!")
		log.Error("err:", err)
		return nil, err
	}
	peer.Connection = conn
	peer.Client = pb.NewChatClient(peer.Connection)
	peer.IsPrimary = false
	log.Warning("newpeer :", peer)
	//TODO handshake operation
	peer.handShake()
	return &peer, nil
}

func (peer *Peer) handShake() {
	//review 开始交换秘钥
	helloMessage := pb.Message{
		MessageType:  pb.Message_HELLO,
		Payload:      peer.TEM.GetLocalPublicKey(),
		From:         peer.localAddr,
		MsgTimeStamp: time.Now().UnixNano(),
	}
	retMessage, err2 := peer.Client.Chat(context.Background(), &helloMessage)
	if err2 != nil {
		log.Error("cannot establish a connection")
		return
	} else {
		//review 取得对方的秘钥
		if retMessage.MessageType == pb.Message_HELLO_RESPONSE {
			remotePublicKey := retMessage.Payload
			genErr := peer.TEM.GenerateSecret(remotePublicKey, peer.RemoteAddr.Hash)
			if genErr != nil {
				log.Error("genErr", genErr)
			}

			log.Notice("secret", len(peer.TEM.GetSecret(peer.RemoteAddr.Hash)))
			peer.ID = retMessage.From.ID
			log.Notice("节点:", peer.RemoteAddr.ID)
			log.Notice("hash:", peer.RemoteAddr.Hash)
			log.Notice("协商秘钥：")
			log.Notice(peer.TEM.GetSecret(peer.RemoteAddr.Hash))
			return
		}
	}
}

// Chat is a function to send a message to peer,
// this function invokes the remote function peer-to-peer,
// which implements the service that prototype file declares
//
func (this *Peer) Chat(msg pb.Message) (*pb.Message, error) {
	log.Debug("Invoke the broadcast method", msg.From.ID, ">>>", this.RemoteAddr.ID)
	this.chatMux.Lock()
	defer this.chatMux.Unlock()
	msg.Payload = this.TEM.EncWithSecret(msg.Payload, this.RemoteAddr.Hash)
	r, err := this.Client.Chat(context.Background(), &msg)
	if err != nil {
		log.Error("err:", err)
	}
	// 返回信息解密
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
