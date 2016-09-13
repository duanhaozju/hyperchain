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
	"strings"
	"strconv"
	"github.com/op/go-logging"
	"hyperchain/p2p/transport"
)

// init the package-level logger system,
// after this declare and init function,
// you can use the `log` whole the package scope
var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("p2p/Server")
}
var DESKEY = []byte("sfe023f_sefiel#fi32lf3e!")


type Peer struct {
	Addr pb.PeerAddress
	Connection *grpc.ClientConn
	Client pb.ChatClient
	Idetity string
}

// NewPeerByString to create a Peer which with a connection,
// this connection address string format is '192.168.1.1:8001'
// you can also create a Peer by function `NewPeerByAddress(peerMessage.address)`
// the peer will auto store into the peer pool.
// when creating a peer, the client instance will create a message whose type is HELLO
// if get a response, save the peer into singleton peer pool instance
func NewPeerByString(address string)(*Peer,error){
	var peer Peer
	arr := strings.Split(address,":")
	p,_ := strconv.Atoi(arr[1])
	peerAddr := pb.PeerAddress{
		Ip:arr[0],
		Port:int32(p),
	}
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		errors.New("Cannot establish a connection!")
		log.Error("err:",err)
		return nil,err
	}
	peer.Connection = conn
	peer.Client = pb.NewChatClient(peer.Connection)
	peer.Addr = peerAddr
	//fromAddr := Server.GetChatServerAddr()
	helloMessage := pb.Message{
		MessageType:pb.Message_HELLO,
		Payload:[]byte("HELLO"),
		//From:&fromAddr,
	}
	retMessage,err2 := peer.Client.Chat(context.Background(),&helloMessage)
	if err2 != nil{
		errors.New("cannot establish a connection!")
		log.Error("cannot establish a connection,errinfo: ",err2)
		return nil,err2
	}else{
		if retMessage.MessageType == pb.Message_RESPONSE {
			// get the peer id
			origData, err := transport.TripleDesDecrypt(retMessage.Payload, DESKEY)
			//log.Notice(string(origData))
			if err != nil{
				log.Error("cannot decrypt the nodeidinfo!")
				errors.New("Decrypt ERROR")
			}
			peer.Idetity = string(origData)
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
	r,err := this.Client.Chat(context.Background(),msg)
	if err != nil{
		log.Error("err:",err)
	}
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