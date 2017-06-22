package network

import (
	"google.golang.org/grpc"
	pb "hyperchain/p2p/message"
	"fmt"
	"golang.org/x/net/context"
	"time"
	"hyperchain/p2p/hts"
	"github.com/pkg/errors"
)

type Client struct {
	addr string
	conn *grpc.ClientConn
	client ChatClient
	MsgChan chan *pb.Message
	hts hts.HTS
}

func NewClient(addr string) *Client{
	return &Client{
		MsgChan: make(chan *pb.Message,100000),
		addr: addr,
	}
}

func(c *Client)Connect(client ChatClient) error{
	if client != nil{
		c.client = client
		return nil
	}
	//TODO hea
	conn, err := grpc.Dial(c.addr,grpc.WithInsecure())
	if err != nil {
		logger.Errorf("cannot create the connection to addr: %s \n",c.addr)
		fmt.Printf("err: %v",err)
		if conn != nil{
			conn.Close()
		}
		return err
	}
	c.conn = conn
	c.client = NewChatClient(conn)
	return nil
}

func(c *Client)Close() error{
	if c.conn != nil{
		return c.conn.Close()
	}else{
		return nil
	}
}


func(c *Client)Chat() (error){
	if c.client == nil{
		fmt.Printf("the client is nil %v \n",c.client)
		return nil
	}
	stream,err := c.client.Chat(context.Background())
	if err != nil{
		fmt.Printf("cannot create stream! %v \n" ,err)
		return err
	}
	for msg := range c.MsgChan{
		fmt.Println("actual send", string(msg.Payload), time.Now().UnixNano())
		err := stream.Send(msg)
		if err != nil{
			fmt.Errorf(err.Error())
		}
	}
	return nil
}
// Greeting doube arrow greeting message transfer
func(c *Client)Greeting(in *pb.Message) (*pb.Message, error){
	if c.client == nil{
		fmt.Printf("the client is nil %v \n",c.client)
		return nil,errors.New(fmt.Sprintf("the client is nil %v \n",c.client))
	}
	return c.client.Greeting(context.Background(),in)
}
// Wisper Transfer the the node health infomation
func(c *Client)Wisper(in *pb.Message) (*pb.Message, error){
	if c.client == nil{
		fmt.Printf("the client is nil %v \n",c.client)
		return nil,errors.New(fmt.Sprintf("the client is nil %v \n",c.client))
	}
	return c.client.Whisper(context.Background(),in)
}
