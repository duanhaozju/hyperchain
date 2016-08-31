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
	"log"
	"strings"
	"strconv"
)

type Peer struct {
	Addr pb.PeerAddress
	Connection *grpc.ClientConn
	Client pb.ChatClient
	Idetity string
}

func NewPeer(address string)(*Peer,error){
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
		log.Println("err:",err)
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
		errors.New("cannot establish a connection!无法建立通讯")
		log.Println("无法建立通讯 err:",err2)
		return nil,err2
	}else{
		if retMessage.MessageType == pb.Message_RESPONSE {
			return &peer,nil
		}
	}
	return nil,errors.New("无法建立连接")
}

func (this *Peer)Chat(msg *pb.Message) (*pb.Message, error){
	r,err := this.Client.Chat(context.Background(),msg)
	if err != nil{
		log.Println("err:",err)
	}
	return r,err
}

func (this *Peer)Close()(bool,error){
	err := this.Connection.Close()
	if err != nil{
		log.Println("err:",err)
		return false,err
	}else{
		return true,nil
	}
}