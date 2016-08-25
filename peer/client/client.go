package client

import (
	"google.golang.org/grpc"
	pb "hyperchain-alpha/peer/peermessage"
	"errors"
	"golang.org/x/net/context"
	"log"
	"strings"
	"strconv"
)

type ChatClient struct {
	Addr pb.PeerAddress
	Connection *grpc.ClientConn
	Client pb.ChatClient
	Idetity string
}

func NewChatClient(address string)(*ChatClient,error){
	var chatClient ChatClient
	arr := strings.Split(address,":")
	p,_ := strconv.Atoi(arr[1])
	peerAddr := pb.PeerAddress{
		Ip:arr[0],
		Port:int32(p),
	}
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		errors.New("cannot establish a connection!")
		log.Println("err:",err)
		return nil,err
	}
	chatClient.Connection = conn
	chatClient.Client = pb.NewChatClient(chatClient.Connection)
	chatClient.Addr = peerAddr
	//fromAddr := Server.GetChatServerAddr()
	helloMessage := pb.Message{
		MessageType:pb.Message_HELLO,
		Payload:[]byte("HELLO"),
		//From:&fromAddr,
	}
	retMessage,err2 := chatClient.Client.Chat(context.Background(),&helloMessage)
	if err2 != nil{
		errors.New("cannot establish a connection!无法建立通讯")
		log.Println("无法建立通讯 err:",err2)
		return nil,err
	}else{
		if retMessage.MessageType == pb.Message_RESPONSE {
			return &chatClient,nil
		}
	}
	return nil,errors.New("无法建立连接")
}

func (cc *ChatClient)Chat(msg *pb.Message) (*pb.Message, error){
	this := cc
	r,err := this.Client.Chat(context.Background(),msg)
	if err != nil{
		log.Println("err:",err)
	}
	return r,err
}

func (cc *ChatClient)Close()(bool,error){
	this := cc
	err := this.Connection.Close()
	if err != nil{
		log.Println("err:",err)
		return false,err
	}else{
		return true,nil
	}
}