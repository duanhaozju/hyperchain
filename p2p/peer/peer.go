// author: chenquan
// date: 16-8-25
// last modified: 16-8-29 13:58
// last Modified Author: chenquan
// change log: add a comment of this file function
//

package client

import (
	"google.golang.org/grpc"
	pb "hyperchain/p2p/peermessage"
	"errors"
	"golang.org/x/net/context"
	"strconv"
	"github.com/op/go-logging"
	"hyperchain/p2p/transport"
	"hyperchain/p2p/peerComm"
	"time"
	"sync"
	"encoding/hex"
)

// init the package-level logger system,
// after this declare and init function,
// you can use the `log` whole the package scope
var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("p2p/Server")
}


type Peer struct {
	Addr *pb.PeerAddress
	Connection *grpc.ClientConn
	Client pb.ChatClient
	CName string
	TEM transport.TransportEncryptManager
	Status int
	ID int
	chatMux sync.Mutex
}

// NewPeerByIpAndPort to create a Peer which with a connection,
// this connection ip string format is like '192.168.1.1'
// the peer will auto store into the peer pool.
// when creating a peer, the client instance will create a message whose type is HELLO
// if get a response, save the peer into singleton peer pool instance
func NewPeerByIpAndPort(ip string,port int32,nid int32,TEM transport.TransportEncryptManager,localAddr *pb.PeerAddress)(*Peer,error){
	var peer Peer
	peer.TEM =TEM
	peerAddr := peerComm.ExtractAddress(ip,int(port),nid)

	conn, err := grpc.Dial(ip + ":" + strconv.Itoa(int(port)), grpc.WithInsecure())
	if err != nil {
		errors.New("Cannot establish a connection!")
		log.Error("err:",err)
		return nil,err
	}
	peer.Connection = conn
	peer.Client = pb.NewChatClient(peer.Connection)
	peer.Addr = peerAddr
	//TODO handshake operation
	//peer.TEM = transport.NewHandShakeManger()
	//package the information
	//review 开始交换秘钥
	helloMessage := pb.Message{
		MessageType:pb.Message_HELLO,
		Payload:peer.TEM.GetLocalPublicKey(),
		From:localAddr,
		MsgTimeStamp:time.Now().UnixNano(),
	}
	retMessage,err2 := peer.Client.Chat(context.Background(),&helloMessage)
	if err2 != nil{
		log.Error("cannot establish a connection")
		return nil,err2
	}else{
		//review 取得对方的秘钥
		if retMessage.MessageType == pb.Message_HELLO_RESPONSE {
			remotePublicKey := retMessage.Payload
			peer.TEM.GenerateSecret(remotePublicKey,peer.Addr.Hash)
			log.Critical("secret",len(peer.TEM.GetSecret(peer.Addr.Hash)))
			peer.ID = int(retMessage.From.ID)
			if err != nil{
				log.Error("cannot decrypt the nodeidinfo!")
				errors.New("Decrypt ERROR")
			}
			log.Notice("节点:",peer.Addr.ID)
			log.Notice("hash:",peer.Addr.Hash)
			log.Notice("协商秘钥：")
			log.Notice(peer.TEM.GetSecret(peer.Addr.Hash))
			return &peer,nil
		}
	}
	return nil,errors.New("cannot establish a connection")
}

// Chat is a function to send a message to peer,
// this function invokes the remote function peer-to-peer,
// which implements the service that prototype file declares
//
func (this *Peer)Chat(msg *pb.Message) (*pb.Message, error){
	this.chatMux.Lock()
	defer this.chatMux.Unlock()
	//review encrypt
	//log.Notice("sec pool size ",this.TEM.GetSceretPoolSize())
	//this.TEM.PrintAllSecHash()
	log.Notice("消息发往>>>",this.Addr.ID)
	//log.Notice("消息发往秘钥>>>",this.TEM.GetSecret(this.Addr.Hash))
	//msg.Payload=[]byte(strconv.Itoa(int(this.Addr.ID)))

	source := msg.Payload
	msg.Payload = this.TEM.EncWithSecret(msg.Payload,this.Addr.Hash)
	//log.Critical("××××××加密传输××××××")
	//log.Critical("to node ",this.Addr.ID)
	log.Critical("原信息:",hex.EncodeToString(source),"加密信息:",hex.EncodeToString(msg.Payload))
	r,err := this.Client.Chat(context.Background(),msg)
	if err != nil{
		log.Error("err:",err)
	}
	// 返回信息解密
	//if r.MessageType !=pb.Message_HELLO && r.MessageType != pb.Message_HELLO_RESPONSE{
	//	r.Payload = this.TEM.DecWithSecret(r.Payload,r.From.Hash)
	//}
	return r,err
}

// Close the peer connection
// this function should ensure no thread use this thead
// this is not thread safety
func (this *Peer)Close()(bool,error){
	err := this.Connection.Close()
	if err != nil{
		log.Error("err:",err)
		return false,err
	}else{
		return true,nil
	}
}